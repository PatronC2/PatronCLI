package files

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDisplayHelp_PrintsUsageAndCommands(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	os.Stdout = w

	commands := map[string]struct {
		Execute func(args []string)
		Help    string
	}{
		"search": {Execute: nil, Help: "Search files"},
		"help":   {Execute: nil, Help: "Show help"},
	}
	displayHelp(commands)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout failed: %v", err)
	}
	_ = r.Close()

	out := buf.String()
	if !strings.Contains(out, "Usage: patron files") {
		t.Fatalf("expected usage in output, got %q", out)
	}
	if !strings.Contains(out, "search") || !strings.Contains(out, "help") {
		t.Fatalf("expected command list in output, got %q", out)
	}
}
