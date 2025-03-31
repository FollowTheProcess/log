package log_test

import (
	"bytes"
	"os"
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

func BenchmarkLogger(b *testing.B) {
	hue.Enabled(true) // Force colour
	buf := &bytes.Buffer{}

	logger := log.New(buf, log.WithLevel(log.LevelDebug))

	for b.Loop() {
		logger.Debug("A message!")
	}
}
