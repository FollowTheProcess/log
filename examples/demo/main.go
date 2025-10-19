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

	logger.Debug(
		"Searing steak",
		slog.String("cook", "rare"),
		slog.Int("temp", 42),
		slog.Duration("time", 2*time.Minute),
	)

	sleep()

	logger.Info(
		"Choosing wine pairing",
		slog.Any("choices", []string{"merlot", "malbec", "rioja"}),
	)

	sleep()

	logger.Error("No malbec left!")

	sleep()

	logger.Warn(
		"Falling back to second choice",
		slog.String("fallback", "rioja"),
	)

	logger.Info(
		"Eating steak",
		slog.String("cut", "sirloin"),
		slog.Bool("enjoying", true),
	)
}

func sleep() {
	// Gen a random (small) duration
	n := 0.5 + rand.Float64() + rand.Float64()

	// Sleep for that random time
	time.Sleep(time.Duration(n) * time.Second)
}
