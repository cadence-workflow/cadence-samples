FROM golang:1.21-alpine

# Install all necessary dependencies for building and running
RUN apk add --no-cache git make gcc musl-dev ca-certificates nano curl bash sed

# Build argument for Cadence host configuration
ARG CADENCE_HOST=localhost:7833

# Create non-root user
RUN addgroup -g 1001 cadence && \
    adduser -D -u 1001 -G cadence cadence

# Set working directory
WORKDIR /home/cadence

# Clone cadence-samples repository
RUN git clone https://github.com/cadence-workflow/cadence-samples.git .

# Update config file with the provided Cadence host
RUN sed -i "s/host: \"localhost:7833\"/host: \"${CADENCE_HOST}\"/" config/development.yaml

# Build all samples
RUN make

# Change ownership of files
RUN chown -R cadence:cadence /home/cadence

# Switch to non-root user
USER cadence

# Default command - interactive shell
CMD ["/bin/bash"]