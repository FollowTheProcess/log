package log_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FollowTheProcess/hue"
	"github.com/FollowTheProcess/log"
	"github.com/FollowTheProcess/test"
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

	// Constantly return the same time
	fixedTime := func() time.Time {
		fixed, err := time.Parse(time.RFC3339, "2025-04-01T13:34:03Z")
		test.Ok(t, err)

		return fixed
	}

	logger := log.New(buf, log.TimeFunc(fixedTime))

	const n = 5

	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			logger.Info(fmt.Sprintf("Something: %d", i))
		}(&wg, i)
	}

	wg.Wait()

	// Make sure they all got written, order doesn't matter because concurrency
	got := strings.TrimSpace(buf.String())
	lines := strings.Split(got, "\n")

	test.Equal(t, len(lines), n, test.Context("expected %d log lines", n))
}

func BenchmarkLogger(b *testing.B) {
	hue.Enabled(true) // Force colour

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
