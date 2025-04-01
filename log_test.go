package log_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/FollowTheProcess/hue"
	"github.com/FollowTheProcess/log"
)

func TestVisual(t *testing.T) {
	hue.Enabled(true) // Force colour
	logger := log.New(os.Stdout, log.WithLevel(log.LevelDebug))

	logger.Debug("Doing some debuggy things")
	logger.Info("Logging in")
	logger.Warn("Config file missing, falling back to defaults")
	logger.Error("File not found")
}

func TestRace(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf)

	const n = 5

	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			logger.Info(fmt.Sprintf("Something: %d", n))
		}(&wg)
	}

	wg.Wait()
}

func BenchmarkLogger(b *testing.B) {
	hue.Enabled(true) // Force colour
	defer hue.Enabled(false)

	b.Run("enabled", func(b *testing.B) {
		buf := &bytes.Buffer{}

		logger := log.New(buf, log.WithLevel(log.LevelDebug))

		for b.Loop() {
			logger.Debug("A message!")
		}
	})

	b.Run("disabled", func(b *testing.B) {
		buf := &bytes.Buffer{}

		logger := log.New(buf, log.WithLevel(log.LevelInfo))

		for b.Loop() {
			logger.Debug("A message!")
		}
	})
}

func BenchmarkDiscard(b *testing.B) {
	// Here to test that effectively nothing is done
	// when w == io.Discard
	logger := log.New(io.Discard, log.WithLevel(log.LevelDebug))

	for b.Loop() {
		logger.Debug("Nothing")
	}
}
