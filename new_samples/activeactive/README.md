# Active-Active Sample

This sample shows Cadence **active-active domains**. Workflows can belong to a **cluster attribute**, and
each is active in one cluster at a time. This allows for logically grouping workflows
and keeping them active together.


The sample uses three regions as cluster attributes, spread over two clusters:

| Cluster attribute | Active cluster at start |
|-------------------|-------------------------|
| `region.us-west`    | `cluster0` |
| `region.us-central` | `cluster0` |
| `region.us-east`    | `cluster1` |

You'll fail over `us-central` from `cluster0` to `cluster1` and back, while `cluster0` stays active
for `us-west` the whole time.

The sample runs:

* Two Cadence clusters replicating to each other, plus Cadence Web:

  | Cluster    | Frontend         |
  |------------|------------------|
  | `cluster0` | `127.0.0.1:7833` |
  | `cluster1` | `127.0.0.1:7843` |

* One worker per cluster. Both poll the same domain and task list, but each only talks to its own
  cluster, and a cluster only dispatches tasks for workflows whose cluster attribute is active in it.
* `ActiveActiveWorkflow`, which runs `GetClusterActivity` once when it starts. `GetClusterActivity`
  returns the cluster of the worker that ran it. The workflow then waits for signals in a loop:
  * `run_activity` runs `GetClusterActivity` again.
  * `complete` completes the workflow and returns every cluster the activity ran in.

The activity results in the workflow history show where the workflow ran before and after each failover.

## Prerequisites

* Docker with Docker Compose.
* Go.
* Cadence CLI v1.4.0 or later, which supports active-active domains: `brew install cadence-workflow`.
  See [the CLI docs](https://cadenceworkflow.io/docs/cli/) for other ways to install it.

Set the domain and transport for the CLI once in your shell. Each command still passes `--address`
to pick the cluster it talks to.

```bash
export CADENCE_CLI_DOMAIN=cadence-samples-active-active
export CADENCE_CLI_TRANSPORT_PROTOCOL=grpc
```

## Steps to run the sample

Run all commands from this folder.

### 1. Start the clusters

```bash
docker compose up -d
```

This starts both Cadence clusters and Cadence Web at [localhost:8088](http://localhost:8088).
The clusters use [dynamicconfig/active_active.yaml](dynamicconfig/active_active.yaml).

### 2. Register the active-active domain

Register the domain on `cluster0`, mapping each region to the cluster that is active for it:

```bash
cadence --address 127.0.0.1:7833 domain register \
  --clusters cluster0,cluster1 \
  --active_cluster cluster0 \
  --active_clusters 'region.us-west:cluster0,region.us-central:cluster0,region.us-east:cluster1'
```

`--active_clusters` uses the format `<scope>.<attribute>:<cluster>`. The domain replicates to
`cluster1`. Check it there:

```bash
cadence --address 127.0.0.1:7843 domain describe --format json
```

### 3. Start a worker for each cluster

In two separate terminals:

```bash
go run . worker --cluster cluster0
```

```bash
go run . worker --cluster cluster1
```

### 4. Start a workflow in each region

```bash
go run . start --region us-west --wid demo-us-west
go run . start --region us-central --wid demo-us-central
go run . start --region us-east --wid demo-us-east
```

`start` sets `ActiveClusterSelectionPolicy` on the start request, so each workflow belongs to its
region's cluster attribute:

```go
client.StartWorkflowOptions{
	// ...
	ActiveClusterSelectionPolicy: &client.ActiveClusterSelectionPolicy{
		ClusterAttribute: &client.ClusterAttribute{Scope: "region", Name: "us-central"},
	},
}
```

All three requests go to `cluster0`, which forwards `demo-us-east` to `cluster1`.

Each workflow runs the activity right away: the `cluster0` worker runs it for `us-west` and
`us-central`, and the `cluster1` worker runs it for `us-east`.

You can also start a workflow with the CLI:

```bash
cadence --address 127.0.0.1:7833 workflow start \
  --tl cadence-samples-worker \
  --wt cadence_samples.ActiveActiveWorkflow \
  --et 3600 \
  --wid demo-us-central \
  --cluster_attribute_scope region \
  --cluster_attribute_name us-central
```

### 5. Fail over us-central to cluster1

Fail over only the `us-central` cluster attribute. Send failover requests to the **destination**
cluster:

```bash
cadence --address 127.0.0.1:7843 domain failover --active_clusters 'region.us-central:cluster1'
```

Wait about 15 seconds, then signal all three workflows to run the activity again:

```bash
for region in us-west us-central us-east; do
  cadence --address 127.0.0.1:7833 workflow signal --wid demo-$region --name run_activity
done
```

This time:

* `us-west` still runs in `cluster0`, which stays active for it.
* `us-central` now runs in `cluster1`.
* `us-east` still runs in `cluster1`.

Why wait? See [Why the waits?](#why-the-waits).

### 6. Fail back and complete

```bash
cadence --address 127.0.0.1:7833 domain failover --active_clusters 'region.us-central:cluster0'
```

Wait about 15 seconds, signal `run_activity` once more (step 5), then complete the workflows:

```bash
for region in us-west us-central us-east; do
  cadence --address 127.0.0.1:7833 workflow signal --wid demo-$region --name complete
done
```

### 7. Check the workflow histories

```bash
cadence --address 127.0.0.1:7833 workflow show --wid demo-us-central
```

The `ActivityTaskCompleted` events show where each activity ran:

```
   7  ActivityTaskCompleted       {Result:["cluster0"], ...
  17  ActivityTaskCompleted       {Result:["cluster1"], ...
  27  ActivityTaskCompleted       {Result:["cluster0"], ...
  35  WorkflowExecutionCompleted  {Result:[["cluster0","cluster1","cluster0"]], ...
```

The other two workflows complete with `["cluster0","cluster0","cluster0"]` (`us-west`) and
`["cluster1","cluster1","cluster1"]` (`us-east`). Only the region you failed over moved.

You can see the same histories in Cadence Web at [localhost:8088](http://localhost:8088), under the
`cadence-samples-active-active` domain. The domain's **Failovers** tab lists both `us-central`
failovers, with the cluster each one moved from and to. Its metadata shows the DescribeDomain
response, with the active cluster for each region under `activeClusters`.

### Clean up

Stop the workers, then:

```bash
docker compose down
```

## Why the waits?

In this sample, signaling immediately without waiting could result in the activity still running in the old region. This is by design.

Cadence replication is eventually consistent. It favors throughput and availability over having every cluster
agree at every instant: clusters replicate domain changes and workflow histories to each other
asynchronously. Inconsistencies are resolved through deterministic conflict resolution.

**Failovers take a moment to apply.** A failover updates the domain immediately, but each host's cache needs to refresh.
Until then, the old cluster can still treat `us-central` as active, so an immediate signal may run the activity in the
old cluster. Waiting about 15 seconds covers the refresh.

**In-flight activities finish.** A failover doesn't cancel work that has already been handed out. An
activity that a worker picked up before the failover is allowed to finish. If you signal during the
failover window, you may see one last activity complete in the old cluster.

Cadence guarantees that Workflows continue to make progress and failovers introduce minimal delays. 

## References

* [Cadence documentation](https://cadenceworkflow.io/docs)
* [Cadence Replication](https://cadenceworkflow.io/docs/concepts/cross-dc-replication/)
* [Cadence Go client](https://github.com/cadence-workflow/cadence-go-client)
* [Cadence server](https://github.com/cadence-workflow/cadence)
