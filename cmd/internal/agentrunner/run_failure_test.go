package agentrunner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunMainReturnsModelFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, err := fmt.Fprint(w, `data: {"type":"response.failed","response":{"id":"failed","status":"failed","output":[],"error":{"code":"insufficient_quota","message":"fixture failure"}}}`+"\n\n")
		if err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	env := map[string]string{
		"UNREAL_HARNESS_LLM_PROVIDER":     "openai",
		"UNREAL_HARNESS_LLM_API_KEY":      "fake-key",
		"UNREAL_HARNESS_LLM_BASE_URL":     server.URL,
		"UNREAL_HARNESS_LLM_MAX_ATTEMPTS": "1",
	}
	code := RunMain(ctx, []string{"-workspace", t.TempDir(), "-session-directory", t.TempDir(), "-p", "hello"},
		func(key string) string { return env[key] }, func() []string { return nil },
		strings.NewReader(""), &stdout, &stderr,
		Config{Name: "repro", ParseRequest: parseTestRequest, Providers: DefaultProviders()})
	if !strings.Contains(stdout.String(), `"Code":"insufficient_quota"`) {
		t.Fatalf("failure was not recorded: %s", stdout.String())
	}
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	for _, want := range []string{"insufficient_quota", "fixture failure"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr = %q, want %q", stderr.String(), want)
		}
	}
}
