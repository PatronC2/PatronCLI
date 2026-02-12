package agents

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDescribeAgent_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/agent/agent-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"id":"agent-1","hostname":"alpha"}}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	out := captureStdout(t, func() {
		err := describeAgent(profile, "agent-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "agent-1") || !strings.Contains(out, "alpha") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestDescribeAgent_InvalidJSON(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	err := describeAgent(profile, "agent-1")
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestDescribeAgent_MissingDataObject(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":"not-an-object"}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	err := describeAgent(profile, "agent-1")
	if err == nil || !strings.Contains(err.Error(), "does not contain 'data'") {
		t.Fatalf("expected data field error, got %v", err)
	}
}

func TestDescribeAgent_RequestError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	profile := profileFromServerURL(t, server.URL)
	server.Close()

	err := describeAgent(profile, "agent-1")
	if err == nil || !strings.Contains(err.Error(), "error fetching agent") {
		t.Fatalf("expected request error, got %v", err)
	}
}
