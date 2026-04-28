package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// runVersion writes to fmt.Printf which goes to os.Stdout. Capture it
// by redirecting Stdout, run the function, and assert on the buffer.
func TestRunVersion_WritesVersionLine(t *testing.T) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	// Pin the test's view of the build vars regardless of -ldflags.
	prevVersion, prevCommit, prevDate := version, commit, buildDate
	version = "v9.9.9-test"
	commit = "abc1234"
	buildDate = "2026-04-28T00:00:00Z"
	defer func() {
		version, commit, buildDate = prevVersion, prevCommit, prevDate
	}()

	runVersion()

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	got := buf.String()
	for _, want := range []string{"v9.9.9-test", "abc1234", "2026-04-28T00:00:00Z", "chronos"} {
		if !strings.Contains(got, want) {
			t.Errorf("version output missing %q; full output: %q", want, got)
		}
	}
}
