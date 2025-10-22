# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hello-world .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/hello-world .

# Expose port 8080 for health checks
EXPOSE 8080

# Environment variables with defaults
ENV CADENCE_HOST=cadence:7833
ENV CADENCE_DOMAIN=test-domain
ENV TASK_LIST=test-worker
ENV CLIENT_NAME=test-worker

# Run the worker
CMD ["./hello-world"]


