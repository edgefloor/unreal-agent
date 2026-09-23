# Handle SIGTERM through runner cancellation

Status: proposal only. No implementation or tests are included.

## Problem

`main` in `cmd/unreal-agent-runner/main.go` registers only `os.Interrupt` with
`signal.NotifyContext`. SIGTERM therefore terminates the process directly instead
of entering the runner's context cancellation and deferred cleanup path.

At base commit `df8b0ba`, a subprocess waiting on a local HTTP fixture exited with
code 130 after SIGINT. The same fixture terminated directly with signal 15 after
SIGTERM. That reproduction establishes signal handling behavior, not whether all
child processes are cleaned up.

## Proposed scope

Register SIGTERM for context cancellation on supported platforms and document an
explicit exit-code policy. Preserve SIGINT behavior and normal completion. Keep
this change focused on signal handling; do not redesign coordinator shutdown.

## Acceptance criteria

- SIGTERM cancels an active model request through the runner context.
- Runner cleanup completes before exit.
- SIGTERM produces a documented nonzero exit status.
- SIGINT and normal completion retain their existing behavior.
- Tests run on the repository's Linux and macOS CI platforms.

## Planned validation

Add a subprocess test that waits for a local HTTP request before sending SIGTERM.
Verify cancellation, cleanup, and the chosen exit status. Keep child-process
cleanup claims limited to behavior actually tested. Run `make test check build`
when implementation is authorized.
