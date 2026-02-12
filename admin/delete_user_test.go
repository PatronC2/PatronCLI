package admin

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunDeleteUserCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotProfileName string
	var gotMethod string
	var gotURL string

	deps := deleteUserDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "172.16.200.185", Port: "8443", Token: "Bearer x"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			return []byte(`{"message":"User deleted successfully"}`), nil
		},
	}

	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--profile", "prod", "--username", "test"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotProfileName != "prod" {
		t.Fatalf("expected profile 'prod', got %q", gotProfileName)
	}
	if gotMethod != "DELETE" {
		t.Fatalf("expected DELETE method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/users/test" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}

	if !strings.Contains(out.String(), "User deleted successfully") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunDeleteUserCommand_EncodesPathUsername(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotURL string
	deps := deleteUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotURL = url
			return []byte(`{"message":"User deleted successfully"}`), nil
		},
	}

	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--profile", "p1", "--username", "test user"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotURL != "https://127.0.0.1:8443/api/admin/users/test%20user" {
		t.Fatalf("unexpected encoded URL: %q", gotURL)
	}
}

func TestRunDeleteUserCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfileName string
	deps := deleteUserDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"message":"User deleted successfully"}`), nil
		},
	}

	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--username", "test"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfileName != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfileName)
	}
}

func TestRunDeleteUserCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := deleteUserDeps{}
	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--username", "test"}, deps, &out)
	if !errors.Is(err, errDeleteUserUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunDeleteUserCommand_MissingUsername(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := deleteUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--profile", "p1"}, deps, &out)
	if !errors.Is(err, errDeleteUserUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
	if !strings.Contains(out.String(), "Missing required field --username") {
		t.Fatalf("expected usage message, got %q", out.String())
	}
}

func TestRunDeleteUserCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := deleteUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{}
		},
	}

	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--profile", "p1", "--username", "test"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunDeleteUserCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := deleteUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--profile", "p1", "--username", "test"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error deleting user") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunDeleteUserCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := deleteUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runDeleteUserCommand([]string{"--profile", "p1", "--username", "test"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
