package files

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"patroncli/types"
	"strings"
	"testing"
)

func TestDownloadFile_WritesToDisk(t *testing.T) {
	content := []byte("hello from server")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/files/download/99" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	output := filepath.Join(t.TempDir(), "downloads", "file.bin")

	_ = captureStdout(t, func() {
		err := downloadFile(profile, "99", output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output file failed: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("unexpected downloaded bytes: %q", string(got))
	}
}

func TestDownloadFile_RequestError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	profile := profileFromServerURL(t, server.URL)
	server.Close()

	err := downloadFile(profile, "1", filepath.Join(t.TempDir(), "x.bin"))
	if err == nil || !strings.Contains(err.Error(), "error downloading file") {
		t.Fatalf("expected request error, got %v", err)
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
