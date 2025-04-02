package main

import (
	"math/rand/v2"
	"os"
	"time"

	"github.com/FollowTheProcess/log"
)

func main() {
	logger := log.New(os.Stderr, log.WithLevel(log.LevelDebug))

	logger.Info("Doing something", "cache", true, "duration", 30*time.Second, "number", 42)
	sleep()

	sub := logger.With("sub", true)
	sub.Info("Hello from the sub logger", "subkey", "yes")
}

func sleep() {
	// Gen a random (small) duration
	n := 0.5 + rand.Float64() + rand.Float64()

	// Sleep for that random time
	time.Sleep(time.Duration(n) * time.Second)
}
