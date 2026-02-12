package admin

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunSetLogSizeCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotProfileName string
	var gotMethod string
	var gotURL string
	var gotBody map[string]interface{}

	deps := setLogSizeDeps{
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
			return []byte(`{"app":"api","message":"Log file max size updated","size_bytes":15728640}`), nil
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--profile", "prod", "--app", "api", "--size", "15", "--unit", "MB"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotProfileName != "prod" {
		t.Fatalf("expected profile 'prod', got %q", gotProfileName)
	}
	if gotMethod != "PUT" {
		t.Fatalf("expected PUT method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/log-size" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}

	wantBody := map[string]interface{}{"app": "api", "size": int64(15), "unit": "MB"}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Fatalf("unexpected request body\nwant: %#v\ngot:  %#v", wantBody, gotBody)
	}

	output := out.String()
	if !strings.Contains(output, "\"message\": \"Log file max size updated\"") || !strings.Contains(output, "\"size_bytes\": 15728640") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunSetLogSizeCommand_NormalizesUnitToUpper(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotBody map[string]interface{}
	deps := setLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			cast := body.(map[string]interface{})
			gotBody = cast
			return []byte(`{"app":"api","message":"Log file max size updated","size_bytes":1073741824}`), nil
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--profile", "p1", "--app", "api", "--size", "1", "--unit", "gb"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotBody["unit"] != "GB" {
		t.Fatalf("expected normalized unit GB, got %#v", gotBody["unit"])
	}
}

func TestRunSetLogSizeCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfileName string
	deps := setLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfileName = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"app":"server","message":"Log file max size updated","size_bytes":10485760}`), nil
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--app", "server", "--size", "10", "--unit", "MB"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfileName != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfileName)
	}
}

func TestRunSetLogSizeCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogSizeDeps{}
	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--app", "api", "--size", "15", "--unit", "MB"}, deps, &out)
	if !errors.Is(err, errSetLogSizeUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunSetLogSizeCommand_MissingOrInvalidFields(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := setLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--profile", "p1", "--app", "api", "--size", "0", "--unit", "MB"}, deps, &out)
	if !errors.Is(err, errSetLogSizeUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
	if !strings.Contains(out.String(), "Missing or invalid fields") {
		t.Fatalf("expected usage message, got %q", out.String())
	}
}

func TestRunSetLogSizeCommand_InvalidUnit(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--profile", "p1", "--app", "api", "--size", "10", "--unit", "TB"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "unit must be MB or GB") {
		t.Fatalf("expected invalid unit error, got %v", err)
	}
}

func TestRunSetLogSizeCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{}
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--profile", "p1", "--app", "api", "--size", "10", "--unit", "MB"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunSetLogSizeCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--profile", "p1", "--app", "api", "--size", "10", "--unit", "MB"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error setting log size") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunSetLogSizeCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := setLogSizeDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runSetLogSizeCommand([]string{"--profile", "p1", "--app", "api", "--size", "10", "--unit", "MB"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
