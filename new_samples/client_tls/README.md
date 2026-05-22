# Client TLS Sample

This sample demonstrates how to connect to a Cadence server using TLS (Transport Layer Security) for secure communication.

## Prerequisites

1. A Cadence server configured with TLS enabled
2. Client certificates in the `credentials/` directory:
   - `credentials/client.crt` - Client certificate
   - `credentials/client.key` - Client private key
   - `credentials/keytest.crt` - Server CA certificate

## Running the Sample

### Register a Domain

Before starting workflows, you may need to register a domain:

```bash
go run . register-domain
```

### Start a Workflow

To start a HelloWorld workflow with TLS:

```bash
go run . start-workflow
```

**Note:** This requires a worker running to execute the workflow. Start a worker from the `hello_world/` sample first.

## What This Sample Demonstrates

### TLS Configuration

The `tls_config.go` file shows how to:
- Load client certificates for mutual TLS authentication
- Load server CA certificates for server verification
- Configure TLS options for gRPC connections

### Cadence Client Setup

The `cadence_client.go` file shows how to:
- Create a YARPC dispatcher with TLS-enabled gRPC transport
- Build a Cadence client using the proto API adapter

### Client Operations

The `main.go` file demonstrates:
- Starting a workflow execution programmatically
- Registering a domain programmatically

## Security Notes

- The sample uses `InsecureSkipVerify: true` for development convenience
- In production, always verify server certificates by setting `InsecureSkipVerify: false`
- Store credentials securely and never commit them to version control

## References

- [Cadence TLS Documentation](https://cadenceworkflow.io/docs/operation-guide/tls)
- [Go TLS Configuration](https://pkg.go.dev/crypto/tls)
