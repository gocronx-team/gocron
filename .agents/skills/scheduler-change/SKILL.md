---
name: scheduler-change
description: Safely implement, diagnose, or review changes to gocron scheduling and task execution. Use for cron registration, leader election, task dispatch, concurrency queues, manual runs, retries, timeouts, cancellation, task logs, agent execution, failover, or shared scheduler state.
---

# Change the gocron scheduler

Assume scheduler bugs can duplicate, lose, or overlap production jobs. Make the
execution semantics explicit before editing.

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
state touched, failure cases tested, and any behavior that remains
best-effort.
