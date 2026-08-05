---
name: scheduler-change
description: Safely implement, diagnose, or review changes to gocron scheduling and task execution. Use for cron registration, leader election, task dispatch, concurrency queues, manual runs, retries, timeouts, cancellation, task logs, agent execution, failover, or shared scheduler state.
---

# Change the gocron scheduler

Assume scheduler bugs can duplicate, lose, or overlap production jobs. Make the
execution semantics explicit before editing.

## Compatibility invariants

- When wire behavior changes, test `N` gocron with `N-1` gocron-node and the
  reverse. Preserve execution, output, cancellation, and terminal status.
- Evolve protobuf/gRPC additively: never renumber or reuse field numbers,
  remove a supported field, or change an existing RPC's semantics. Reserve
  removed identifiers.
- Keep the previous RPC/execution path during the compatibility window. Check
  capability and use a deterministic fallback before invoking new behavior.
- Never retry or reroute after an ambiguous remote acceptance; that can run a
  user's job twice. Preserve retry counts and status transitions on fallback.
- When distributed ownership or state changes, verify temporary `N`/`N-1` HA
  mixtures do not duplicate dispatch, stick in `Running`, or falsely cancel.

## Performance invariants

- Establish a baseline for materially affected hot paths with representative
  task/node counts and measure applicable scheduler metrics.
- Bound goroutines, queues, retries, log buffers, streams, timers, and retained
  execution state. Define backpressure and overload behavior explicitly.
- Do not hold locks during database/network/filesystem/command/notification I/O
  or serialize unrelated tasks behind one global lock.
- Batch or coalesce high-frequency status/log writes when semantics permit.
  Verify SQLite concurrent read/write and busy/lock behavior separately.
- Add a benchmark or deterministic load regression test for material hot-path
  changes. Compare before/after under the same workload and report the result.

## Define invariants

Write down which behavior the change must preserve:

- whether execution is at-most-once, at-least-once, or best-effort;
- what happens during leader loss, restart, network timeout, or agent retry;
- how scheduled and manual runs interact;
- whether the same task may overlap and how queue limits behave;
- when task counts, instance maps, logs, and final status are updated;
- who owns cancellation and how resources are released.

If the intended behavior is ambiguous and different choices change user-visible
execution, ask before implementing.

## Inspect the whole execution path

Trace from route or cron callback through `internal/service/task.go`, scheduler
state, models, RPC/agent code, task logs, and notifications. Search every read
and write of changed global state. Identify goroutine ownership, locks, channel
capacity, blocking sends, callbacks, timers, and cleanup paths.

Do not:

- hold a mutex while performing database, network, command, or notification I/O;
- start an unbounded goroutine per event;
- close a channel from a receiver or from multiple owners;
- use check-then-act state without the same lock or atomic operation;
- treat a timeout as proof that remote execution did not start;
- silently drop queued work or overwrite the terminal task status.

Keep refactors incremental. Extracting a component is useful when it creates a
test seam; moving globals without defining ownership is not.

## Test adversarial sequences

Add deterministic tests for the changed invariant, including the relevant
subset of:

- simultaneous triggers for one task;
- queue full, cancellation, timeout, and shutdown;
- disable/remove while queued or running;
- repeated callback or retry;
- leader handoff and stale leader behavior;
- remote execution accepted but response lost;
- `N-1` protocol doubles in both client/server directions and fallback paths;
- panic/error cleanup and counter consistency.

Avoid sleep-based assertions where a channel, fake clock, barrier, or polling
condition can make the test deterministic.

Run focused tests first, then:

```bash
go test -race ./internal/service/... ./internal/modules/... ./internal/rpc/...
go test -race -count=10 <changed-package>
```

Adapt package paths if the changed execution path differs. Invoke `$verify`
before committing. Report the promised execution semantics, race-sensitive
state touched, failure cases tested, performance evidence when relevant, and
any behavior that remains best-effort.
