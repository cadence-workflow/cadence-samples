### Background 

This sample workflow demos a normal and local activity as they deal with a crashing, timing out, misbehaving activity through an aggressive retry policy.

Steps to run this sample:
1) You need a cadence service running. See details in cmd/samples/README.md
2) Run the following command to start worker
```
./bin/localactivityfailure -m worker
```
3) Run the following command to trigger a local workflow execution. You should see workflowID and runID print out on screen and a final log entry showing the final result.
```
./bin/localactivityfailure -m trigger-local
```

## Test

The test is relatively simple: Run an activity which will fail quite a lot, or timeout, and retry it sufficiently that it completes, eventually. Do this 20 times.

As a simple verification, the integer returned (5 in this instance) run 20 times ought to equal 100 if all the activities eventually pass. More than that would be a failure of correctless and less would indicate the failures caused the activity to not complete sufficiently.