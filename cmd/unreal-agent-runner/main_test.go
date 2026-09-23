package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRunnerSignals(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "runner")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build runner: %v\n%s", err, output)
	}

	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(signal.String(), func(t *testing.T) {
			started := make(chan struct{}, 1)
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				started <- struct{}{}
				select {
				case <-r.Context().Done():
				case <-release:
				}
			}))
			t.Cleanup(server.Close)
			t.Cleanup(func() { close(release) })

			workspace := t.TempDir()
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			t.Cleanup(cancel)
			cmd := exec.CommandContext(ctx, binary,
				"-workspace", workspace,
				"-session-directory", filepath.Join(workspace, "sessions"),
				"-p", "hello")
			cmd.Env = []string{
				"PATH=" + os.Getenv("PATH"),
				"SHELL=/bin/sh",
				"UNREAL_HARNESS_LLM_PROVIDER=openai",
				"UNREAL_HARNESS_LLM_API_KEY=fake-key",
				"UNREAL_HARNESS_LLM_BASE_URL=" + server.URL,
				"UNREAL_HARNESS_LLM_MAX_ATTEMPTS=1",
			}
			var output bytes.Buffer
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			go func() {
				_ = cmd.Wait()
				close(done)
			}()
			t.Cleanup(func() {
				cancel()
				<-done
			})

			select {
			case <-started:
			case <-done:
				t.Fatalf("runner exited before requesting the model: %s\n%s", cmd.ProcessState, &output)
			case <-ctx.Done():
				t.Fatal("runner did not request the model before timeout")
			}
			if err := cmd.Process.Signal(signal); err != nil {
				t.Fatal(err)
			}
			<-done
			if cmd.ProcessState.ExitCode() != 130 {
				t.Fatalf("runner exit = %s, want exit status 130\n%s", cmd.ProcessState, &output)
			}
		})
	}
}
