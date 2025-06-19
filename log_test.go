package log_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"go.followtheprocess.codes/hue"
	"go.followtheprocess.codes/log"
	"go.followtheprocess.codes/test"
)

func TestVisual(t *testing.T) {
	hue.Enabled(true) // Force colour

	logger := log.New(os.Stdout, log.WithLevel(log.LevelDebug))

	prefixed := logger.Prefixed("cooking")

	logger.Debug("Doing some debuggy things")
	logger.Info("Logging in")
	logger.Warn("Config file missing, falling back to defaults")
	logger.Error("File not found")

	prefixed.Warn("Pizza is burning!", "flavour", "pepperoni")
	prefixed.Info("Response from oven API", "status", http.StatusOK, "duration", 57*time.Millisecond)
}

func TestDebug(t *testing.T) {
	hue.Enabled(false) // Force no color

	// Constantly return the same time
	fixedTime := func() time.Time {
		fixed, err := time.Parse(time.RFC3339, "2025-04-01T13:34:03Z")
		test.Ok(t, err)

		return fixed
	}

	fixedTimeString := fixedTime().Format(time.RFC3339)

	tests := []struct {
		name    string       // Name of the test case
		msg     string       // Message to log
		kv      []any        // Key value pairs to pass to the log method
		want    string       // Expected log line
		options []log.Option // Options to customise the logger under test
	}{
		{
			name: "disabled",
			options: []log.Option{
				log.WithLevel(log.LevelInfo),
			},
			msg:  "You should not see me",
			want: "", // Debug logs should not show up if the level is info
		},
		{
			name: "enabled",
			options: []log.Option{
				log.WithLevel(log.LevelDebug),
			},
			msg:  "Hello debug!",
			want: "[TIME] DEBUG: Hello debug!\n",
		},
		{
			name: "prefix",
			options: []log.Option{
				log.WithLevel(log.LevelDebug),
				log.Prefix("building"),
			},
			msg:  "Hello debug!",
			want: "[TIME] DEBUG building: Hello debug!\n",
		},
		{
			name: "with kv",
			options: []log.Option{
				log.WithLevel(log.LevelDebug),
			},
			msg:  "Hello debug!",
			kv:   []any{"number", 12, "duration", 30 * time.Second, "enabled", true},
			want: "[TIME] DEBUG: Hello debug! number=12 duration=30s enabled=true\n",
		},
		{
			name: "with kv quoted",
			options: []log.Option{
				log.WithLevel(log.LevelDebug),
			},
			msg:  "Hello debug!",
			kv:   []any{"number", 12, "duration", 30 * time.Second, "sentence", "this has spaces"},
			want: `[TIME] DEBUG: Hello debug! number=12 duration=30s sentence="this has spaces"` + "\n",
		},
		{
			name: "with kv escape chars",
			options: []log.Option{
				log.WithLevel(log.LevelDebug),
			},
			msg:  "Hello debug!",
			kv:   []any{"number", 12, "duration", 30 * time.Second, "sentence", "ooh\t\nstuff"},
			want: `[TIME] DEBUG: Hello debug! number=12 duration=30s sentence="ooh\t\nstuff"` + "\n",
		},
		{
			name: "with kv odd number",
			options: []log.Option{
				log.WithLevel(log.LevelDebug),
			},
			msg:  "One is missing",
			kv:   []any{"enabled", true, "file", "./file.txt", "elapsed"},
			want: "[TIME] DEBUG: One is missing enabled=true file=./file.txt elapsed=<MISSING>\n",
		},
		{
			name: "custom time format",
			options: []log.Option{
				log.WithLevel(log.LevelDebug),
				log.TimeFormat(time.Kitchen),
			},
			msg:  "The oven is done",
			want: "1:34PM DEBUG: The oven is done\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			// Ensure that the time is always deterministic
			tt.options = append(tt.options, log.TimeFunc(fixedTime))

			logger := log.New(buf, tt.options...)

			logger.Debug(tt.msg, tt.kv...)

			got := buf.String()
			got = strings.ReplaceAll(got, fixedTimeString, "[TIME]")

			test.Diff(t, got, tt.want)
		})
	}
}

func TestWith(t *testing.T) {
	// Constantly return the same time
	fixedTime := func() time.Time {
		fixed, err := time.Parse(time.RFC3339, "2025-04-01T13:34:03Z")
		test.Ok(t, err)

		return fixed
	}

	fixedTimeString := fixedTime().Format(time.RFC3339)

	buf := &bytes.Buffer{}

	logger := log.New(buf, log.TimeFunc(fixedTime))

	logger.Info("I'm an info message")

	sub := logger.With("sub", true, "missing")

	sub.Info("I'm also an info message")

	got := buf.String()
	got = strings.TrimSpace(strings.ReplaceAll(got, fixedTimeString, "[TIME]")) + "\n"

	want := "[TIME] INFO: I'm an info message\n[TIME] INFO: I'm also an info message sub=true missing=<MISSING>\n"
	test.Diff(t, got, want)
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
	sub := logger.Prefixed("sub")

	const n = 5

	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			logger.Info(fmt.Sprintf("Something: %d", i))
		}(&wg, i)
	}

	wg.Add(n)
	for i := range n {
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			sub.Info(fmt.Sprintf("Other: %d", i))
		}(&wg, i)
	}

	wg.Wait()

	// Make sure they all got written, order doesn't matter because concurrency
	got := strings.TrimSpace(buf.String())
	lines := strings.Split(got, "\n")

	test.Equal(t, len(lines), n*2, test.Context("expected %d log lines", n*2))
}

func TestContext(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		buf := &bytes.Buffer{}

		// Constantly return the same time
		fixedTime := func() time.Time {
			fixed, err := time.Parse(time.RFC3339, "2025-04-01T13:34:03Z")
			test.Ok(t, err)

			return fixed
		}

		// Configure it a bit so we know we're getting the right one
		logger := log.New(buf, log.TimeFunc(fixedTime), log.TimeFormat(time.Kitchen))

		logger.Info("Before")

		ctx := t.Context()

		ctx = log.WithContext(ctx, logger)

		after := log.FromContext(ctx)

		after.Info("After")

		got := buf.String()

		test.Diff(t, got, "1:34PM INFO: Before\n1:34PM INFO: After\n")
	})

	t.Run("missing", func(t *testing.T) {
		_, stderr := test.CaptureOutput(t, func() error {
			log.FromContext(t.Context()).Info("FromContext")
			return nil
		})

		test.True(t, strings.Contains(stderr, "FromContext"))
	})
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

	b.Run("discard", func(b *testing.B) {
		// Here to test that effectively nothing is done
		// when w == io.Discard
		logger := log.New(io.Discard, log.WithLevel(log.LevelDebug))

		for b.Loop() {
			logger.Debug("Nothing")
		}
	})
}
