package benchmark

import (
	"fmt"
	"time"

	"github.com/Sumedhhhhh/parallel-log-analyzer-go/internal/analyzer"
)

type Result struct {
	Mode               string
	Workers            int
	Duration           time.Duration
	ThroughputLinesSec float64
	TotalRequests      int
}

func Run(filePath string) ([]Result, error) {
	results := make([]Result, 0)

	singleResult, err := runSingle(filePath)
	if err != nil {
		return nil, err
	}

	results = append(results, singleResult)

	workerCounts := []int{1, 2, 4, 8, 16}

	for _, workers := range workerCounts {
		result, err := runConcurrent(filePath, workers)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil

}

func runSingle(filePath string) (Result, error) {
	start := time.Now()

	metrics, err := analyzer.AnalyzeSingleThreaded(filePath)
	if err != nil {
		return Result{}, err
	}

	duration := time.Since(start)

	return buildResult("single", 1, duration, metrics.TotalRequests), nil
}

func runConcurrent(filePath string, workers int) (Result, error) {
	start := time.Now()

	metrics, err := analyzer.AnalyzeConcurrent(filePath, workers)
	if err != nil {
		return Result{}, err
	}

	duration := time.Since(start)

	return buildResult("concurrent", workers, duration, metrics.TotalRequests), nil
}

func buildResult(mode string, workers int, duration time.Duration, totalRequests int) Result {
	seconds := duration.Seconds()

	throughput := 0.0
	if seconds > 0 {
		throughput = float64(totalRequests) / seconds
	}

	return Result{
		Mode:               mode,
		Workers:            workers,
		Duration:           duration,
		ThroughputLinesSec: throughput,
		TotalRequests:      totalRequests,
	}
}

func PrintResults(results []Result) {
	fmt.Println("========== Benchmark Results ==========")
	fmt.Printf("%-13s %-9s %-15s %-20s\n", "Mode", "Workers", "Duration", "Throughput")

	for _, result := range results {
		fmt.Printf(
			"%-13s %-9d %-15s %.0f lines/sec\n",
			result.Mode,
			result.Workers,
			result.Duration,
			result.ThroughputLinesSec,
		)
	}
}
