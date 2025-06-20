# Multi-stage build for optimization
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all samples using existing Makefile
RUN make

# Final stage - minimal runtime image
FROM alpine:3.22

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash curl

# Build argument for Cadence host configuration
ARG CADENCE_HOST=localhost:7833

# Create non-root user
RUN addgroup -g 1001 cadence && \
    adduser -D -u 1001 -G cadence cadence

# Set working directory
WORKDIR /home/cadence

# Copy built binaries from builder stage
COPY --from=builder /app/bin/ ./bin/

# Copy configuration files
COPY --from=builder /app/config/ ./config/

# Copy cmd directory
COPY --from=builder /app/cmd/ ./cmd/

# Copy new_samples directory
COPY --from=builder /app/new_samples/ ./new_samples/

# Update config file with the provided Cadence host
RUN sed -i "s/host: \"localhost:7833\"/host: \"${CADENCE_HOST}\"/" config/development.yaml

# Change ownership of files
RUN chown -R cadence:cadence /home/cadence

# Switch to non-root user
USER cadence

# Add bin directory to PATH
ENV PATH="/home/cadence/bin:${PATH}"

# Default command - interactive shell
CMD ["/bin/bash"]