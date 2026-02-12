package auth

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRunAuth_NoArgs_ShowsHelpAndUsageError(t *testing.T) {
	commands := map[string]authSubcommand{
		"help": {Execute: func(args []string) error { return nil }, Help: "Show help"},
	}
	var out bytes.Buffer

	err := runAuth(nil, commands, &out)
	if !errors.Is(err, errAuthUsage) {
		t.Fatalf("expected errAuthUsage, got %v", err)
	}
	if !strings.Contains(out.String(), "Usage: patron auth") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}

func TestRunAuth_UnknownSubcommand(t *testing.T) {
	commands := map[string]authSubcommand{
		"help": {Execute: func(args []string) error { return nil }, Help: "Show help"},
	}
	var out bytes.Buffer

	err := runAuth([]string{"nope"}, commands, &out)
	if !errors.Is(err, errAuthUsage) {
		t.Fatalf("expected errAuthUsage, got %v", err)
	}
	if !strings.Contains(out.String(), "Unknown auth subcommand: nope") {
		t.Fatalf("expected unknown command output, got %q", out.String())
	}
}

func TestRunAuth_DispatchesSubcommand(t *testing.T) {
	called := false
	commands := map[string]authSubcommand{
		"login": {
			Execute: func(args []string) error {
				called = true
				if len(args) != 1 || args[0] != "--profile=test" {
					t.Fatalf("unexpected args: %#v", args)
				}
				return nil
			},
			Help: "Login",
		},
	}
	var out bytes.Buffer

	err := runAuth([]string{"login", "--profile=test"}, commands, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected login command to be called")
	}
}
