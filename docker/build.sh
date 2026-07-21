#!/bin/bash

# Default values
DOCKER_REGISTRY=${DOCKER_REGISTRY:-"docker.io/your-username"}
TAG=${TAG:-"latest"}

# Build the Docker image
echo "Building Docker image..."
docker build -t ${DOCKER_REGISTRY}/cadence-helloworld-worker:${TAG} -f docker/docker/Dockerfile .

# Push the image if requested
if [ "$1" == "push" ]; then
    echo "Pushing Docker image to registry..."
    docker push ${DOCKER_REGISTRY}/cadence-helloworld-worker:${TAG}
fi