package admin

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunCreateUserCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotProfileName string
	var gotMethod string
	var gotURL string
	var gotBody map[string]interface{}

	deps := createUserDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "172.16.200.185", Port: "8443", Token: "Bearer x"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			cast, ok := body.(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected body type %T", body)
			}
			gotBody = cast
			return []byte(`{"message":"Created user testusername with role readOnly"}`), nil
		},
	}

	var out bytes.Buffer
	err := runCreateUserCommand([]string{"--profile", "prod", "--username", "testusername", "--role", "readOnly", "--password", "testpassword"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotProfileName != "prod" {
		t.Fatalf("expected profile 'prod', got %q", gotProfileName)
	}
	if gotMethod != "POST" {
		t.Fatalf("expected POST method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/users" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}

	wantBody := map[string]interface{}{
		"username":        "testusername",
		"role":            "readOnly",
		"password":        "testpassword",
		"confirmPassword": "testpassword",
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Fatalf("unexpected request body\nwant: %#v\ngot:  %#v", wantBody, gotBody)
	}

	if !strings.Contains(out.String(), "Created user testusername with role readOnly") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunCreateUserCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfileName string
	deps := createUserDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"message":"Created user x with role y"}`), nil
		},
	}

	var out bytes.Buffer
	err := runCreateUserCommand([]string{"--username", "u", "--role", "r", "--password", "p"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfileName != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfileName)
	}
}

func TestRunCreateUserCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := createUserDeps{}
	var out bytes.Buffer
	err := runCreateUserCommand([]string{"--username", "u", "--role", "r", "--password", "p"}, deps, &out)
	if !errors.Is(err, errCreateUserUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunCreateUserCommand_MissingRequiredFields(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := createUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runCreateUserCommand([]string{"--profile", "p1", "--username", "u", "--role", "r"}, deps, &out)
	if !errors.Is(err, errCreateUserUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
	if !strings.Contains(out.String(), "Missing required fields") {
		t.Fatalf("expected usage message, got %q", out.String())
	}
}

func TestRunCreateUserCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := createUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{}
		},
	}

	var out bytes.Buffer
	err := runCreateUserCommand([]string{"--profile", "p1", "--username", "u", "--role", "r", "--password", "p"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunCreateUserCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := createUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runCreateUserCommand([]string{"--profile", "p1", "--username", "u", "--role", "r", "--password", "p"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error creating user") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunCreateUserCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := createUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runCreateUserCommand([]string{"--profile", "p1", "--username", "u", "--role", "r", "--password", "p"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
