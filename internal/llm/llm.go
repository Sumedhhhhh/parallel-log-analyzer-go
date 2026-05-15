package llm

import "fmt"

type Record struct {
	Timestamp        string
	RequestID        string
	UserID           string
	Service          string
	Endpoint         string
	Model            string
	StatusCode       int
	LatencyMs        int
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	EstimatedCostUSD float64
	ErrorType        string
}

type ModelPrice struct {
	PromptPer1MTokens     float64
	CompletionPer1MTokens float64
}

var ModelPricing = map[string]ModelPrice{
	"gpt-4o-mini":  {PromptPer1MTokens: 0.15, CompletionPer1MTokens: 0.60},
	"gpt-4.1-mini": {PromptPer1MTokens: 0.40, CompletionPer1MTokens: 1.60},
	"claude-haiku": {PromptPer1MTokens: 0.25, CompletionPer1MTokens: 1.25},
}

func EstimateCostUSD(model string, promptTokens int, completionTokens int) float64 {
	price, ok := ModelPricing[model]
	if !ok {
		fmt.Printf("missing pricing for model: %q\n", model)
		return 0
	}

	promptCost := float64(promptTokens) / 1_000_000 * price.PromptPer1MTokens
	completionCost := float64(completionTokens) / 1_000_0000 * price.CompletionPer1MTokens

	return promptCost + completionCost
}
