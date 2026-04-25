package agents

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"patroncli/types"
)

func TestSearchAgents_Success(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/agents/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		capturedQuery = r.URL.Query()
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "a1", "hostname": "host-a", "ip": "10.0.0.1"},
			},
			"totalCount": 1,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	out := captureStdout(t, func() {
		err := searchAgents(
			profile,
			"host-a",
			"10.0.0.1",
			"Online",
			"and",
			[]string{"env:prod", "  ", "team:red"},
			"hostname:asc",
			20,
			5,
			"hostname, ip",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if capturedQuery.Get("hostname") != "host-a" || capturedQuery.Get("ip") != "10.0.0.1" {
		t.Fatalf("missing expected query filters: %#v", capturedQuery)
	}
	if capturedQuery.Get("status") != "Online" || capturedQuery.Get("logic") != "and" {
		t.Fatalf("missing expected status/logic filters: %#v", capturedQuery)
	}
	if capturedQuery.Get("sort") != "hostname:asc" || capturedQuery.Get("limit") != "20" || capturedQuery.Get("offset") != "5" {
		t.Fatalf("missing expected pagination/sort filters: %#v", capturedQuery)
	}
	if tags := capturedQuery["tag"]; len(tags) != 2 || tags[0] != "env:prod" || tags[1] != "team:red" {
		t.Fatalf("unexpected tag query params: %#v", capturedQuery["tag"])
	}

	if !strings.Contains(out, "host-a") || !strings.Contains(out, "10.0.0.1") {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, "\"id\"") {
		t.Fatalf("expected query field filtering to remove id, got output: %q", out)
	}
}

func TestSearchAgents_InvalidJSON(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	err := searchAgents(profile, "", "", "", "", nil, "", 0, -1, "")
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestSearchAgents_RequestError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	profile := profileFromServerURL(t, server.URL)
	server.Close()

	err := searchAgents(profile, "", "", "", "", nil, "", 0, -1, "")
	if err == nil || !strings.Contains(err.Error(), "error fetching agents") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestSearchAgents_OmitsInvalidLimitOffset(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[],"totalCount":0}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	_ = captureStdout(t, func() {
		err := searchAgents(profile, "", "", "", "", nil, "", 0, -1, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if _, ok := capturedQuery["limit"]; ok {
		t.Fatalf("did not expect limit query param, got %#v", capturedQuery)
	}
	if _, ok := capturedQuery["offset"]; ok {
		t.Fatalf("did not expect offset query param, got %#v", capturedQuery)
	}
}

func profileFromServerURL(t *testing.T, raw string) types.Credential {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse server url: %v", err)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("failed to split host/port: %v", err)
	}
	return types.Credential{IP: host, Port: port}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read captured stdout failed: %v", err)
	}
	_ = r.Close()

	return buf.String()
}
