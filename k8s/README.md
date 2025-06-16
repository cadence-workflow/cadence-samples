# Cadence Samples Usage Guide

This guide explains how to build, deploy, and use the Cadence samples container for testing workflows.

## Prerequisites: Domain Registration

Before running any samples, you must first register the domain in Cadence. Execute this command in the cadence-frontend pod:

```bash
# Access the cadence-frontend pod
kubectl exec -it <cadence-frontend-pod-name> -n cadence -- /bin/bash

# Register the default domain
cadence --address $(hostname -i):7833 \
    --transport grpc \
    --domain default \
    domain register \
    --retention 1
```

**Note**: Replace `<cadence-frontend-pod-name>` with your actual cadence-frontend pod name and adjust the namespace if needed.

## Building the Docker Image

Build the samples image with your Cadence host configuration:

```bash
docker build --build-arg CADENCE_HOST="cadence-frontend.cadence.svc.cluster.local:7833" -t cadence-samples:latest .
```

**Important**: Replace `cadence-frontend.cadence.svc.cluster.local:7833` with your actual Cadence frontend service address.

### Examples for Different Environments

```bash
# Local development
docker build --build-arg CADENCE_HOST="localhost:7833" -t cadence-samples:latest -f Dockerfile.samples .

# Kubernetes cluster (same namespace)
docker build --build-arg CADENCE_HOST="cadence-frontend.cadence.svc.cluster.local:7833" -t cadence-samples:latest .

# Different namespace
docker build --build-arg CADENCE_HOST="cadence-frontend.my-namespace.svc.cluster.local:7833" -t cadence-samples:latest .
```

## Upload to Container Registry

Tag and push your image to your container registry:

```bash
# Tag the image
docker tag cadence-samples:latest your-registry.com/cadence-samples:latest

# Push to registry
docker push your-registry.com/cadence-samples:latest
```

## Kubernetes Deployment

### Pod Configuration

Edit the provided YAML file:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: cadence-samples
  namespace: cadence # Change to your namespace
  labels:
    app: cadence-samples
spec:
  containers:
  - name: cadence-samples
    image: cadence-samples:latest  # Change to your registry image
    imagePullPolicy: IfNotPresent
    command: ["/bin/bash"]
    args: ["-c", "sleep infinity"]
    workingDir: /home/cadence
    env:
    - name: HOME
      value: "/home/cadence"
    resources:
      requests:
        memory: "128Mi"
        cpu: "100m"
      limits:
        memory: "1Gi"
        cpu: "1"
  restartPolicy: Always
  securityContext:
    runAsUser: 1001
    runAsGroup: 1001
    fsGroup: 1001
```

**Required Changes**:
1. **`namespace`**: Change to your Cadence namespace
2. **`image`**: Change to your registry image path

### Deploy the Pod

```bash
kubectl apply -f cadence-samples-pod.yaml
```

## Using the Samples

### Step 1: Access the Container

```bash
kubectl exec -it cadence-samples -n cadence -- /bin/bash
```

### Step 2: Run Workflow Examples

#### Terminal 1 - Start the Worker
```bash
# Example: Hello World worker
./bin/helloworld -m worker
```

#### Terminal 2 - Trigger the Workflow
Open a second terminal and execute:
```bash
kubectl exec -it cadence-samples -n cadence -- /bin/bash
./bin/helloworld -m trigger
```

#### Stop the Worker
In Terminal 1, press `Ctrl+C` to stop the worker.

### Some Available Sample Commands

```bash
# Hello World
./bin/helloworld -m worker
./bin/helloworld -m trigger

# File Processing
./bin/fileprocessing -m worker
./bin/fileprocessing -m trigger

# DSL Example
./bin/dsl -m worker
./bin/dsl -m trigger -dslConfig cmd/samples/dsl/workflow1.yaml
./bin/dsl -m trigger -dslConfig cmd/samples/dsl/workflow2.yaml
```

## Complete Sample Documentation

For all available samples, detailed explanations, and source code, visit:
**https://github.com/cadence-workflow/cadence-samples**

This repository contains comprehensive documentation for each sample workflow pattern and advanced usage examples.