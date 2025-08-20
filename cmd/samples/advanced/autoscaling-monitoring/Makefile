# Makefile for autoscaling monitoring sample

.PHONY: build clean test

# Build the autoscaling monitoring sample
build:
	go build -o ../../../../bin/autoscaling-monitoring *.go

# Clean build artifacts
clean:
	rm -f ../../../../bin/autoscaling-monitoring

# Run tests
test:
	go test -v .

# Install dependencies
deps:
	go mod tidy

# Run the sample in different modes
run-worker: build
	../../../../bin/autoscaling-monitoring -m worker

run-trigger: build
	../../../../bin/autoscaling-monitoring -m trigger
