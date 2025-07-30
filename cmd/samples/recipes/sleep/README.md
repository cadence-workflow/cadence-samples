# Sleep Workflow Sample

This sample workflow demonstrates how to use the `workflow.Sleep` function in Cadence workflows. The sleep functionality allows workflows to pause execution for a specified duration before continuing with subsequent activities.

## Sample Description

The sample workflow:
- Takes a sleep duration as input parameter
- Uses `workflow.Sleep` to pause workflow execution for the specified duration
- Executes a main activity after the sleep period completes
- Demonstrates proper error handling for sleep operations
- Shows how to configure activity options for post-sleep activities

The workflow is useful for scenarios where you need to:
- Implement delays or timeouts in workflow logic
- Wait for external events or conditions
- Implement retry mechanisms with exponential backoff
- Create scheduled or periodic workflows

## Key Components

- **Workflow**: `sleepWorkflow` demonstrates the sleep functionality with activity execution
- **Activity**: `mainSleepActivity` is executed after the sleep period
- **Sleep Duration**: Configurable duration (default: 30 seconds) passed as workflow input
- **Test**: Includes unit tests to verify sleep and activity execution

## Steps to Run Sample

1. You need a cadence service running. See details in cmd/samples/README.md

2. Run the following command to start the worker:
   ```
   ./bin/sleep -m worker
   ```

3. Run the following command to execute the workflow:
   ```
   ./bin/sleep -m trigger
   ```

You should see logs showing:
- Workflow start with sleep duration
- Sleep completion message
- Main activity execution
- Workflow completion

## Customization

To modify the sleep behavior:
- Change the `sleepDuration` in `main.go` to adjust the default sleep time
- Modify the activity options to configure timeouts for post-sleep activities
- Add additional activities or logic after the sleep period
- Implement conditional sleep based on workflow state

## Use Cases

This pattern is useful for:
- **Scheduled Tasks**: Implement workflows that need to wait before processing
- **Rate Limiting**: Add delays between API calls or external service interactions
- **Retry Logic**: Implement exponential backoff for failed operations
- **Event-Driven Workflows**: Wait for specific time periods before checking conditions
- **Batch Processing**: Add delays between batch operations to avoid overwhelming systems 