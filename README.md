# Parallel Log Analyzer in Go

A high-performance command-line analytics engine built in Go to generate synthetic logs, analyze service metrics, and benchmark single-threaded vs concurrent processing.

Supports two log formats:
- **Service logs** — HTTP-style request logs with status codes and latency
- **LLM logs** — LLM API request logs with token usage, cost, and model-level metrics

This project was built to explore Go concurrency patterns such as goroutines, channels, worker pools, WaitGroups, local aggregation, and performance benchmarking.

---

## Features

- Generate large synthetic service or LLM API log files
- Analyze logs for backend/SRE-style metrics or LLM observability metrics
- Break down LLM metrics by service and model
- Surface top-N most expensive and slowest LLM requests
- Compare single-threaded and concurrent processing
- Benchmark throughput across different worker counts

---

## Log Formats

### Service Log Format

Each generated line contains:

```text
timestamp service endpoint statusCode latencyMs userId region requestId
```

Example:

```text
2026-05-08T20:15:30Z auth-service /login 200 42 user-1023 us-east req-948292
2026-05-08T20:15:31Z payment-service /checkout 500 812 user-456 eu-west req-112931
```

### LLM Log Format

Each generated line is pipe-delimited with 13 fields:

```text
timestamp|requestId|userId|service|endpoint|model|statusCode|latencyMs|promptTokens|completionTokens|totalTokens|estimatedCostUSD|errorType
```

Example:

```text
2026-05-14T10:00:00Z|req-48291031|user-3821|rag-api|/query|gpt-4o-mini|200|812|1204|387|1591|0.00041265|none
2026-05-14T10:00:01Z|req-72910482|user-9104|agent-service|/agent/run|claude-haiku|500|3204|8821|1022|9843|0.00345100|model_error
```

#### LLM Services and Endpoints

| Service | Endpoint |
|---|---|
| rag-api | /query |
| agent-service | /agent/run |
| summarizer | /summarize |
| chat-service | /chat |
| embedding-service | /embed |

#### Supported Models and Pricing

| Model | Prompt (per 1M tokens) | Completion (per 1M tokens) |
|---|---|---|
| gpt-4o-mini | $0.15 | $0.60 |
| gpt-4.1-mini | $0.40 | $1.60 |
| claude-haiku | $0.25 | $1.25 |

---

## Metrics Computed

### Service Log Metrics

- Total requests
- Error count and error rate
- Average, P95, P99, and max latency
- Request count by service
- Error count by endpoint

### LLM Log Metrics

- Total requests, successful requests, failed requests, and error rate
- Total and average tokens per request (prompt, completion, total)
- Estimated total and average cost per request (USD)
- Average, P95, P99, and max latency
- Request count, token usage, and cost broken down by service
- Request count, token usage, and cost broken down by model
- Error count by endpoint and by error type (`timeout`, `rate_limit`, `model_error`, `context_length_exceeded`, `upstream_error`)
- Top-N most expensive requests
- Top-N highest latency requests

---

## Architecture

```
         +----------------------+
         |   Log Generator      |
         | (service or LLM)     |
         +----------+-----------+
                    |
                    v
            Generated Log File
                    |
      +-------------+-------------+
      |                           |
      v                           v
Single-threaded Analyzer   Concurrent Analyzer
                                  |
                                  v
                         Worker Pool Processing
                                  |
                                  v
                   Local Partial Metrics per Worker
                                  |
                                  v
                         Merge & Finalize Metrics
```

---

## Concurrency Design

The concurrent analyzer uses a worker pool pattern:

```
file reader
    |
    v
jobs channel (buffered, 1024)
    |
    v
worker goroutines
    |
    v
partial metrics per worker (local aggregation)
    |
    v
merge final metrics
```

Each worker maintains its own local metrics struct. At the end, all partial results are merged into a single final result. This avoids shared mutable state and eliminates lock contention.

---

## Project Structure

```
parallel-log-analyzer-go/
├── cmd/
│   └── parallel-log-analyzer/
│       └── main.go
├── internal/
│   ├── analyzer/
│   │   ├── analyzer.go
│   │   ├── llm_analyzer.go
│   │   └── llm_parser.go
│   ├── benchmark/
│   │   └── benchmark.go
│   ├── generator/
│   │   ├── generator.go
│   │   └── llm_generator.go
│   └── llm/
│       └── llm.go
├── go.mod
├── .gitignore
└── README.md
```

---

## Setup

Clone the repository:

```bash
git clone https://github.com/Sumedhhhhh/parallel-log-analyzer-go.git
cd parallel-log-analyzer-go
```

Build the CLI:

```bash
go build -o parallel-log-analyzer ./cmd/parallel-log-analyzer
```

---

## Generate Logs

Generate 10,000 service log lines:

```bash
./parallel-log-analyzer --mode generate --out data/logs.txt --lines 10000
```

Generate 10,000 LLM log records:

```bash
./parallel-log-analyzer --mode generate --logType llm --out data/llm_logs.txt --lines 10000
```

Generate 1 million lines of either type:

```bash
./parallel-log-analyzer --mode generate --out data/logs_1m.txt --lines 1000000
./parallel-log-analyzer --mode generate --logType llm --out data/llm_logs_1m.txt --lines 1000000
```

---

## Analyze Logs

### Service Logs

Run the single-threaded analyzer:

```bash
./parallel-log-analyzer --mode analyze --file data/logs_1m.txt --analyzerMode single
```

Run the concurrent analyzer with 4 workers:

```bash
./parallel-log-analyzer --mode analyze --file data/logs_1m.txt --analyzerMode concurrent --workers 4
```

Example output:

```
========== Log Analysis Results ==========
Total Requests: 1000000
Error Count: 214269
Error Rate: 21.43%
Average Latency: 209.86 ms
P95 Latency: 308 ms
P99 Latency: 1765 ms
Max Latency: 2308 ms

Requests By Service:
  auth-service: 166392
  payment-service: 166871
  ...

Analyzer Mode: concurrent
Workers: 4
Duration: 324.92ms
```

### LLM Logs

Run the single-threaded analyzer:

```bash
./parallel-log-analyzer --mode analyze --logType llm --file data/llm_logs_1m.txt --analyzerMode single
```

Run the concurrent analyzer with 4 workers:

```bash
./parallel-log-analyzer --mode analyze --logType llm --file data/llm_logs_1m.txt --analyzerMode concurrent --workers 4
```

Example output:

```
========== LLM Log Analysis Results ==========
Total Requests:        1000000
Successful Requests:   960000
Failed Requests:       40000
Error Rate:            4.00%

Total Tokens:          4812930201
Avg Tokens/Request:    4812.93
Estimated Total Cost:  $2048.31
Avg Cost/Request:      $0.000002048

Avg Latency:           1204.83 ms
P95 Latency:           5812 ms
P99 Latency:           8921 ms
Max Latency:           11203 ms

Cost By Model:
  gpt-4o-mini:         $821.20
  gpt-4.1-mini:        $903.44
  claude-haiku:        $323.67

Errors By Type:
  rate_limit:          14821
  timeout:             9204
  model_error:         8312

Analyzer Mode: concurrent
Workers: 4
Duration: 412.31ms
```

---

## Benchmark

Run benchmark mode for service logs:

```bash
./parallel-log-analyzer --mode benchmark --file data/logs_1m.txt
```

Run benchmark mode for LLM logs:

```bash
./parallel-log-analyzer --mode benchmark --logType llm --file data/llm_logs_1m.txt
```

Example output:

```
========== Benchmark Results ==========
Mode          Workers   Duration        Throughput
single        1         313.43ms        3190417 lines/sec
concurrent    1         355.10ms        2816104 lines/sec
concurrent    2         340.81ms        2934165 lines/sec
concurrent    4         324.99ms        3077012 lines/sec
concurrent    8         359.22ms        2783819 lines/sec
concurrent    16        390.88ms        2558270 lines/sec
```

---

## Performance Notes

LLM records are more expensive to parse than service logs — 13 pipe-delimited fields with multiple numeric conversions (token counts, latency, float cost). This raises the per-line CPU cost, which makes concurrent processing more beneficial for LLM workloads than for simple HTTP logs.

For both formats, the concurrent version carries coordination overhead (channel communication, goroutine scheduling, result merging). The sweet spot is typically 2–4 workers. Beyond that, scheduling overhead starts to eat into the gains.

This is an important systems lesson: concurrency improves throughput only when the work per unit is large enough to offset coordination overhead.

---

## What I Learned

- Go project structure and package organization
- CLI development in Go
- Structs, slices, maps, and error handling
- File I/O with buffered readers and writers
- Goroutines and channels
- WaitGroups and worker pools
- Local aggregation and merge-based concurrency
- Token-based cost modeling for LLM APIs
- Measuring execution time and throughput
- Interpreting performance bottlenecks realistically

---

## Future Improvements

- Add JSON log format support
- Add CSV export for benchmark and analysis results
- Add time-series bucketing (requests and cost over time)
- Add streaming mode for live LLM API traffic
- Add unit tests for parser and metrics aggregation
- Add pprof profiling
- Add Kafka streaming mode
- Add Docker support
- Add approximate percentile calculation using t-digest or DDSketch

---

## Tech Stack

- Go
- Goroutines
- Channels
- Worker pools
- Buffered file I/O
- CLI flags
- Performance benchmarking
- LLM token/cost modeling
