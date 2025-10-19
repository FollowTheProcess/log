package main

import (
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	"go.followtheprocess.codes/log"
)

func main() {
	logger := log.New(os.Stderr, log.WithLevel(log.LevelDebug))

	logger.Info(
		"Doing something",
		slog.Bool("cache", true),
		slog.Duration("duration", 30*time.Second),
		slog.Int("number", 42),
	)

	sleep()

	sub := logger.With(slog.Bool("sub", true))

	sub.Info(
		"Hello from the sub logger",
		slog.String("subkey", "yes"),
	)
}

func sleep() {
	// Gen a random (small) duration
	n := 0.5 + rand.Float64() + rand.Float64()

	// Sleep for that random time
	time.Sleep(time.Duration(n) * time.Second)
}
