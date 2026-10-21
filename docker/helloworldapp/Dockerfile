FROM golang:1.21-alpine AS builder

# Install git for downloading dependencies
RUN apk add --no-cache git

WORKDIR /app
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /cadence-worker .

# Use minimal base image
FROM alpine:3.18

WORKDIR /
COPY --from=builder /cadence-worker /cadence-worker

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Expose port for health checks
EXPOSE 8080

# Environment variable for Cadence host:port
ENV CADENCE_HOST_PORT=cadence:7833

ENTRYPOINT ["/cadence-worker"]