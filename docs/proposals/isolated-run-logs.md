# Isolate logs for simultaneous runner invocations

Status: proposal only. No implementation or tests are included.

## Problem

`openDatetimeLog` in `cmd/internal/agentrunner/run.go` names log files with UTC
timestamps at second precision and opens them with `O_APPEND`. Invocations that
share a log directory and start within one second append to the same file.

At base commit `df8b0ba`, calls with supplied timestamps 100 milliseconds apart
opened the same path. The first invocation's log contained both invocations'
output. Individual session records do not contain a session ID that could separate
the combined histories.

## Proposed scope

Create each invocation's log exclusively with a unique suffix. Retain a readable
UTC timestamp, the `.jsonl` extension, and private file permissions. Update the
existing filename assertion. Do not change canonical session storage or JSONL
record contents.

## Acceptance criteria

- Two calls with the same supplied timestamp create distinct files.
- Each file contains only its invocation's output.
- Existing files are neither appended to nor overwritten by a new invocation.
- Log filenames retain the timestamp and `.jsonl` extension.
- New log files retain private permissions.

## Planned validation

Add deterministic tests using identical supplied timestamps and distinct output
contents. Check collision handling and permissions without timing-sensitive sleeps.
Run `make test check build` when implementation is authorized.
