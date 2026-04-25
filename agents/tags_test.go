package agents

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTags_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/tags/agent-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"tags":[{"id":1,"key":"env","value":"prod"}]}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	out := captureStdout(t, func() {
		err := getTags(profile, "agent-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, `"env"`) || !strings.Contains(out, `"prod"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestGetTags_Errors(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		profile := profileFromServerURL(t, server.URL)
		server.Close()
		err := getTags(profile, "agent-1")
		if err == nil || !strings.Contains(err.Error(), "error fetching tags") {
			t.Fatalf("expected request error, got %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()
		profile := profileFromServerURL(t, server.URL)
		err := getTags(profile, "agent-1")
		if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestPutTags_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/tag" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		agents, ok := body["agents"].([]interface{})
		if !ok || len(agents) != 2 || agents[0] != "agent-1" || agents[1] != "agent-2" {
			t.Fatalf("unexpected agents payload: %#v", body["agents"])
		}
		if body["key"] != "env" || body["value"] != "prod" {
			t.Fatalf("unexpected key/value payload: %#v", body)
		}

		_, _ = w.Write([]byte(`{"message":"Tags updated successfully for all agents"}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	out := captureStdout(t, func() {
		err := putTags(profile, []string{"agent-1", "agent-2"}, "env", "prod")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Tags updated successfully") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestPutTags_Errors(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		profile := profileFromServerURL(t, server.URL)
		server.Close()
		err := putTags(profile, []string{"agent-1"}, "env", "prod")
		if err == nil || !strings.Contains(err.Error(), "error updating tags") {
			t.Fatalf("expected request error, got %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("nope"))
		}))
		defer server.Close()
		profile := profileFromServerURL(t, server.URL)
		err := putTags(profile, []string{"agent-1"}, "env", "prod")
		if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestDeleteTag_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/tag/123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"message":"deleted tag successfully"}`))
	}))
	defer server.Close()

	profile := profileFromServerURL(t, server.URL)
	out := captureStdout(t, func() {
		err := deleteTag(profile, 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "deleted tag successfully") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestDeleteTag_Errors(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		profile := profileFromServerURL(t, server.URL)
		server.Close()
		err := deleteTag(profile, 1)
		if err == nil || !strings.Contains(err.Error(), "error deleting tag") {
			t.Fatalf("expected request error, got %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("nope"))
		}))
		defer server.Close()
		profile := profileFromServerURL(t, server.URL)
		err := deleteTag(profile, 1)
		if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestSplitCSV(t *testing.T) {
	got := splitCSV("agent-1, agent-2, ,agent-3")
	want := []string{"agent-1", "agent-2", "agent-3"}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected split value at %d: got %q want %q", i, got[i], want[i])
		}
	}
}
