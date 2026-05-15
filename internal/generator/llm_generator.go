package generator

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/Sumedhhhhh/parallel-log-analyzer-go/internal/llm"
)

var llmService = []string{
	"rag-api",
	"agent-service",
	"summarizer",
	"chat-service",
	"embedding-service",
}

var llmEndpoints = []string{
	"/query",
	"/agent/run",
	"/summarize",
	"/chat",
	"/embed",
}

var llmModels = []string{
	"gpt-4o-mini",
	"gpt-4.1-mini",
	"claude-haiku",
}

var llmErrorTypes = []string{
	"none",
	"timeout",
	"rate_limit",
	"model_error",
	"context_length_exceeded",
	"upstream_error",
}

func GenerateLLMLogs(outputPath string, recordCound int) error {
	err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	startTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < recordCound; i++ {
		record := randomLLMRecord(r, startTime, i)

		_, err := writer.WriteString(formatLLMRecord(record) + "\n")
		if err != nil {
			return err
		}
	}

	return nil

}

func randomLLMRecord(r *rand.Rand, startTime time.Time, index int) llm.Record {
	model := randomLLMString(r, llmModels)
	service := randomLLMString(r, llmService)
	endpoint := endPointForService(service)

	statusCode := randomLLMStatusCode(r)
	errorType := "none"

	if statusCode >= 500 {
		errorType = randomNonNoneErrorType(r)
	}

	promptTokens := randomPromptTokens(r, endpoint)
	completionTokens := randomCompletionTokens(r, endpoint)
	totalTokens := promptTokens + completionTokens
	cost := llm.EstimateCostUSD(model, promptTokens, completionTokens)

	return llm.Record{
		Timestamp:        startTime.Add(time.Duration(index) * time.Millisecond).Format(time.RFC3339),
		RequestID:        fmt.Sprintf("req-%d", r.Intn(100000000)),
		UserID:           fmt.Sprintf("user-%d", r.Intn(100000)),
		Service:          service,
		Endpoint:         endpoint,
		Model:            model,
		StatusCode:       statusCode,
		LatencyMs:        randomLLMLatency(r, endpoint),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		EstimatedCostUSD: cost,
		ErrorType:        errorType,
	}

}

func formatLLMRecord(record llm.Record) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%d|%d|%d|%d|%d|%.8f|%s",
		record.Timestamp,
		record.RequestID,
		record.UserID,
		record.Service,
		record.Endpoint,
		record.Model,
		record.StatusCode,
		record.LatencyMs,
		record.PromptTokens,
		record.CompletionTokens,
		record.TotalTokens,
		record.EstimatedCostUSD,
		record.ErrorType,
	)
}

func randomLLMString(r *rand.Rand, values []string) string {
	return values[r.Intn(len(values))]
}

func endPointForService(service string) string {
	switch service {
	case "rag-api":
		return "/query"
	case "agent-service":
		return "/agent/run"
	case "summarizer":
		return "/summarize"
	case "chat-service":
		return "/chat"
	case "embedding-service":
		return "/embed"
	default:
		return "/query"
	}
}

func randomLLMStatusCode(r *rand.Rand) int {
	value := r.Intn(100)

	if value < 96 {
		return 200
	}

	if value < 98 {
		return 429
	}

	return []int{500, 502, 503, 504}[r.Intn(4)]
}

func randomNonNoneErrorType(r *rand.Rand) string {
	return llmErrorTypes[1+r.Intn(len(llmErrorTypes)-1)]
}

func randomPromptTokens(r *rand.Rand, endpoint string) int {
	switch endpoint {
	case "/embed":
		return 200 + r.Intn(1200)
	case "/summarize":
		return 1500 + r.Intn(7000)
	case "/agent/run":
		return 1000 + r.Intn(10000)
	case "/query":
		return 500 + r.Intn(6000)
	default:
		return 300 + r.Intn(3000)
	}
}

func randomCompletionTokens(r *rand.Rand, endpoint string) int {
	switch endpoint {
	case "/embed":
		return 0
	case "/summarize":
		return 200 + r.Intn(1200)
	case "/agent/run":
		return 300 + r.Intn(3000)
	case "/query":
		return 100 + r.Intn(1500)
	default:
		return 100 + r.Intn(1000)
	}
}

func randomLLMLatency(r *rand.Rand, endpoint string) int {
	base := 0

	switch endpoint {
	case "/embed":
		base = 100 + r.Intn(600)
	case "/summarize":
		base = 800 + r.Intn(2500)
	case "/agent/run":
		base = 1200 + r.Intn(5000)
	case "/query":
		base = 600 + r.Intn(3000)
	default:
		base = 300 + r.Intn(1500)
	}

	if r.Intn(100) < 3 {
		return base + r.Intn(5000)
	}

	return base
}
