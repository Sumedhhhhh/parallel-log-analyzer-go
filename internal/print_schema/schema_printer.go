package printschema

import (
	"fmt"
	"sort"

	"github.com/Sumedhhhhh/parallel-log-analyzer-go/internal/analyzer"
)

type intMetricItem struct {
	Name  string
	Value int
}

type int64MetricItem struct {
	Name  string
	Value int64
}

type floatMetricItem struct {
	Name  string
	Value float64
}

func sortedIntMapDesc(m map[string]int) []intMetricItem {
	items := make([]intMetricItem, 0, len(m))

	for name, value := range m {
		items = append(items, intMetricItem{
			Name:  name,
			Value: value,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Name < items[j].Name
		}
		return items[i].Value > items[j].Value
	})

	return items

}

func sortedInt64MapDesc(m map[string]int64) []int64MetricItem {
	items := make([]int64MetricItem, 0, len(m))

	for name, value := range m {
		items = append(items, int64MetricItem{
			Name:  name,
			Value: value,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Name < items[j].Name
		}
		return items[i].Value > items[j].Value
	})

	return items
}

func sortedFloatMapDesc(m map[string]float64) []floatMetricItem {
	items := make([]floatMetricItem, 0, len(m))

	for name, value := range m {
		items = append(items, floatMetricItem{
			Name:  name,
			Value: value,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Name < items[j].Name
		}
		return items[i].Value > items[j].Value
	})

	return items
}

func PrintSchema(metrics analyzer.Metrics) {
	fmt.Println("========== Log Analysis Results ==========")
	fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Error Count: %d\n", metrics.ErrorCount)
	fmt.Printf("Error Rate: %.2f%%\n", metrics.ErrorRate)
	fmt.Printf("Average Latency: %.2f ms\n", metrics.AverageLatencyMs)
	fmt.Printf("P95 Latency: %d ms\n", metrics.P95LatencyMs)
	fmt.Printf("P99 Latency: %d ms\n", metrics.P99LatencyMs)
	fmt.Printf("Max Latency: %d ms\n", metrics.MaxLatencyMs)
	fmt.Println()
	fmt.Println("Requests By Service:")
	for service, count := range metrics.RequestByService {
		fmt.Printf("  %s: %d\n", service, count)
	}

	fmt.Println()
	fmt.Println("Errors By Endpoint:")
	for endpoint, count := range metrics.ErrorsByEndpoint {
		fmt.Printf("  %s: %d\n", endpoint, count)
	}
}

func PrintLLMSchema(metrics analyzer.LLMMetrics, topN int) {
	fmt.Println("========== LLM Usage Analysis ==========")
	fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", metrics.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", metrics.FailedRequests)
	fmt.Printf("Error Rate: %.2f%%\n", metrics.ErrorRate)

	fmt.Println()
	fmt.Printf("Prompt Tokens: %d\n", metrics.TotalPromptTokens)
	fmt.Printf("Completion Tokens: %d\n", metrics.TotalCompletionTokens)
	fmt.Printf("Total Tokens: %d\n", metrics.TotalTokens)
	fmt.Printf("Average Tokens / Request: %.2f\n", metrics.AverageTokensPerReq)

	fmt.Println()
	fmt.Printf("Estimated Total Cost: $%.6f\n", metrics.EstimatedTotalCostUSD)
	fmt.Printf("Average Cost / Request: $%.8f\n", metrics.AverageCostPerRequest)

	fmt.Println()
	fmt.Printf("Average Latency: %.2f ms\n", metrics.AverageLatencyMs)
	fmt.Printf("P95 Latency: %d ms\n", metrics.P95LatencyMs)
	fmt.Printf("P99 Latency: %d ms\n", metrics.P99LatencyMs)
	fmt.Printf("Max Latency: %d ms\n", metrics.MaxLatencyMs)

	fmt.Println()
	fmt.Println("Requests By Service:")
	for _, item := range sortedIntMapDesc(metrics.RequestsByService) {
		fmt.Printf("  %s: %d\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Println("Tokens By Service:")
	for _, item := range sortedInt64MapDesc(metrics.TokensByService) {
		fmt.Printf("  %s: %d\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Println("Cost By Service:")
	for _, item := range sortedFloatMapDesc(metrics.CostByService) {
		fmt.Printf("  %s: $%.6f\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Println("Requests By Model:")
	for _, item := range sortedIntMapDesc(metrics.RequestsByModel) {
		fmt.Printf("  %s: %d\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Println("Tokens By Model:")
	for _, item := range sortedInt64MapDesc(metrics.TokensByModel) {
		fmt.Printf("  %s: %d\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Println("Cost By Model:")
	for _, item := range sortedFloatMapDesc(metrics.CostByModel) {
		fmt.Printf("  %s: $%.6f\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Println("Errors By Endpoint:")
	for _, item := range sortedIntMapDesc(metrics.ErrorsByEndpoint) {
		fmt.Printf("  %s: %d\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Println("Errors By Error Type:")
	for _, item := range sortedIntMapDesc(metrics.ErrorsByErrorType) {
		fmt.Printf("  %s: %d\n", item.Name, item.Value)
	}

	fmt.Println()
	fmt.Printf("Top %d Highest Cost Requests:\n", topN)
	for _, request := range metrics.TopCostRequests {
		fmt.Printf(
			"  %s | service=%s | model=%s | tokens=%d | cost=$%.8f | latency=%dms\n",
			request.RequestID,
			request.Service,
			request.Model,
			request.TotalTokens,
			request.CostUSD,
			request.LatencyMs,
		)
	}

	fmt.Println()
	fmt.Printf("Top %d Highest Latency Requests:\n", topN)
	for _, request := range metrics.TopLatencyRequests {
		fmt.Printf(
			"  %s | service=%s | model=%s | tokens=%d | cost=$%.8f | latency=%dms\n",
			request.RequestID,
			request.Service,
			request.Model,
			request.TotalTokens,
			request.CostUSD,
			request.LatencyMs,
		)
	}
}
