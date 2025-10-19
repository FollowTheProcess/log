package main

import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"go.followtheprocess.codes/log"
)

func main() {
	logger := log.New(os.Stderr)
	prefixed := logger.Prefixed("http")

	logger.Info(
		"Calling GitHub API",
		slog.String("url", "https://api.github.com/"),
	)

	sleep()

	prefixed.Warn(
		"Slow endpoint",
		slog.String("endpoint", "users/slow"),
		slog.Duration("duration", 10*time.Second),
	)

	sleep()

	prefixed.Info(
		"Response from get repos",
		slog.Int("status", http.StatusOK),
		slog.Duration("duration", 500*time.Millisecond),
	)

	sleep()

	prefixed.Error(
		"Response from something else",
		slog.Int("status", http.StatusBadRequest),
		slog.Duration("duration", 33*time.Millisecond),
	)
}

func sleep() {
	// Gen a random (small) duration
	n := 0.5 + rand.Float64() + rand.Float64()

	// Sleep for that random time
	time.Sleep(time.Duration(n) * time.Second)
}
