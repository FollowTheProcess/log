package main

import (
	"math/rand/v2"
	"os"
	"time"

	"github.com/FollowTheProcess/log"
)

func main() {
	logger := log.New(os.Stderr, log.WithLevel(log.LevelDebug))

	logger.Debug("Searing steak", "cook", "rare", "temp", "42°C", "time", 2*time.Minute)
	sleep()
	logger.Info("Choosing wine pairing", "choices", []string{"merlot", "malbec", "rioja"})
	sleep()
	logger.Error("No malbec left!")
	sleep()
	logger.Warn("Falling back to second choice", "fallback", "rioja")

	logger.Info("Eating steak", "cut", "sirloin", "enjoying", true)
}

func sleep() {
	// Gen a random (small) duration
	n := 0.5 + rand.Float64() + rand.Float64()

	// Sleep for that random time
	time.Sleep(time.Duration(n) * time.Second)
}
