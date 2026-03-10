package files

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

func TestSearchFiles_Success(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/files/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		capturedQuery = r.URL.Query()
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"FileID": 1, "Path": "/tmp/a", "Status": "ready"},
			},
			"totalCount": 1,
			"nextOffset": 1,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	profile := fileProfileFromServerURL(t, server.URL)
	out := captureFilesStdout(t, func() {
		err := searchFiles(profile, "and", []string{"env:prod", " ", "team:red"}, 20, 5, "Path, Status")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if capturedQuery.Get("logic") != "and" {
		t.Fatalf("unexpected logic query: %#v", capturedQuery)
	}
	if capturedQuery.Get("limit") != "20" || capturedQuery.Get("offset") != "5" {
		t.Fatalf("unexpected pagination query: %#v", capturedQuery)
	}
	if tags := capturedQuery["tag"]; len(tags) != 2 || tags[0] != "env:prod" || tags[1] != "team:red" {
		t.Fatalf("unexpected tag query params: %#v", capturedQuery["tag"])
	}
	if !strings.Contains(out, `"Path": "/tmp/a"`) || !strings.Contains(out, `"Status": "ready"`) {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, `"FileID"`) {
		t.Fatalf("expected query projection to remove FileID, got %q", out)
	}
}

func TestSearchFiles_Errors(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		profile := fileProfileFromServerURL(t, server.URL)
		server.Close()

		err := searchFiles(profile, "or", nil, 20, 0, "")
		if err == nil || !strings.Contains(err.Error(), "error fetching files") {
			t.Fatalf("expected request error, got %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()

		profile := fileProfileFromServerURL(t, server.URL)
		err := searchFiles(profile, "or", nil, 20, 0, "")
		if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestSearchFiles_OmitsInvalidLimitOffset(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[],"totalCount":0,"nextOffset":0}`))
	}))
	defer server.Close()

	profile := fileProfileFromServerURL(t, server.URL)
	_ = captureFilesStdout(t, func() {
		err := searchFiles(profile, "or", nil, 0, -1, "")
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

func fileProfileFromServerURL(t *testing.T, raw string) types.Credential {
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

func captureFilesStdout(t *testing.T, fn func()) string {
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
		t.Fatalf("read stdout failed: %v", err)
	}
	_ = r.Close()

	return buf.String()
}
