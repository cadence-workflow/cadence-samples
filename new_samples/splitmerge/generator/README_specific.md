## How It Works

This sample demonstrates the **split-merge** pattern using Cadence coroutines:

```
                    ┌─────────────────┐
                    │SplitMergeWorkflow│
                    │  (workerCount=3) │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        workflow.Go    workflow.Go    workflow.Go
              │              │              │
              ▼              ▼              ▼
        ┌─────────┐    ┌─────────┐    ┌─────────┐
        │ Chunk 1 │    │ Chunk 2 │    │ Chunk 3 │
        │Activity │    │Activity │    │Activity │
        └────┬────┘    └────┬────┘    └────┬────┘
              │              │              │
              └──────────────┼──────────────┘
                             ▼
                    ┌─────────────────┐
                    │  Merge Results  │
                    │ (totalSum, etc) │
                    └─────────────────┘
```

Key concepts:
- **workflow.Go**: Launch concurrent coroutines (NOT goroutines)
- **workflow.NewChannel**: Create channels for coroutine communication
- Results are collected and merged after all chunks complete

## Running the Sample

Start the worker:
```bash
go run .
```

Trigger the workflow with 3 parallel workers:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.SplitMergeWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '3'
```

