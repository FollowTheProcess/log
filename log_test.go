package log_test

import (
	"testing"

	"github.com/FollowTheProcess/log"
)

func TestHello(t *testing.T) {
	got := log.Hello()
	want := "Hello log"

	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
