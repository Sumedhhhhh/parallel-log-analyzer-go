package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Metrics struct {
	TotalRequests    int
	ErrorCount       int
	ErrorRate        float64
	AverageLatencyMs float64
	P95LatencyMs     int
	P99LatencyMs     int
	RequestByService map[string]int
	ErrorsByEndpoint map[string]int
	MaxLatencyMs     int
}

type parsedLogLine struct {
	Service    string
	Endpoint   string
	StatusCode int
	LatencyMs  int
}

type partialMetrics struct {
	totalRequests     int
	errorCount        int
	latencySum        int
	maxLatencyMs      int
	latencies         []int
	requestsByService map[string]int
	errorsByEndpoint  map[string]int
}

func AnalyzeSingleThreaded(filePath string) (Metrics, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Metrics{}, err
	}
	defer file.Close()

	partial := newPartialMetrics()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		logLine, err := parseLogLine(line)
		if err != nil {
			return Metrics{}, err
		}

		updatePartialMetrics(&partial, logLine)
	}

	if err := scanner.Err(); err != nil {
		return Metrics{}, err
	}

	return finalizeMetrics(partial), nil
}

func AnalyzeConcurrent(filePath string, workerCount int) (Metrics, error) {
	if workerCount <= 0 {
		return Metrics{}, fmt.Errorf("worker count must be greater than 0")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return Metrics{}, err
	}
	defer file.Close()
	// creates a channel to send data to go-routines
	// 1024 is buffer capacity. Sender can queue upto 1024 lines before blocking without this sender will block until a worker receieves
	jobs := make(chan string, 1024)
	results := make(chan partialMetrics, workerCount)
	errorsChan := make(chan error, workerCount)

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			// no mutex required as each worker has its own partial. No other worker touches it
			// A mutex will be required when multiple go routines share and mutate same memory
			partial := newPartialMetrics()

			for line := range jobs {
				logLine, err := parseLogLine(line)
				if err != nil {
					errorsChan <- err
					continue
				}

				updatePartialMetrics(&partial, logLine)
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
		return Metrics{}, err
	}

	wg.Wait()
	close(results)
	close(errorsChan)

	for err := range errorsChan {
		if err != nil {
			return Metrics{}, err
		}
	}

	finalPartial := newPartialMetrics()

	for partial := range results {
		mergePartialMetrics(&finalPartial, partial)
	}

	return finalizeMetrics(finalPartial), nil
}

func newPartialMetrics() partialMetrics {
	return partialMetrics{
		latencies:         make([]int, 0),
		requestsByService: make(map[string]int),
		errorsByEndpoint:  make(map[string]int),
	}
}

func updatePartialMetrics(metrics *partialMetrics, logLine parsedLogLine) {
	metrics.totalRequests++
	metrics.requestsByService[logLine.Service]++
	metrics.latencySum += logLine.LatencyMs
	metrics.latencies = append(metrics.latencies, logLine.LatencyMs)

	if logLine.LatencyMs > metrics.maxLatencyMs {
		metrics.maxLatencyMs = logLine.LatencyMs
	}

	if logLine.StatusCode >= 500 {
		metrics.errorCount++
		metrics.errorsByEndpoint[logLine.Endpoint]++
	}
}

func mergePartialMetrics(target *partialMetrics, source partialMetrics) {
	target.totalRequests += source.totalRequests
	target.errorCount += source.errorCount
	target.latencySum += source.latencySum

	if source.maxLatencyMs > target.maxLatencyMs {
		target.maxLatencyMs = source.maxLatencyMs
	}

	target.latencies = append(target.latencies, source.latencies...)

	for service, count := range source.requestsByService {
		target.requestsByService[service] += count
	}

	for endpoint, count := range source.errorsByEndpoint {
		target.errorsByEndpoint[endpoint] += count
	}

}

func finalizeMetrics(partial partialMetrics) Metrics {
	metrics := Metrics{
		TotalRequests:    partial.totalRequests,
		ErrorCount:       partial.errorCount,
		MaxLatencyMs:     partial.maxLatencyMs,
		RequestByService: partial.requestsByService,
		ErrorsByEndpoint: partial.errorsByEndpoint,
	}

	if partial.totalRequests > 0 {
		metrics.ErrorRate = float64(partial.errorCount) / float64(partial.totalRequests) * 100
		metrics.AverageLatencyMs = float64(partial.latencySum) / float64(partial.totalRequests)
		metrics.P95LatencyMs = percentile(partial.latencies, 0.95)
		metrics.P99LatencyMs = percentile(partial.latencies, 0.99)
	}

	return metrics
}

func parseLogLine(line string) (parsedLogLine, error) {
	parts := strings.Fields(line)

	if len(parts) != 8 {
		return parsedLogLine{}, fmt.Errorf("invalid log line format; %s", line)
	}

	statusCode, err := strconv.Atoi(parts[3])
	if err != nil {
		return parsedLogLine{}, fmt.Errorf("invalid status code: %s", parts[3])
	}

	latencyMs, err := strconv.Atoi(parts[4])
	if err != nil {
		return parsedLogLine{}, fmt.Errorf("invalid latency: %s", parts[4])
	}

	return parsedLogLine{
		Service:    parts[1],
		Endpoint:   parts[2],
		StatusCode: statusCode,
		LatencyMs:  latencyMs,
	}, nil

}

func percentile(values []int, p float64) int {
	if len(values) == 0 {
		return 0
	}

	sort.Ints(values)

	index := int(float64(len(values)) * p)

	if index >= len(values) {
		index = len(values) - 1
	}

	return values[index]
}
