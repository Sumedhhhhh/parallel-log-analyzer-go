package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/Sumedhhhhh/parallel-log-analyzer-go/internal/llm"
)

type LLMRequestSummary struct {
	RequestID   string
	Service     string
	Endpoint    string
	Model       string
	TotalTokens int
	CostUSD     float64
	LatencyMs   int
	StatusCode  int
	ErrorType   string
}

type LLMMetrics struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	ErrorRate          float64

	TotalPromptTokens     int64
	TotalCompletionTokens int64
	TotalTokens           int64
	AverageTokensPerReq   float64

	EstimatedTotalCostUSD float64
	AverageCostPerRequest float64

	AverageLatencyMs float64
	P95LatencyMs     int
	P99LatencyMs     int
	MaxLatencyMs     int

	RequestsByService map[string]int
	TokensByService   map[string]int64
	CostByService     map[string]float64

	RequestsByModel map[string]int
	TokensByModel   map[string]int64
	CostByModel     map[string]float64

	ErrorsByEndpoint  map[string]int
	ErrorsByErrorType map[string]int

	TopCostRequests    []LLMRequestSummary
	TopLatencyRequests []LLMRequestSummary
}

type llmPartialMetrics struct {
	totalRequests      int
	successfulRequests int
	failedRequests     int

	totalPromptTokens     int64
	totalCompletionTokens int64
	totalTokens           int64

	totalCostUSD float64

	latencySumMs int64
	maxLatencyMs int
	latencies    []int

	requestsByService map[string]int
	tokensByService   map[string]int64
	costByService     map[string]float64

	requestsByModel map[string]int
	tokensByModel   map[string]int64
	costByModel     map[string]float64

	errorsByEndpoint  map[string]int
	errorsByErrorType map[string]int

	topCostRequests    []LLMRequestSummary
	topLatencyRequests []LLMRequestSummary
}

func AnalyzeLLMSingleThreaded(filePath string, topN int) (LLMMetrics, error) {
	if topN <= 0 {
		topN = 10
	}

	file, err := os.Open(filePath)
	if err != nil {
		return LLMMetrics{}, err
	}
	defer file.Close()

	partial := newLLMPartialMetrics()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		record, err := parseLLMLogLine(scanner.Text())
		if err != nil {
			return LLMMetrics{}, err
		}

		updateLLMPartialMetrics(&partial, record, topN)
	}

	if err := scanner.Err(); err != nil {
		return LLMMetrics{}, err
	}

	return finalizeLLMMetrics(partial), nil
}

func AnalyzeLLMConcurrent(filePath string, workerCount int, topN int) (LLMMetrics, error) {
	if workerCount <= 0 {
		return LLMMetrics{}, fmt.Errorf("worker count must be greater than 0")
	}

	if topN <= 0 {
		topN = 10
	}

	file, err := os.Open(filePath)
	if err != nil {
		return LLMMetrics{}, err
	}
	defer file.Close()

	jobs := make(chan string, 1024)
	results := make(chan llmPartialMetrics, workerCount)
	errors := make(chan error, workerCount)

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			partial := newLLMPartialMetrics()

			for line := range jobs {
				record, err := parseLLMLogLine(line)
				if err != nil {
					errors <- err
					continue
				}

				updateLLMPartialMetrics(&partial, record, topN)
			}

			results <- partial
		}()
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		jobs <- scanner.Text()
	}

	close(jobs)

	if err := scanner.Err(); err != nil {
		return LLMMetrics{}, err
	}

	wg.Wait()
	close(results)
	close(errors)

	for err := range errors {
		if err != nil {
			return LLMMetrics{}, err
		}
	}

	finalPartial := newLLMPartialMetrics()

	for partial := range results {
		mergeLLMPartialMetrics(&finalPartial, partial, topN)
	}

	return finalizeLLMMetrics(finalPartial), nil
}

func newLLMPartialMetrics() llmPartialMetrics {
	return llmPartialMetrics{
		latencies: make([]int, 0),

		requestsByService: make(map[string]int),
		tokensByService:   make(map[string]int64),
		costByService:     make(map[string]float64),

		requestsByModel: make(map[string]int),
		tokensByModel:   make(map[string]int64),
		costByModel:     make(map[string]float64),

		errorsByEndpoint:  make(map[string]int),
		errorsByErrorType: make(map[string]int),

		topCostRequests:    make([]LLMRequestSummary, 0),
		topLatencyRequests: make([]LLMRequestSummary, 0),
	}
}

func updateLLMPartialMetrics(metrics *llmPartialMetrics, record llm.Record, topN int) {
	metrics.totalRequests++

	if record.StatusCode >= 200 && record.StatusCode < 300 {
		metrics.successfulRequests++
	} else {
		metrics.failedRequests++
		metrics.errorsByEndpoint[record.Endpoint]++

		if record.ErrorType != "" && record.ErrorType != "none" {
			metrics.errorsByErrorType[record.ErrorType]++
		}
	}

	metrics.totalPromptTokens += int64(record.PromptTokens)
	metrics.totalCompletionTokens += int64(record.CompletionTokens)
	metrics.totalTokens += int64(record.TotalTokens)
	metrics.totalCostUSD += record.EstimatedCostUSD

	metrics.latencySumMs += int64(record.LatencyMs)
	metrics.latencies = append(metrics.latencies, record.LatencyMs)

	if record.LatencyMs > metrics.maxLatencyMs {
		metrics.maxLatencyMs = record.LatencyMs
	}

	metrics.requestsByService[record.Service]++
	metrics.tokensByService[record.Service] += int64(record.TotalTokens)
	metrics.costByService[record.Service] += record.EstimatedCostUSD

	metrics.requestsByModel[record.Model]++
	metrics.tokensByModel[record.Model] += int64(record.TotalTokens)
	metrics.costByModel[record.Model] += record.EstimatedCostUSD

	summary := LLMRequestSummary{
		RequestID:   record.RequestID,
		Service:     record.Service,
		Endpoint:    record.Endpoint,
		Model:       record.Model,
		TotalTokens: record.TotalTokens,
		CostUSD:     record.EstimatedCostUSD,
		LatencyMs:   record.LatencyMs,
		StatusCode:  record.StatusCode,
		ErrorType:   record.ErrorType,
	}

	metrics.topCostRequests = addTopCostRequest(metrics.topCostRequests, summary, topN)
	metrics.topLatencyRequests = addTopLatencyRequest(metrics.topLatencyRequests, summary, topN)
}

func mergeLLMPartialMetrics(target *llmPartialMetrics, source llmPartialMetrics, topN int) {
	target.totalRequests += source.totalRequests
	target.successfulRequests += source.successfulRequests
	target.failedRequests += source.failedRequests

	target.totalPromptTokens += source.totalPromptTokens
	target.totalCompletionTokens += source.totalCompletionTokens
	target.totalTokens += source.totalTokens
	target.totalCostUSD += source.totalCostUSD

	target.latencySumMs += source.latencySumMs
	target.latencies = append(target.latencies, source.latencies...)

	if source.maxLatencyMs > target.maxLatencyMs {
		target.maxLatencyMs = source.maxLatencyMs
	}

	for service, count := range source.requestsByService {
		target.requestsByService[service] += count
	}

	for service, tokens := range source.tokensByService {
		target.tokensByService[service] += tokens
	}

	for service, cost := range source.costByService {
		target.costByService[service] += cost
	}

	for model, count := range source.requestsByModel {
		target.requestsByModel[model] += count
	}

	for model, tokens := range source.tokensByModel {
		target.tokensByModel[model] += tokens
	}

	for model, cost := range source.costByModel {
		target.costByModel[model] += cost
	}

	for endpoint, count := range source.errorsByEndpoint {
		target.errorsByEndpoint[endpoint] += count
	}

	for errorType, count := range source.errorsByErrorType {
		target.errorsByErrorType[errorType] += count
	}

	for _, request := range source.topCostRequests {
		target.topCostRequests = addTopCostRequest(target.topCostRequests, request, topN)
	}

	for _, request := range source.topLatencyRequests {
		target.topLatencyRequests = addTopLatencyRequest(target.topLatencyRequests, request, topN)
	}
}

func finalizeLLMMetrics(partial llmPartialMetrics) LLMMetrics {
	metrics := LLMMetrics{
		TotalRequests:      partial.totalRequests,
		SuccessfulRequests: partial.successfulRequests,
		FailedRequests:     partial.failedRequests,

		TotalPromptTokens:     partial.totalPromptTokens,
		TotalCompletionTokens: partial.totalCompletionTokens,
		TotalTokens:           partial.totalTokens,

		EstimatedTotalCostUSD: partial.totalCostUSD,

		MaxLatencyMs: partial.maxLatencyMs,

		RequestsByService: partial.requestsByService,
		TokensByService:   partial.tokensByService,
		CostByService:     partial.costByService,

		RequestsByModel: partial.requestsByModel,
		TokensByModel:   partial.tokensByModel,
		CostByModel:     partial.costByModel,

		ErrorsByEndpoint:  partial.errorsByEndpoint,
		ErrorsByErrorType: partial.errorsByErrorType,

		TopCostRequests:    partial.topCostRequests,
		TopLatencyRequests: partial.topLatencyRequests,
	}

	if partial.totalRequests > 0 {
		metrics.ErrorRate = float64(partial.failedRequests) / float64(partial.totalRequests) * 100
		metrics.AverageTokensPerReq = float64(partial.totalTokens) / float64(partial.totalRequests)
		metrics.AverageCostPerRequest = partial.totalCostUSD / float64(partial.totalRequests)
		metrics.AverageLatencyMs = float64(partial.latencySumMs) / float64(partial.totalRequests)
		metrics.P95LatencyMs = percentile(partial.latencies, 0.95)
		metrics.P99LatencyMs = percentile(partial.latencies, 0.99)
	}

	return metrics
}

func addTopCostRequest(requests []LLMRequestSummary, candidate LLMRequestSummary, topN int) []LLMRequestSummary {
	requests = append(requests, candidate)

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].CostUSD > requests[j].CostUSD
	})

	if len(requests) > topN {
		requests = requests[:topN]
	}

	return requests
}

func addTopLatencyRequest(requests []LLMRequestSummary, candidate LLMRequestSummary, topN int) []LLMRequestSummary {
	requests = append(requests, candidate)

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].LatencyMs > requests[j].LatencyMs
	})

	if len(requests) > topN {
		requests = requests[:topN]
	}

	return requests
}
