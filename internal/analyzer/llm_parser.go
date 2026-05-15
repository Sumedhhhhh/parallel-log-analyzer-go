package analyzer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Sumedhhhhh/parallel-log-analyzer-go/internal/llm"
)

func parseLLMLogLine(line string) (llm.Record, error) {
	parts := strings.Split(line, "|")

	if len(parts) != 13 {
		return llm.Record{}, fmt.Errorf("invalid llm log line format : expected 13 fields, got %d", len(parts))
	}

	statusCode, err := strconv.Atoi(parts[6])
	if err != nil {
		return llm.Record{}, fmt.Errorf("invalid status code: %s", parts[6])
	}

	latencyMs, err := strconv.Atoi(parts[7])
	if err != nil {
		return llm.Record{}, fmt.Errorf("invalid latency_ms: %s", parts[7])
	}

	promptTokens, err := strconv.Atoi(parts[8])
	if err != nil {
		return llm.Record{}, fmt.Errorf("invalid prompt_tokens: %s", parts[8])
	}

	completionTokens, err := strconv.Atoi(parts[9])
	if err != nil {
		return llm.Record{}, fmt.Errorf("invalid completion_tokens: %s", parts[9])
	}

	totalTokens, err := strconv.Atoi(parts[10])
	if err != nil {
		return llm.Record{}, fmt.Errorf("invalid total_tokens: %s", parts[10])
	}

	estimatedCostUSD, err := strconv.ParseFloat(parts[11], 64)
	if err != nil {
		return llm.Record{}, fmt.Errorf("invalid estimated_cost_usd: %s", parts[11])
	}

	return llm.Record{
		Timestamp:        parts[0],
		RequestID:        parts[1],
		UserID:           parts[2],
		Service:          parts[3],
		Endpoint:         parts[4],
		Model:            parts[5],
		StatusCode:       statusCode,
		LatencyMs:        latencyMs,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		EstimatedCostUSD: estimatedCostUSD,
		ErrorType:        parts[12],
	}, nil
}
