package agents

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListFiles_SuccessWithQuery(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/agents/files/list/agent-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":[{"FileID":1,"Path":"/tmp/a","Status":"ready"}]}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	out := captureStdout(t, func() {
		err := listFiles(profile, "agent-1", "Path,Status")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(out, "FileID") || !strings.Contains(out, "Path") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestListFiles_Errors(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		profile := profileFromServerURL(t, server.URL)
		server.Close()
		err := listFiles(profile, "agent-1", "")
		if err == nil || !strings.Contains(err.Error(), "error listing files") {
			t.Fatalf("expected request error, got %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()
		profile := profileFromServerURL(t, server.URL)
		err := listFiles(profile, "agent-1", "")
		if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

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

func TestUploadFile_UploadTypeWithoutDiskRead(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/agents/files/upload" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart failed: %v", err)
		}
		if r.FormValue("transfertype") != "Upload" || r.FormValue("uuid") != "agent-1" || r.FormValue("path") != "/tmp/x" {
			t.Fatalf("unexpected form values: transfertype=%q uuid=%q path=%q", r.FormValue("transfertype"), r.FormValue("uuid"), r.FormValue("path"))
		}
		if _, _, err := r.FormFile("file"); err == nil {
			t.Fatalf("did not expect file part for Upload transfertype")
		}

		_, _ = w.Write([]byte(`{"status":"Uploaded successfully"}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	out := captureStdout(t, func() {
		err := uploadFile(profile, "agent-1", "/tmp/x", "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Uploaded successfully") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestUploadFile_WithLocalSourceFile(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart failed: %v", err)
		}
		if r.FormValue("transfertype") != "Download" {
			t.Fatalf("unexpected transfertype: %q", r.FormValue("transfertype"))
		}
		if r.FormValue("uuid") != "agent-1" || r.FormValue("path") != "C:/temp/out.bin" {
			t.Fatalf("unexpected form values: uuid=%q path=%q", r.FormValue("uuid"), r.FormValue("path"))
		}

		file, hdr, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("expected file part: %v", err)
		}
		defer file.Close()
		if hdr.Filename != "input.bin" {
			t.Fatalf("unexpected filename: %s", hdr.Filename)
		}

		buf := make([]byte, 64)
		n, _ := file.Read(buf)
		if string(buf[:n]) != "abc123" {
			t.Fatalf("unexpected uploaded file content: %q", string(buf[:n]))
		}

		_, _ = w.Write([]byte(`{"status":"Uploaded successfully"}`))
	}))
	defer server.Close()

	src := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(src, []byte("abc123"), 0o644); err != nil {
		t.Fatalf("write source file failed: %v", err)
	}

	profile := profileFromServerURL(t, server.URL)
	if err := uploadFile(profile, "agent-1", "C:/temp/out.bin", src, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadFile_Errors(t *testing.T) {
	t.Run("missing source when needed", func(t *testing.T) {
		profile := profileFromServerURL(t, httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).URL)
		err := uploadFile(profile, "agent-1", "/tmp/x", "", "Download")
		if err == nil || !strings.Contains(err.Error(), "--source is required") {
			t.Fatalf("expected missing source error, got %v", err)
		}
	})

	t.Run("source read error", func(t *testing.T) {
		profile := profileFromServerURL(t, httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).URL)
		err := uploadFile(profile, "agent-1", "/tmp/x", "/does/not/exist", "")
		if err == nil || !strings.Contains(err.Error(), "failed to read source file") {
			t.Fatalf("expected source read error, got %v", err)
		}
	})

	t.Run("invalid response json", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				t.Fatalf("parse multipart failed: %v", err)
			}
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()

		src := filepath.Join(t.TempDir(), "input.bin")
		if err := os.WriteFile(src, []byte("abc123"), 0o644); err != nil {
			t.Fatalf("write source file failed: %v", err)
		}
		profile := profileFromServerURL(t, server.URL)
		err := uploadFile(profile, "agent-1", "/tmp/x", src, "")
		if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestResolveTransferType(t *testing.T) {
	t.Run("infer upload when no source", func(t *testing.T) {
		got, err := resolveTransferType("", "")
		if err != nil || got != "Upload" {
			t.Fatalf("expected Upload,nil got %q,%v", got, err)
		}
	})

	t.Run("infer download when source exists", func(t *testing.T) {
		got, err := resolveTransferType("/tmp/file.bin", "")
		if err != nil || got != "Download" {
			t.Fatalf("expected Download,nil got %q,%v", got, err)
		}
	})

	t.Run("invalid override", func(t *testing.T) {
		_, err := resolveTransferType("", "Sideways")
		if err == nil || !strings.Contains(err.Error(), "invalid --transfer-type") {
			t.Fatalf("expected invalid override error, got %v", err)
		}
	})

	t.Run("upload override with source", func(t *testing.T) {
		_, err := resolveTransferType("/tmp/file.bin", "Upload")
		if err == nil || !strings.Contains(err.Error(), "cannot be used") {
			t.Fatalf("expected conflict error, got %v", err)
		}
	})
}

// compile-time sanity check: multipart parsing helpers are linked as expected.
var _ = multipart.ErrMessageTooLarge
var _ = json.Marshal
