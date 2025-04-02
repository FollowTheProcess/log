package main

import (
	"math/rand/v2"
	"os"
	"time"

	"github.com/FollowTheProcess/log"
)

func main() {
	logger := log.New(os.Stderr)
	prefixed := logger.Prefixed("http")

	logger.Info("Calling GitHub API", "url", "https://api.github.com/")
	sleep()

	prefixed.Warn("Slow endpoint", "endpoint", "users/slow", "duration", 10*time.Second)
	sleep()
	prefixed.Info("Response from get repos", "status", 200, "duration", 500*time.Millisecond)
	sleep()
	prefixed.Error("Response from something else", "status", 400, "duration", 33*time.Millisecond)
}

func sleep() {
	// Gen a random (small) duration
	n := 0.5 + rand.Float64() + rand.Float64()

	// Sleep for that random time
	time.Sleep(time.Duration(n) * time.Second)
}
