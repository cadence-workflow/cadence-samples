# Versioning Workflow Example

This example demonstrates how to safely deploy versioned workflows using Cadence's versioning APIs. It shows how to handle workflow evolution while maintaining backward compatibility and enabling safe rollbacks.

## Overview

The versioning sample implements a workflow that evolves through multiple versions (V1 → V2 → V3 → V4) with rollbacks, demonstrating:

- **Safe Deployment**: How to deploy new workflow versions without breaking existing executions
- **Backward Compatibility**: How to handle workflows started with older versions
- **Rollback Capability**: How to safely rollback to previous versions
- **Version Isolation**: How different versions can execute different logic paths

## Workflow Versions

### Version 1 (V1)
- Executes `FooActivity` only
- Uses `workflow.DefaultVersion` for the change ID

### Version 2 (V2) 
- Supports both `FooActivity` and `BarActivity`
- Uses `workflow.GetVersion()` with `workflow.ExecuteWithMinVersion()` to handle both old and new workflows
- Workflows started by V1 continue using `FooActivity`

### Version 3 (V3)
- Similar to V2 but uses standard `workflow.GetVersion()` (without `ExecuteWithMinVersion`)
- All new workflows use version 1 of the change ID

### Version 4 (V4)
- Only supports `BarActivity`
- Forces all workflows to use version 1 of the change ID
- **Breaking change**: Cannot execute workflows started by V1

## Key Cadence APIs Used

- `workflow.GetVersion()`: Determines which version of code to execute
- `workflow.ExecuteWithVersion()`: Executes code with a specific version
- `workflow.ExecuteWithMinVersion()`: Executes code with minimum version requirement
- `workflow.DefaultVersion`: Represents the original version before any changes

## Safe Deployment Flow

This example demonstrates a safe deployment strategy that allows you to:

1. **Deploy new versions** while keeping old workers running
2. **Test compatibility** before fully switching over
3. **Rollback safely** if issues are discovered
4. **Gradually migrate** workflows to new versions


## Important Notes

- **Single Workflow Limitation**: This sample allows only one workflow at a time to simplify the signal handling mechanism. In production, you would typically handle multiple workflows.
- **Signal Method**: The workflow uses a simple signal method to stop gracefully, keeping the implementation straightforward.
- **Breaking Changes**: V4 demonstrates what happens when you introduce a breaking change - workflows started by V1 cannot be executed.

## Version Compatibility Matrix

| Started By | V1 Worker | V2 Worker | V3 Worker | V4 Worker |
|------------|-----------|-----------|-----------|-----------|
| V1         | ✅        | ✅        | ✅        | ❌        |
| V2         | ❌        | ✅        | ✅        | ✅        |
| V3         | ❌        | ✅        | ✅        | ✅        |
| V4         | ❌        | ✅        | ✅        | ✅        |

## Running the Example

### Prerequisites

Make sure you have Cadence server running and the sample compiled:

```bash
# Build the sample
go build -o bin/versioning cmd/samples/recipes/versioning/*.go
```

### Step-by-Step Deployment Simulation

#### 1. Start Worker V1
```bash
./bin/versioning -m worker -v 1
```

#### 2. Trigger a Workflow
```bash
./bin/versioning -m trigger
```

Wait for logs in the V1 worker to ensure that a workflow has been executed by worker V1.

#### 3. Deploy Worker V2
Let's simulate a deployment from V1 to V2 and run a V2 worker alongside the V1 worker:

```bash
./bin/versioning -m worker -v 2
```

The workflow should still be executed by worker V1.

#### 4. Test V2 Compatibility
Let's simulate that worker V1 is shut down and the workflow will be rescheduled to the V2 worker:
* Kill the process of worker V1 (Ctrl+C), then wait 5 seconds to see workflow rescheduling to worker V2 without errors.

Verify logs of the V2 worker - it should handle the workflow started by V1.

#### 5. Upgrade to Version V3
Let's continue the deployment and upgrade to V3, running a V3 worker alongside the V2 worker:

```bash
./bin/versioning -m worker -v 3
```

The workflow should still be executed by worker V2.

#### 6. Test V3 Compatibility
Let's simulate that worker V2 is shut down and the workflow will be rescheduled to the V3 worker:

* Kill the process of worker V2, then wait 5 seconds to see workflow rescheduling to worker V3 without errors.

Verify logs of the V3 worker - it should handle the workflow started by V2.

#### 7. Gracefully Stop the Workflow
Before upgrading to V4, we should ensure that the workflow has been stopped, otherwise it will fail. For this, we need to send a signal to stop it gracefully:

```bash
./bin/versioning -m stop
```

You should see that the workflow has been stopped.

#### 8. Start a New Workflow
Let's start a new workflow:

```bash
./bin/versioning -m trigger
```

The workflow will use version 1 of the change ID (V3's and V4's default).

#### 9. Rollback to Worker V2
Let's imagine that V3 has an issue and we need to rollback to V2. Let's start a worker V2:

```bash
./bin/versioning -m worker -v 2
```

* Kill the process of worker V3, then wait for workflow rescheduling.
* Verify logs of V2 worker - V2 worker should handle workflows started by V3.

#### 10. Aggressive Upgrade: V2 to V4 (Breaking Change)
We decide to combine getting rid of support for V1 and make an upgrade straightforward to V4:

```bash
./bin/versioning -m worker -v 4
```

* Kill the process of worker V2, then wait for workflow rescheduling.
* Verify logs of V4 worker - V4 worker should handle workflows started by V4.


## Command Reference

```bash
# Start a worker with specific version
./bin/versioning -m worker -v <version>

# Start a new workflow
./bin/versioning -m trigger

# Stop the running workflow
./bin/versioning -m stop
```

Where `<version>` can be:
- `1` or `v1` - Version 1 (FooActivity only, DefaultVersion)
- `2` or `v2` - Version 2 (FooActivity + BarActivity, DefaultVersion)
- `3` or `v3` - Version 3 (FooActivity + BarActivity, Version #1)
- `4` or `v4` - Version 4 (BarActivity only, Version #1)
