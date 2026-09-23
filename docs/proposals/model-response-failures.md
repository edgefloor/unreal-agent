# Propagate terminal model response failures

Status: proposal only. No implementation or tests are included.

## Problem

The Responses adapter represents a terminal `response.failed` event as
`llm.Response.Failure` with no Go error. The coordinator checks the Go error but
does not treat the failure field as an unsuccessful run.

At base commit `df8b0ba`, a local HTTP fixture returning `response.failed` with
`insufficient_quota` produced a persisted failure and runner exit code 0. An
equivalent SSE `error` event produced exit code 1 and an error event.

Relevant code: `harness/coordinator/loop.go`, `harness/llm/responsesapi/response.go`,
and `cmd/internal/agentrunner/run.go`.

## Proposed scope

Recognize terminal model failures after provider retries finish. Preserve the
failed response in session history and return an actionable error to the runner.
Prevent tool calls from the failed response from executing, including after
session recovery. Preserve the adapter's distinction between failures and stop
reasons. Do not change the session storage format or retry policy.

## Acceptance criteria

- A permanent model failure produces a nonzero exit code and a useful error.
- An exhausted transient failure produces the same unsuccessful-run behavior.
- The failed response remains available in persisted history.
- Tool calls from a failed response do not execute immediately or on resume.
- Successful responses and supported incomplete stop reasons retain their behavior.

## Planned validation

Add runner tests using a local HTTP server for permanent failure, retry exhaustion,
and success. Add focused coordinator recovery coverage for a failed response that
contains tool calls. Run `make test check build` when implementation is authorized.
