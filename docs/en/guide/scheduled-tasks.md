# Scheduled Tasks

## Crontab Time Expression

::: tip Time Expression Format
gocron uses Linux-crontab time expression syntax and supports second-level task definition.

**Format**: `second minute hour day month week`
:::

### Expression Examples

```bash
1 * * * * *           # Run at the first second of every minute
*/20 * * * * *        # Run every 20 seconds
0 30 21 * * *         # Run once every day at 21:30:00
0 0 23 * * 6          # Run once every Saturday at 23:00:00
```

### Shortcut Syntax

gocron supports the following shortcut syntax:

| Shortcut | Description | Equivalent Expression |
|----------|-------------|----------------------|
| `@yearly` | Run once a year | `0 0 0 1 1 *` |
| `@monthly` | Run once a month | `0 0 0 1 * *` |
| `@weekly` | Run once a week | `0 0 0 * * 0` |
| `@daily` | Run once a day | `0 0 0 * * *` |
| `@midnight` | Run once at midnight every day | `0 0 0 * * *` |
| `@hourly` | Run once an hour | `0 0 * * * *` |
| `@every 30s` | Run every 30 seconds | - |
| `@every 1m20s` | Run every 1 minute and 20 seconds | - |
| `@every 3h5m10s` | Run every 3 hours, 5 minutes and 10 seconds | - |

## Execution Methods

gocron supports two task execution methods:

### 1. Shell Command

Execute shell commands on remote host.

**Examples**:
```bash
ps aux | grep gocron
cd /tmp && ls -la
echo "Hello, gocron!" > /tmp/test.txt
```

::: warning Note
Shell task execution time cannot exceed 86400 seconds (24 hours)
:::

### 2. HTTP Request

Execute HTTP-GET request.

**Examples**:
```
https://api.example.com/task/run
http://localhost:8080/webhook
```

::: warning Note
HTTP task execution time cannot exceed 300 seconds (5 minutes)
:::

## Task Configuration

### Task Timeout

Tasks will be forcefully terminated after timeout. Default is 0 (no limit).

- **Shell Tasks**: Maximum execution time cannot exceed 86400 seconds
- **HTTP Tasks**: Maximum execution time cannot exceed 300 seconds

### Task Retry on Failure

Tasks are marked as failed in the following situations:
- Unable to connect to remote host
- Shell command returns non-zero value
- HTTP response code is not 200

**Retry Mechanism**:
- Retry count can be set (range: 1-10)
- Retry interval = retry count × 1 minute
- Retries at intervals of 1 minute, 2 minutes, 3 minutes...

**Example**: Retry count set to `2`
1. Task execution failed
2. Sleep 1 minute, first retry
3. If still failed, sleep 2 minutes, second retry
4. If still failed, task is finally marked as failed

::: tip Tip
Default value is 0, which means no retry
:::

### Single Instance Run

When single instance run is enabled, if the previous task has not been completed, the next task will not be executed, and the task log status will show "canceled".

This feature prevents duplicate task execution, especially suitable for long-running tasks.

## Task Dependencies

gocron models an execution chain with master and child tasks:

1. Create each downstream task with its type set to **Child Task**. Child tasks are not scheduled independently by a Cron expression.
2. Create or edit the **Master Task** and enter the child task IDs in **Child Task IDs**. Separate multiple IDs with commas, for example `12,15`.
3. After the master finishes, gocron triggers the child tasks that still exist concurrently.

The dependency mode controls what happens when the master fails:

- **Strong**: trigger child tasks only after the master succeeds.
- **Weak**: trigger child tasks after the master finishes, whether it succeeds or fails.

::: warning Note
The dependency relationship is stored on the master task. Do not enter the master task ID on a child task.
:::

## Related Documentation

- [Log Management](./log-management)
- [API Documentation](./api)
