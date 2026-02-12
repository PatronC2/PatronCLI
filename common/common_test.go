package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"patroncli/types"
)

func TestMakeRequest_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content-type: %s", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	profile := types.Credential{Token: "Bearer test-token"}
	body := map[string]interface{}{"hello": "world"}

	got, err := MakeRequest(http.MethodPost, server.URL, profile, body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(got) != `{"ok":true}` {
		t.Fatalf("unexpected response body: %s", string(got))
	}
}

func TestMakeRequest_HTTPErrorIncludesStatusAndBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad input", http.StatusBadRequest)
	}))
	defer server.Close()

	_, err := MakeRequest(http.MethodGet, server.URL, types.Credential{}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "status=400") || !strings.Contains(err.Error(), "bad input") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMakeRequest_InvalidURL(t *testing.T) {
	t.Parallel()

	_, err := MakeRequest(http.MethodGet, "://bad-url", types.Credential{}, nil)
	if err == nil || !strings.Contains(err.Error(), "failed to create request") {
		t.Fatalf("expected request creation error, got %v", err)
	}
}

func TestMakeRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	badBody := map[string]interface{}{"fn": func() {}}
	_, err := MakeRequest(http.MethodPost, "https://example.com", types.Credential{}, badBody)
	if err == nil || !strings.Contains(err.Error(), "failed to serialize body") {
		t.Fatalf("expected serialize error, got %v", err)
	}
}

func TestMakeRequest_SOCKS5EnabledWithoutHostPort(t *testing.T) {
	t.Parallel()

	profile := types.Credential{SOCKS5Enabled: true}
	_, err := MakeRequest(http.MethodGet, "https://example.com", profile, nil)
	if err == nil || !strings.Contains(err.Error(), "SOCKS5 proxy enabled but host/port not set") {
		t.Fatalf("expected SOCKS5 config error, got %v", err)
	}
}

func TestGetCreds_ReturnsMatchingProfile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	credentialsPath := filepath.Join(tmpHome, ".patron", "credentials")
	if err := os.MkdirAll(filepath.Dir(credentialsPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	data, err := json.Marshal([]types.Credential{
		{Profile: "dev", Token: "dev-token"},
		{Profile: "prod", Token: "prod-token"},
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(credentialsPath, data, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got := GetCreds("prod")
	if got.Profile != "prod" || got.Token != "prod-token" {
		t.Fatalf("unexpected credential: %#v", got)
	}
}

func TestGetCreds_MissingOrInvalidFileReturnsEmpty(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	missing := GetCreds("any")
	if missing != (types.Credential{}) {
		t.Fatalf("expected zero-value for missing file, got %#v", missing)
	}

	credentialsPath := filepath.Join(tmpHome, ".patron", "credentials")
	if err := os.MkdirAll(filepath.Dir(credentialsPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(credentialsPath, []byte("not-json"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	invalid := GetCreds("any")
	if invalid != (types.Credential{}) {
		t.Fatalf("expected zero-value for invalid json, got %#v", invalid)
	}
}

func TestGetCreds_NoMatchingProfileReturnsEmpty(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	credentialsPath := filepath.Join(tmpHome, ".patron", "credentials")
	if err := os.MkdirAll(filepath.Dir(credentialsPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	data, err := json.Marshal([]types.Credential{
		{Profile: "dev", Token: "dev-token"},
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(credentialsPath, data, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got := GetCreds("prod")
	if got != (types.Credential{}) {
		t.Fatalf("expected zero-value credential, got %#v", got)
	}
}
