# Client TLS Sample

This sample demonstrates how to connect to a Cadence server using TLS (Transport Layer Security) for secure communication.

## Prerequisites

1. A Cadence server started with the TLS development config, from the Cadence repo:
   `./cadence-server --zone tls start` (uses `config/development_tls.yaml`)
2. The matching development certificates copied into this sample's `credentials/` directory:
   ```bash
   mkdir -p credentials
   cp <path-to-cadence>/config/credentials/{client.crt,client.key,keytest.crt} credentials/
   ```
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

- The client verifies the server certificate chain against `credentials/keytest.crt` and checks
  the server hostname. Do not disable this with `InsecureSkipVerify`: it also disables the CA check,
  so the client would accept any server
- If you use your own certificates, the server certificate's Subject Alternative Name must match
  the address the client dials (`127.0.0.1` by default). Set `CADENCE_TLS_SERVER_NAME` if the
  certificate is issued for a different name
- Store credentials securely and never commit them to version control

## References

- [Cadence TLS Documentation](https://cadenceworkflow.io/docs/operation-guide/tls)
- [Go TLS Configuration](https://pkg.go.dev/crypto/tls)
