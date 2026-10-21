## How It Works

This sample demonstrates how to use **search attributes** — indexed metadata attached to a workflow that enables rich visibility queries against Cadence's ElasticSearch-backed visibility store.

> **Prerequisites**: Custom search attributes (e.g. `CustomIntField`, `CustomKeywordField`) require ElasticSearch to be configured in the Cadence server. Use the ES-enabled docker-compose profile:
> ```bash
> docker compose -f docker/docker-compose-es.yml up
> ```

```
┌───────────────────────────────────────┐
│        searchAttributesWorkflow        │
│                                       │
│  1. Read CustomIntField=1 (from start) │
│  2. UpsertSearchAttributes(...)        │  ← 6 attributes updated
│  3. UpsertSearchAttributes(...)        │  ← 1 attribute updated
│  4. Sleep 2s (wait for ES indexing)    │
│  5. listExecutions activity            │  ← queries ES visibility
└───────────────────────────────────────┘
```

Key concepts:
- **SearchAttributes on start**: Pass initial values via `--search_attr_key` / `--search_attr_value` CLI flags or `StartWorkflowOptions.SearchAttributes`
- **UpsertSearchAttributes**: Updates indexed fields mid-workflow without restarting it
- **ListWorkflow**: Queries the visibility store using SQL-like expressions (requires ES)

## Running the Sample

Start the worker:
```bash
go run .
```

Start the workflow with an initial search attribute:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.searchAttributesWorkflow \
  --tl cadence-samples-worker \
  --et 300 \
  --search_attr_key CustomIntField \
  --search_attr_value 1
```

While the workflow is sleeping, query the visibility store:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow list \
  --query "CustomIntField=2 and CustomKeywordField='Update2'"
```
