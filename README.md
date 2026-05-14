# Parallel Log Analyzer in Go

A high-performance command-line log analytics engine built in Go to generate synthetic production-style logs, analyze service metrics, and benchmark single-threaded vs concurrent processing.

This project was built to explore Go concurrency patterns such as goroutines, channels, worker pools, WaitGroups, local aggregation, and performance benchmarking.

---

## Features

- Generate large synthetic production-style log files
- Analyze logs for backend/SRE-style metrics
- Compare single-threaded and concurrent processing
- Benchmark throughput across different worker counts
- Compute latency percentiles and service-level error metrics

---

## Log Format

Each generated log line contains:

```text
timestamp service endpoint statusCode latencyMs userId region requestId
```

Example:

```text
2026-05-08T20:15:30Z auth-service /login 200 42 user-1023 us-east req-948292
2026-05-08T20:15:31Z payment-service /checkout 500 812 user-456 eu-west req-112931
```

---

## Metrics Computed

The analyzer computes:

- Total requests
- Error count
- Error rate
- Average latency
- Max latency
- P95 latency
- P99 latency
- Request count by service
- Error count by endpoint

---

## Architecture

```
         +----------------------+
         | Synthetic Generator  |
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
                          Local Aggregation
                                  |
                                  v
                            Merge Metrics
```

---

## Concurrency Design

The concurrent analyzer uses a worker pool pattern:

```
file reader
    |
    v
jobs channel
    |
    v
worker goroutines
    |
    v
partial metrics per worker
    |
    v
merge final metrics
```

Instead of using one shared global metrics object protected by a mutex, each worker maintains its own local metrics. At the end, the partial results are merged into final metrics.

This reduces lock contention and makes the design easier to reason about.

---

## Project Structure

```
parallel-log-analyzer-go/
├── cmd/
│   └── parallel-log-analyzer/
│       └── main.go
├── internal/
│   ├── analyzer/
│   │   └── analyzer.go
│   ├── benchmark/
│   │   └── benchmark.go
│   └── generator/
│       └── generator.go
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

Run hello mode:

```bash
./parallel-log-analyzer --mode hello --name Sumedh
```

Expected output:

```
Mode: hello
Hello, Sumedh. Welcome to parallel-log-analyzer-go!
```

---

## Generate Logs

Generate 10,000 log lines:

```bash
./parallel-log-analyzer --mode generate --out data/logs.txt --lines 10000
```

Generate 1 million log lines:

```bash
./parallel-log-analyzer --mode generate --out data/logs_1m.txt --lines 1000000
```

Example output:

```
Generated 1000000 log lines
Output file: data/logs_1m.txt
Duration: 920.21ms
```

---

## Analyze Logs

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
  search-service: 166423
  profile-service: 166519
  recommendation-service: 166772
  notification-service: 167023

Errors By Endpoint:
  /login: 26972
  /checkout: 26684
  /search: 26569
  /profile: 26957
  /recommendations: 26647
  /notify: 26898
  /health: 26542
  /logout: 27000

Analyzer Mode: concurrent
Workers: 4
Duration: 324.92ms
```

---

## Benchmark

Run benchmark mode:

```bash
./parallel-log-analyzer --mode benchmark --file data/logs_1m.txt
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

The benchmark results show that concurrency does not automatically guarantee better performance.

For this workload, each log line requires a small amount of CPU work:

- Split the line
- Parse status code
- Parse latency
- Update counters
- Append latency value

The concurrent version introduces additional overhead:

- Channel communication
- Goroutine scheduling
- Worker coordination
- Result merging

Because the work per line is relatively small, the single-threaded version may perform as well as or better than the concurrent version.

This is an important systems lesson: concurrency improves throughput only when the work per unit is large enough to offset coordination overhead.

---

## What I Learned

This project helped me understand:

- Go project structure
- CLI development in Go
- Structs, slices, maps, and error handling
- File I/O with buffered readers and writers
- Goroutines and channels
- WaitGroups and worker pools
- Local aggregation and merge-based concurrency
- Measuring execution time and throughput
- Interpreting performance bottlenecks realistically

---

## Future Improvements

Possible extensions:

- Add Kafka streaming mode
- Add JSON log format support
- Add CSV output for benchmark results
- Add sorted output for service and endpoint metrics
- Add unit tests for parser and metrics aggregation
- Add pprof profiling
- Add approximate percentile calculation using histograms
- Add Docker support
- Add live metrics mode for streaming logs

---

## Tech Stack

- Go
- Goroutines
- Channels
- Worker pools
- Buffered file I/O
- CLI flags
- Performance benchmarking
