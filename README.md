# Cadence Samples ![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/uber-common/cadence-samples/build.yml)

Welcome to the Cadence Samples repository! This collection demonstrates the powerful capabilities of Cadence workflow orchestration through practical, real-world examples. Whether you're new to Cadence or looking to implement specific patterns, these samples will help you understand how to build reliable, scalable, and maintainable workflow applications.

## What is Cadence?

Cadence is a distributed, scalable, durable, and highly available orchestration engine that helps developers build reliable applications. It provides:

- **Reliability**: Automatic retry mechanisms, error handling, and fault tolerance
- **Scalability**: Distributed execution across multiple workers
- **Durability**: Persistent workflow state that survives failures
- **Observability**: Built-in monitoring, tracing, and querying capabilities

Learn more about Cadence at:
- [Documentation](https://cadenceworkflow.io)
- [Cadence Server](https://github.com/cadence-workflow/cadence)
- [Cadence Go Client](https://github.com/cadence-workflow/cadence-go-client)

## 🚀 Quick Start

### Prerequisites

1. **Clone the Repository**:
```bash
git clone https://github.com/uber-common/cadence-samples.git && cd cadence-samples
```

2. Start Cadence Server*
```bash
curl -LO https://raw.githubusercontent.com/cadence-workflow/cadence/refs/heads/master/docker/docker-compose.yml && docker-compose up --wait
```

This downloads and starts all required dependencies including Cadence server, database, and [Cadence Web UI](https://github.com/uber/cadence-web). You can view your sample workflows at [http://localhost:8088](http://localhost:8088).

3. **Build All Samples**:
```bash
make
```

## 📚 Sample Categories

### 🎯 **Basic Examples**

#### [Hello World](cmd/samples/recipes/helloworld/)
* **Shows**: Basic Cadence workflow concepts and activity execution.
* **What it does**: Executes a single activity that returns a greeting message.
* **Real-world use case**: Foundation for understanding workflow structure, activity execution, and basic error handling.
* **Key concepts**: Workflow definition, activity execution, error handling, worker setup.

##### How to run
Start Worker:
```bash
./bin/helloworld -m worker
```

Start Workflow:
```bash
./bin/helloworld -m trigger
```

#### [Greetings](cmd/samples/recipes/greetings/)
* **Shows**: Sequential activity execution and result passing between activities.
* **What it does**: Executes three activities in sequence: get greeting, get name, then combine them.
* **Real-world use case**: Multi-step processes like user registration, order processing, or data transformation pipelines.
* **Key concepts**: Sequential execution, activity chaining, result passing between activities.

##### How to run
Start Worker:
```bash
./bin/greetings -m worker
```

Start Workflow:
```bash
./bin/greetings -m trigger
```

#### [Cron](cmd/samples/cron/)
* **Shows**: Automated recurring tasks and cron scheduling.
* **What it does**: Executes a workflow based on cron expressions (e.g., every minute, daily at 2 AM).
* **Real-world use case**: Data backups, report generation, system maintenance, periodic data synchronization.
* **Key concepts**: Cron scheduling, workflow persistence, time-based execution.

##### How to run
Start Worker:
```bash
./bin/cron -m worker
```

Start Workflow:
```bash
./bin/cron -m trigger -cron "* * * * *"  # Run every minute
```

#### [Timer](cmd/samples/recipes/timer/)
* **Shows**: Timeout and delay handling with parallel execution.
* **What it does**: Starts a long-running process and sends a notification if it takes too long.
* **Real-world use case**: Order processing with SLA monitoring, payment processing with timeout alerts, API calls with fallback mechanisms.
* **Key concepts**: Timer creation, timeout handling, parallel execution with cancellation.

##### How to run
Start Worker:
```bash
./bin/timer -m worker
```

Start Workflow:
```bash
./bin/timer -m trigger
```

#### [Delay Start](cmd/samples/recipes/delaystart/)
* **Shows**: Deferred execution and delayed workflow execution.
* **What it does**: Waits for a specified duration before executing the main workflow logic.
* **Real-world use case**: Scheduled maintenance windows, delayed notifications, batch processing at specific times.
* **Key concepts**: Delayed execution, time-based workflow scheduling.

##### How to run
Start Worker:
```bash
./bin/delaystart -m worker
```

Start Workflow:
```bash
./bin/delaystart -m trigger
```

### 🔄 **Parallel Execution Examples**

#### [Branch](cmd/samples/recipes/branch/)
* **Shows**: Parallel activity execution and concurrent activity management.
* **What it does**: Executes multiple activities in parallel and waits for all to complete.
* **Real-world use case**: Processing multiple orders simultaneously, calling multiple APIs in parallel, batch data processing.
* **Key concepts**: Parallel execution, Future handling, concurrent activity management.

##### How to run
Start Worker:
```bash
./bin/branch -m worker
```

Start Single Branch Workflow:
```bash
./bin/branch -m trigger -c branch
```

Start Parallel Branch Workflow:
```bash
./bin/branch -m trigger -c parallel
```

#### [Split-Merge](cmd/samples/recipes/splitmerge/)
* **Shows**: Divide and conquer pattern with parallel processing.
* **What it does**: Splits a large task into chunks, processes them in parallel, then merges results.
* **Real-world use case**: Large file processing, batch data analysis, image/video processing, ETL pipelines.
* **Key concepts**: Work splitting, parallel processing, result aggregation, worker coordination.

##### How to run
Start Worker:
```bash
./bin/splitmerge -m worker
```

Start Workflow:
```bash
./bin/splitmerge -m trigger
```

#### [Pick First](cmd/samples/recipes/pickfirst/)
* **Shows**: Race condition handling and activity cancellation.
* **What it does**: Runs multiple activities in parallel and uses the result from whichever completes first.
* **Real-world use case**: Multi-provider API calls, redundant service calls, failover mechanisms, load balancing.
* **Key concepts**: Parallel execution, cancellation, race condition handling.

##### How to run
Start Worker:
```bash
./bin/pickfirst -m worker
```

Start Workflow:
```bash
./bin/pickfirst -m trigger
```

### 🔍 **Advanced Examples**

#### [Choice](cmd/samples/recipes/choice/)
* **Shows**: Conditional execution and decision-based activity routing.
* **What it does**: Executes different activities based on the result of a decision activity.
* **Real-world use case**: Order routing based on type, user authentication flows, approval workflows, conditional processing.
* **Key concepts**: Conditional logic, decision trees, workflow branching.

##### How to run
Start Worker:
```bash
./bin/choice -m worker
```

Start Single Choice Workflow:
```bash
./bin/choice -m trigger -c single
```

Start Multi-Choice Workflow:
```bash
./bin/choice -m trigger -c multi
```

#### [Retry Activity](cmd/samples/recipes/retryactivity/)
* **Shows**: Resilient processing with retry policies and heartbeat tracking.
* **What it does**: Demonstrates activity retry policies with heartbeat progress tracking.
* **Real-world use case**: API calls with intermittent failures, database operations, external service integration.
* **Key concepts**: Retry policies, heartbeat mechanisms, progress tracking, failure recovery.

##### How to run
Start Worker:
```bash
./bin/retryactivity -m worker
```

Start Workflow:
```bash
./bin/retryactivity -m trigger
```

#### [Cancel Activity](cmd/samples/recipes/cancelactivity/)
* **Shows**: Graceful cancellation and cleanup operations.
* **What it does**: Shows how to cancel running activities and perform cleanup operations.
* **Real-world use case**: User-initiated cancellations, timeout handling, resource cleanup, emergency stops.
* **Key concepts**: Cancellation handling, cleanup operations, graceful shutdown.

##### How to run
Start Worker:
```bash
./bin/cancelactivity -m worker
```

Start Workflow:
```bash
./bin/cancelactivity -m trigger
```

**Cancel Workflow:**
```bash
./bin/cancelactivity -m cancel -w <WorkflowID>
```

#### [Mutex](cmd/samples/recipes/mutex/)
* **Shows**: Resource locking and distributed locking patterns.
* **What it does**: Ensures only one workflow can access a specific resource at a time.
* **Real-world use case**: Database migrations, configuration updates, resource allocation, critical section protection.
* **Key concepts**: Distributed locking, resource coordination, mutual exclusion.

##### How to run
Start Worker:
```bash
./bin/mutex -m worker
```

Start Workflow:
```bash
./bin/mutex -m trigger
```

#### [Query](cmd/samples/recipes/query/)
* **Shows**: Workflow state inspection and custom query handlers.
* **What it does**: Demonstrates custom query handlers to inspect workflow state.
* **Real-world use case**: Progress monitoring, status dashboards, debugging running workflows, user interfaces.
* **Key concepts**: Query handlers, state inspection, workflow monitoring.

##### How to run
* Check **[Detailed Guide](cmd/samples/recipes/query/README.md)** to run the sample

#### [Consistent Query](cmd/samples/recipes/consistentquery/)
* **Shows**: Consistent state queries and signal handling.
* **What it does**: Shows how to query workflow state consistently while handling signals.
* **Real-world use case**: Real-time dashboards, progress tracking, state synchronization.
* **Key concepts**: Consistent queries, signal handling, state management.

##### How to run
* Check **[Detailed Guide](cmd/samples/recipes/consistentquery/README.md)** to run the sample

#### [Child Workflow](cmd/samples/recipes/childworkflow/)
* **Shows**: Workflow composition and parent-child workflow relationships.
* **What it does**: Demonstrates parent-child workflow relationships with ContinueAsNew pattern.
* **Real-world use case**: Complex business processes, workflow decomposition, modular workflow design.
* **Key concepts**: Child workflows, ContinueAsNew, workflow composition.

##### How to run
Start Worker:
```bash
./bin/childworkflow -m worker
```

Start Workflow:
```bash
./bin/childworkflow -m trigger
```

#### [Dynamic](cmd/samples/recipes/dynamic/)
* **Shows**: Dynamic activity invocation and string-based execution.
* **What it does**: Demonstrates calling activities using string names for dynamic behavior.
* **Real-world use case**: Plugin systems, dynamic workflow composition, configuration-driven workflows.
* **Key concepts**: Dynamic activity invocation, string-based execution, flexible workflow design.

##### How to run
Start Worker:
```bash
./bin/dynamic -m worker
```

Start Workflow:
```bash
./bin/dynamic -m trigger
```

#### [Local Activity](cmd/samples/recipes/localactivity/)
* **Shows**: High-performance local execution and lightweight operations.
* **What it does**: Shows how to use local activities for quick operations that don't need external execution.
* **Real-world use case**: Data validation, simple calculations, condition checking, fast decision making.
* **Key concepts**: Local activities, performance optimization, lightweight operations.

##### How to run
* Check **[Detailed Guide](cmd/samples/recipes/localactivity/README.md)** to run the sample

#### [Versioning](cmd/samples/recipes/versioning/)
* **Shows**: Safe workflow evolution and backward compatibility.
* **What it does**: Shows workflow versioning with backward compatibility and safe rollbacks.
* **Real-world use case**: Production deployments, feature rollouts, backward compatibility, safe migrations.
* **Key concepts**: Workflow versioning, backward compatibility, safe deployments.

##### How to run
* Check **[Detailed Guide](cmd/samples/recipes/versioning/README.md)** to run the sample

#### [Search Attributes](cmd/samples/recipes/searchattributes/)
* **Shows**: Workflow indexing and search for workflow discovery.
* **What it does**: Shows how to add searchable attributes to workflows and query them.
* **Real-world use case**: Workflow discovery, filtering, reporting, operational dashboards.
* **Key concepts**: Search attributes, workflow indexing, ElasticSearch integration.

##### How to run
* Check **[Detailed Guide](cmd/samples/recipes/searchattributes/README.md)** to run the sample

#### [Context Propagation](cmd/samples/recipes/ctxpropagation/)
* **Shows**: Cross-workflow context and context propagation.
* **What it does**: Demonstrates passing context (like user info, trace IDs) through workflow execution.
* **Real-world use case**: Distributed tracing, user context propagation, audit trails, debugging.
* **Key concepts**: Context propagation, distributed tracing, cross-service context.

##### How to run
Start Worker:
```bash
./bin/ctxpropagation -m worker
```

Start Workflow:
```bash
./bin/ctxpropagation -m trigger
```

#### [Tracing](cmd/samples/recipes/tracing/)
* **Shows**: Distributed tracing and integration with tracing systems.
* **What it does**: Shows how to add distributed tracing to Cadence workflows.
* **Real-world use case**: Performance monitoring, debugging, observability, APM integration.
* **Key concepts**: Distributed tracing, Jaeger integration, observability.

##### How to run
Start Worker:
```bash
./bin/tracing -m worker
```

Start Workflow:
```bash
./bin/tracing -m trigger
```

#### [Side Effect](cmd/samples/recipes/sideeffect/)
* **Shows**: Non-deterministic operations and replay safety.
* **What it does**: Demonstrates the SideEffect API for handling non-deterministic operations.
* **Real-world use case**: ID generation, random number generation, external state queries.
* **Key concepts**: Side effects, non-deterministic operations, replay safety.

##### How to run
Start Workflow:
```bash
./bin/sideeffect
```

#### [Batch](cmd/samples/batch/)
* **Shows**: Batch processing and concurrency control.
* **What it does**: Processes large batches of tasks with controlled concurrency.
* **Real-world use case**: Batch data processing, bulk operations, ETL jobs, report generation.
* **Key concepts**: Batch processing, concurrency control, task distribution.

##### How to run
Start Worker:
```bash
./bin/batch -m worker
```

Start Workflow:
```bash
./bin/batch -m trigger
```

#### [PSO (Particle Swarm Optimization)](cmd/samples/pso/)
* **Shows**: Complex mathematical workflows and long-running optimization workflows.
* **What it does**: Implements particle swarm optimization with child workflows and ContinueAsNew.
* **Real-world use case**: Mathematical optimization, machine learning training, complex calculations.
* **Key concepts**: Long-running workflows, ContinueAsNew, child workflows, custom data converters.

##### How to run
* Check **[Detailed Guide](cmd/samples/pso/README.md)** to run the sample

#### [Recovery](cmd/samples/recovery/)
* **Shows**: Workflow recovery and failure handling.
* **What it does**: Shows how to restart failed workflows and replay signals.
* **Real-world use case**: Disaster recovery, workflow repair, system restoration.
* **Key concepts**: Workflow recovery, signal replay, failure handling.

##### How to run
* Check **[Detailed Guide](cmd/samples/recovery/README.md)** to run the sample

### 🏢 **Business Application Examples**

#### [Expense](cmd/samples/expense/)
* **Shows**: Human-in-the-loop workflows and approval workflows.
* **What it does**: Creates an expense report, waits for approval, then processes payment.
* **Real-world use case**: Expense approval, purchase orders, document review, approval workflows.
* **Key concepts**: Human-in-the-loop, async completion, approval workflows.

##### How to run
* Check **[Detailed Guide](cmd/samples/expense/README.md)** to run the sample

#### [File Processing](cmd/samples/fileprocessing/)
* **Shows**: Distributed file processing across multiple hosts.
* **What it does**: Downloads, processes, and uploads files with host-specific execution.
* **Real-world use case**: Large file processing, ETL pipelines, media processing, data transformation.
* **Key concepts**: File processing, host-specific execution, session management, retry policies.

##### How to run:
Start Worker:
```bash
./bin/fileprocessing -m worker
```

Start Workflow:
```bash
./bin/fileprocessing -m trigger
```

#### [DSL](cmd/samples/dsl/)
* **Shows**: Domain-specific language and custom workflow language creation.
* **What it does**: Implements a simple DSL for defining workflows using YAML configuration.
* **Real-world use case**: Business user workflow definition, configuration-driven workflows, workflow templates.
* **Key concepts**: DSL implementation, YAML parsing, dynamic workflow creation.

##### How to run
* Check **[Detailed Guide](cmd/samples/dsl/README.md)** to run the sample

#### [Page Flow](cmd/samples/pageflow/)
* **Shows**: UI-driven workflows and web application integration.
* **What it does**: Shows a React frontend that interacts with Cadence workflows through signals and queries.
* **Real-world use case**: Multi-step forms, wizard interfaces, approval workflows, user onboarding.
* **Key concepts**: UI integration, signal handling, state management, frontend-backend coordination.

##### How to run
* Check **[Detailed Guide](cmd/samples/pageflow/README.md)** to run the sample

## 🛠 **Development & Testing**

### Building Samples
```bash
make
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run specific sample tests
go test ./cmd/samples/recipes/helloworld/
```

### Worker Modes
Most samples support these modes:
- `worker`: Start a worker to handle workflow execution
- `trigger`: Start a new workflow execution
- `query`: Query a running workflow (where applicable)
- `signal`: Send a signal to a workflow (where applicable)


## 🤝 **Contributing**

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 **License**

Apache 2.0 License - see [LICENSE](LICENSE) for details.

## 🆘 **Getting Help**

- **Documentation**: [Cadence Documentation](https://cadenceworkflow.io/docs/)
- **Community**: [Cadence Community](https://cadenceworkflow.io/community/)
- **Issues**: [GitHub Issues](https://github.com/uber-common/cadence-samples/issues)

---

**Happy Workflowing! 🚀**