# Temporal Conventions

> **When writing or reviewing any Temporal workflow/activity code, read this file first.**
>
> These guidelines cover best practices for building reliable, maintainable workflows with [Temporal](https://docs.temporal.io/) in Go.

---

## Table of Contents

- [Workflow Design — Determinism](#workflow-design--determinism)
- [Activity Design](#activity-design)
- [Worker Registration](#worker-registration)
- [Error Handling](#error-handling)
- [Workflow Versioning](#workflow-versioning)
- [Signals, Queries, and Updates](#signals-queries-and-updates)
- [Retries and Timeouts](#retries-and-timeouts)
- [Activity Retry Patterns](#activity-retry-patterns)
- [ContinueAsNew](#continueasnew)
- [Testing](#testing)
- [Naming Conventions](#naming-conventions)
- [Code Quality Checklist](#code-quality-checklist)

---

## Workflow Design — Determinism

Temporal replays workflow history to reconstruct state. Any non-deterministic operation in a workflow function will cause a non-deterministic error during replay.

**Rules:**
- Never call `time.Now()` or `time.Sleep()` directly — use `workflow.Now(ctx)` and `workflow.Sleep(ctx, d)`.
- Never use `math/rand` directly — use `workflow.NewRandom(ctx)` for random values.
- Never read from environment variables, files, or external state inside a workflow function.
- Never spawn raw goroutines (`go func() {...}`) — use `workflow.Go(ctx, func(ctx workflow.Context) {...})`.
- Never use `context.Context` as workflow parameters — use `workflow.Context` instead.
- Never use `select` with real channels — use `workflow.NewSelector(ctx)` instead.
- All side-effecting or non-deterministic logic must be pushed into Activities.

```go
// Bad — not deterministic
func MyWorkflow(ctx workflow.Context) error {
    t := time.Now()         // non-deterministic
    _ = os.Getenv("KEY")    // non-deterministic
    return nil
}

// Good
func MyWorkflow(ctx workflow.Context) error {
    t := workflow.Now(ctx)  // deterministic replay-safe clock
    _ = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
        return os.Getenv("KEY") // side effect captured once
    })
    return nil
}
```

---

## Activity Design

Activities are the unit of work that performs all I/O and non-deterministic operations. They must be designed to be safe to retry.

- **Idempotency**: every Activity must be safe to run more than once with the same inputs. Use idempotency keys when calling external APIs. For non-idempotent operations (e.g. charging a card), set `MaximumAttempts: 1` and apply exactly-once execution patterns.
- **Heartbeating**: long-running Activities must call `activity.RecordHeartbeat(ctx, details...)` periodically so the worker can detect failures and resume from the last checkpoint.
- **Context propagation**: always accept `context.Context` as the first parameter so that Temporal can cancel the Activity on timeout.
- **Return meaningful errors**: return non-retryable errors with `temporal.NewNonRetryableApplicationError` for permanent failures; return retryable errors (or plain `error`) for transient failures.

```go
func ProcessOrderActivity(ctx context.Context, orderID string) error {
    // Heartbeat for long-running work
    activity.RecordHeartbeat(ctx, "starting")

    if err := externalClient.Process(ctx, orderID); err != nil {
        if isPermanentFailure(err) {
            // Temporal will NOT retry this
            return temporal.NewNonRetryableApplicationError(
                "permanent failure processing order",
                "PermanentProcessError",
                err,
            )
        }
        return err // Temporal will retry
    }
    return nil
}
```

---

## Worker Registration

- Register all Workflows and Activities in the composition root / `main()` — never inside domain packages.
- Group registrations by task queue; each `worker.Worker` instance must correspond to exactly one task queue.
- Verify that every Workflow and Activity used in production is registered on at least one worker.

```go
w := worker.New(temporalClient, "order-processing", worker.Options{})
w.RegisterWorkflow(OrderWorkflow)
w.RegisterActivity(ProcessOrderActivity)
w.RegisterActivity(NotifyCustomerActivity)
```

---

## Error Handling

Use Temporal's typed error hierarchy to distinguish between retryable and non-retryable failures:

| Error type | Retried by default | Use when |
|---|---|---|
| `temporal.ApplicationError` (retryable) | Yes | Transient failures; caller does not need to match the type |
| `temporal.ApplicationError` (non-retryable) | No | Permanent / business-rule failures |
| `temporal.CanceledError` | No | Activity/workflow was cancelled |
| `temporal.TimeoutError` | No | Execution exceeded a deadline |

Unwrap Temporal errors with `errors.As` to inspect the type at the calling layer:

```go
var appErr *temporal.ApplicationError
if errors.As(err, &appErr) {
    log.Error("activity failed", zap.String("type", appErr.Type()), zap.Error(appErr))
}
```

---

## Workflow Versioning

When changing workflow logic that may affect in-flight executions, use `workflow.GetVersion` to maintain backward compatibility. Never remove a `GetVersion` check until all workflows that used the old branch have completed.

```go
const (
    _featureAddNotification = "add-notification"
    _defaultVersion         = workflow.DefaultVersion
    _versionOne             = 1
)

func OrderWorkflow(ctx workflow.Context, order Order) error {
    v := workflow.GetVersion(ctx, _featureAddNotification, _defaultVersion, _versionOne)

    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)

    if err := workflow.ExecuteActivity(ctx, ProcessOrderActivity, order.ID).Get(ctx, nil); err != nil {
        return fmt.Errorf("process order: %w", err)
    }

    if v >= _versionOne {
        if err := workflow.ExecuteActivity(ctx, NotifyCustomerActivity, order.ID).Get(ctx, nil); err != nil {
            return fmt.Errorf("notify customer: %w", err)
        }
    }
    return nil
}
```

---

## Signals, Queries, and Updates

- **Signals** are fire-and-forget messages delivered to a running workflow. Handle them with `workflow.GetSignalChannel(ctx, name).Receive(ctx, &val)` or via a `workflow.NewSelector`.
- **Queries** must be read-only and must not modify workflow state. Register with `workflow.SetQueryHandler`.
- **Updates** (Temporal ≥ 1.21) combine signal delivery and synchronous validation. Use `workflow.SetUpdateHandlerWithOptions` and provide a validator function to reject invalid inputs before they are written to history.
- Name signals, queries, and updates as lowercase kebab-case strings (e.g. `"cancel-order"`, `"get-status"`).

```go
const _signalCancelOrder = "cancel-order"

func OrderWorkflow(ctx workflow.Context, order Order) error {
    cancelCh := workflow.GetSignalChannel(ctx, _signalCancelOrder)

    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)

    sel := workflow.NewSelector(ctx)
    var processErr error
    actFuture := workflow.ExecuteActivity(ctx, ProcessOrderActivity, order.ID)
    sel.AddFuture(actFuture, func(f workflow.Future) {
        processErr = f.Get(ctx, nil)
    })
    sel.AddReceive(cancelCh, func(c workflow.ReceiveChannel, _ bool) {
        var reason string
        c.Receive(ctx, &reason)
        processErr = temporal.NewCanceledError(reason)
    })
    sel.Select(ctx)

    return processErr
}
```

---

## Retries and Timeouts

Always configure timeouts and retry policies explicitly — never rely on Temporal's unlimited defaults in production.

**Timeout options:**

| Option | Purpose | Guideline |
|---|---|---|
| `StartToCloseTimeout` | Max time for a single Activity attempt | Required; must be greater than the upstream service's own request timeout |
| `ScheduleToCloseTimeout` | Max total time including all retries | Must accommodate the worst-case cumulative retry duration; set for Activities with SLA requirements |
| `HeartbeatTimeout` | Max time between heartbeats | Required for long-running Activities |
| `WorkflowExecutionTimeout` | Max lifetime of an entire workflow | Set to prevent unbounded open workflows |
| `WorkflowRunTimeout` | Max lifetime of a single workflow run | Useful for workflows that use `ContinueAsNew` |

**Timeout rules:**
- `StartToCloseTimeout` must be greater than the upstream service's own request timeout.
- `ScheduleToCloseTimeout` must cover the worst-case cumulative retry scenario.

**Exponential backoff strategy:**
- Use a `BackoffCoefficient` between **1.5 and 2.0**.
- Cap `MaximumInterval` based on the nature of the upstream:
  - HTTP / lightweight internal APIs: **10–30 seconds**
  - Slow or unstable third-party services: **30–60 seconds**

**Retry limits:**
- Avoid unlimited retries (`MaximumAttempts: 0`) for critical external calls.
- `MaximumAttempts` **includes the initial attempt**: `1` → no retries, `0` → unlimited.

```go
// Fast upstream (e.g. internal HTTP service)
fastAO := workflow.ActivityOptions{
    StartToCloseTimeout:    10 * time.Second,
    ScheduleToCloseTimeout: 2 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:        time.Second,
        BackoffCoefficient:     2.0,
        MaximumInterval:        30 * time.Second,
        MaximumAttempts:        5,
        NonRetryableErrorTypes: []string{"PermanentProcessError"},
    },
}

// Slow / unstable upstream (e.g. third-party payment provider)
slowAO := workflow.ActivityOptions{
    StartToCloseTimeout:    60 * time.Second,
    ScheduleToCloseTimeout: 10 * time.Minute,
    HeartbeatTimeout:       15 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:        3 * time.Second,
        BackoffCoefficient:     1.5,
        MaximumInterval:        60 * time.Second,
        MaximumAttempts:        4,
        NonRetryableErrorTypes: []string{"PermanentProcessError"},
    },
}

// Single-shot operation (e.g. charge a payment — must not retry)
singleShotAO := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 1,
    },
}
```

---

## Activity Retry Patterns

Choose a retry pattern based on the nature of the operation and the acceptable failure mode.

| Pattern | When to use | Key config |
|---|---|---|
| **Standard retry** | Transient failures on idempotent calls (HTTP timeouts, rate limits, temporary unavailability) | `MaximumAttempts > 1`, exponential backoff |
| **Single-shot (no retry)** | Non-idempotent operations where a duplicate execution causes harm (e.g. charge a card, send an SMS) | `MaximumAttempts: 1`, strong alerting |
| **Retry with compensation** | Retries are exhausted and a partial side effect must be rolled back (saga pattern) | `MaximumAttempts > 1` + a compensation Activity in the failure path |
| **Heartbeat-based resume** | Long-running operations that can checkpoint progress and resume from the last known state | `HeartbeatTimeout` set, checkpoint stored in heartbeat details |
| **Manual retry via signal** | Human approval or intervention is required before a retry | `MaximumAttempts: 1`, workflow waits for a signal before re-executing |
| **Polling with delay** | External resource is not yet ready and must be re-checked after a fixed wait | `workflow.Sleep` between attempts inside the workflow loop |
| **Custom retry logic in Activity** | Built-in retry policy is too coarse; different error types need different back-off | Manual loop with typed error inspection inside the Activity |

> **Tip:** See [temporalio/samples-go](https://github.com/temporalio/samples-go) for runnable examples of these and other patterns in Go.

### Standard Retry

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout:    10 * time.Second,
    ScheduleToCloseTimeout: 2 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:        time.Second,
        BackoffCoefficient:     2.0,
        MaximumInterval:        30 * time.Second,
        MaximumAttempts:        5,
        NonRetryableErrorTypes: []string{"ItemNotFound"},
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
if err := workflow.ExecuteActivity(ctx, ReserveInventoryActivity, orderID).Get(ctx, nil); err != nil {
    return fmt.Errorf("reserve inventory: %w", err)
}
```

### Single-Shot (No Retry)

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 1, // no retries
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
if err := workflow.ExecuteActivity(ctx, ChargePaymentActivity, order).Get(ctx, nil); err != nil {
    return fmt.Errorf("charge payment: %w", err)
}
```

### Retry with Compensation (Saga)

```go
reserveCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: 10 * time.Second,
    RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
})
if err := workflow.ExecuteActivity(reserveCtx, ReserveInventoryActivity, orderID).Get(ctx, nil); err != nil {
    return fmt.Errorf("reserve inventory: %w", err)
}

chargeCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
})
if err := workflow.ExecuteActivity(chargeCtx, ChargePaymentActivity, order).Get(ctx, nil); err != nil {
    // Compensate: release the reservation that was already committed.
    compensateCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
        RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
    })
    _ = workflow.ExecuteActivity(compensateCtx, ReleaseInventoryActivity, orderID).Get(ctx, nil)
    return fmt.Errorf("charge payment: %w", err)
}
```

### Heartbeat-Based Resume

```go
func ProcessFileActivity(ctx context.Context, fileID string) error {
    startPage := 0
    if details := activity.GetInfo(ctx).HeartbeatDetails; details != nil {
        _ = converter.GetDefaultDataConverter().FromPayloads(details, &startPage)
    }

    for page := startPage; ; page++ {
        if err := ctx.Err(); err != nil {
            return err
        }
        done, err := processPage(ctx, fileID, page)
        if err != nil {
            return err
        }
        activity.RecordHeartbeat(ctx, page+1)
        if done {
            return nil
        }
    }
}
```

### Polling with Delay

```go
func ExportWorkflow(ctx workflow.Context, jobID string) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
        RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    const (
        _maxPolls     = 20
        _pollInterval = 15 * time.Second
    )
    for i := 0; i < _maxPolls; i++ {
        var done bool
        if err := workflow.ExecuteActivity(ctx, CheckExportStatusActivity, jobID).Get(ctx, &done); err != nil {
            return fmt.Errorf("check export status: %w", err)
        }
        if done {
            return nil
        }
        if err := workflow.Sleep(ctx, _pollInterval); err != nil {
            return err
        }
    }
    return temporal.NewNonRetryableApplicationError("export job timed out", "ExportTimeout", nil)
}
```

---

## ContinueAsNew

Use `workflow.NewContinueAsNewError` to restart a workflow with fresh history when it processes unbounded event streams or long-running loops. This prevents history size from growing indefinitely.

```go
const _maxIterations = 1000

func PollingWorkflow(ctx workflow.Context, state State) error {
    for i := 0; i < _maxIterations; i++ {
        // ... process one iteration ...
    }
    // Hand off to a new run with updated state
    return workflow.NewContinueAsNewError(ctx, PollingWorkflow, state)
}
```

---

## Testing

Use `testsuite.WorkflowTestSuite` for unit-testing workflows and activities in a deterministic, in-process environment. Do not call a real Temporal server in unit tests.

- Mock Activities with `env.OnActivity(...)` to control return values and assert call counts.
- Use `env.RegisterDelayedCallback` to simulate signals and timers.
- Test Activities independently with plain Go unit tests — they are ordinary functions.

```go
func TestOrderWorkflow_Success(t *testing.T) {
    suite.Run(t, new(OrderWorkflowTestSuite))
}

type OrderWorkflowTestSuite struct {
    suite.Suite
    testsuite.WorkflowTestSuite
}

func (s *OrderWorkflowTestSuite) Test_ProcessOrder() {
    env := s.NewTestWorkflowEnvironment()
    env.OnActivity(ProcessOrderActivity, mock.Anything, "order-123").Return(nil)
    env.OnActivity(NotifyCustomerActivity, mock.Anything, "order-123").Return(nil)

    env.ExecuteWorkflow(OrderWorkflow, Order{ID: "order-123"})

    s.True(env.IsWorkflowCompleted())
    s.NoError(env.GetWorkflowError())
    env.AssertExpectations(s.T())
}
```

---

## Naming Conventions

| Artifact | Convention | Example |
|---|---|---|
| Workflow function | `MixedCaps` + `Workflow` suffix | `OrderWorkflow` |
| Activity function | `MixedCaps` + `Activity` suffix | `ProcessOrderActivity` |
| Task queue name | kebab-case string constant | `"order-processing"` |
| Signal / query / update name | kebab-case string constant | `"cancel-order"` |
| Workflow ID | Unique, human-readable, business-scoped | `"order-" + orderID` |

---

## Code Quality Checklist

Before requesting a review on Temporal-related changes, verify:

- [ ] Workflow functions contain no non-deterministic code (no `time.Now`, no I/O, no raw goroutines).
- [ ] Every Activity is idempotent and accepts `context.Context` as the first parameter.
- [ ] Long-running Activities call `activity.RecordHeartbeat` with a `HeartbeatTimeout` set.
- [ ] `StartToCloseTimeout` is set on every `ActivityOptions` block and is greater than the upstream service timeout.
- [ ] `ScheduleToCloseTimeout` covers the worst-case cumulative retry duration.
- [ ] `MaximumAttempts` is explicitly set; unlimited retries (`0`) are avoided for critical external calls.
- [ ] Non-idempotent Activities use `MaximumAttempts: 1` with alerting, or an exactly-once deduplication strategy.
- [ ] Non-retryable errors use `temporal.NewNonRetryableApplicationError`.
- [ ] Workflow logic changes use `workflow.GetVersion` to protect in-flight executions.
- [ ] Workflows that process unbounded events use `workflow.NewContinueAsNewError`.
- [ ] Worker registration happens in the composition root, not inside domain packages.
- [ ] Workflow and Activity unit tests use `testsuite.WorkflowTestSuite` and mock all Activities.
