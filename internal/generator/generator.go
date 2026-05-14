package generator

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// A struct in Go is similar to a simple Java class or Python dataclass.
// Go does not require getters/setters for simple data containers.
type LogEntry struct {
	Timestamp  time.Time
	Service    string
	Endpoint   string
	StatusCode int
	LatencyMs  int
	UserID     string
	Region     string
	RequestID  string
}

var services = []string{
	"auth-service",
	"payment-service",
	"search-service",
	"profile-service",
	"recommendation-service",
	"notification-service",
}

var endpoints = []string{
	"/login",
	"/logout",
	"/checkout",
	"/search",
	"/profile",
	"/recommendations",
	"/notify",
	"/health",
}

var regions = []string{
	"us-east",
	"us-west",
	"eu-west",
	"ap-south",
	"ap-southeast",
}

var statusCodes = []int{
	200, 200, 200, 200, 200,
	201, 204,
	400, 401, 403, 404,
	500, 502, 503,
}

// This function:
// Creates the output directory.
// Creates the log file.
// Uses buffered writing.
// Generates lineCount fake log entries.
// Writes each one as a line.
// in Go, instead of throwing exceptions, we return error.
func GenerateLogs(outputPath string, lineCount int) error {
	err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	//defer means: Run this when the current function exits. It ensures cleanup happens even if we return early.
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	startTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < lineCount; i++ {
		entry := LogEntry{
			Timestamp:  startTime.Add(time.Duration(i) * time.Millisecond),
			Service:    randomString(r, services),
			Endpoint:   randomString(r, endpoints),
			StatusCode: randomStatusCode(r),
			LatencyMs:  randomLatency(r),
			UserID:     fmt.Sprintf("user-%d", r.Intn(100000)),
			Region:     randomString(r, regions),
			RequestID:  fmt.Sprintf("req-%d", r.Intn(10000000)),
		}

		line := formatLogEntry(entry)

		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func randomString(r *rand.Rand, values []string) string {
	return values[r.Intn(len(values))]
}

func randomStatusCode(r *rand.Rand) int {
	return statusCodes[r.Intn(len(statusCodes))]
}

func randomLatency(r *rand.Rand) int {
	base := r.Intn(300) + 10

	if r.Intn(100) < 5 {
		return base + r.Intn(2000)
	}

	return base
}

func formatLogEntry(entry LogEntry) string {
	return fmt.Sprintf(
		"%s %s %s %d %d %s %s %s",
		entry.Timestamp.Format(time.RFC3339),
		entry.Service,
		entry.Endpoint,
		entry.StatusCode,
		entry.LatencyMs,
		entry.UserID,
		entry.Region,
		entry.RequestID,
	)
}
