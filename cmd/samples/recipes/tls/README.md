# TLS Connection Verification Recipe

This recipe demonstrates how to verify TLS connections between Cadence SDK and frontend using a Cadence workflow, showing both successful and unsuccessful scenarios.

## What This Recipe Shows

1. **TLS Workflow Execution** - Orchestrates TLS testing using Cadence workflow and activities
2. **Certificate Management** - Automated certificate creation and cleanup
3. **TLS Connection Testing** - Tests connections with valid certificates
4. **Error Handling** - Tests connections with missing certificates
5. **Standard Connection Fallback** - Tests non-TLS connections

## Prerequisites

- Cadence server running on `localhost:7933`
- OpenSSL installed for certificate generation
- Go environment set up

## Steps to Run Sample

1. You need a cadence service running. See details in cmd/samples/README.md

2. Run the following command to start the worker:
   ```
   ./bin/tls -m worker
   ```

3. Run the following command to execute the workflow:
   ```
   ./bin/tls -m trigger
   ```

## Workflow Structure

The recipe follows the standard Cadence worker/trigger pattern:

- **Worker Mode**: Starts a Cadence worker that can execute the TLS workflow and activities
- **Trigger Mode**: Triggers a new execution of the TLS workflow

## Activities

1. **setupCertificatesActivity** - Creates fresh TLS certificates for testing
2. **testTLSConnectionActivity** - Tests TLS connections with provided certificates
3. **testStandardConnectionActivity** - Tests standard non-TLS connections
4. **cleanupCertificatesActivity** - Cleans up generated certificates

## Expected Workflow Output

When you run the trigger mode, the workflow will execute and you'll see logs showing:

```
INFO  TLS Workflow started
INFO  Setting up certificates for testing
INFO  Certificates setup completed
INFO  Testing TLS connection with valid certificates
INFO  Valid TLS connection test result: SUCCESS - TLS connection established
INFO  Testing TLS connection with missing certificates
INFO  Missing certificates test failed as expected
INFO  Testing standard non-TLS connection
INFO  Standard connection test result: SUCCESS - Standard connection established
INFO  Cleaning up certificates
INFO  Certificate cleanup completed
INFO  TLS Workflow completed successfully
```

## Certificate Management

The workflow automatically:
- Creates a `cmd/samples/recipes/tls/certs/` directory
- Generates CA, server, and client certificates using OpenSSL
- Tests connections with the generated certificates
- Cleans up all certificate files after testing

Generated certificates include:
- `ca.key` / `ca.crt` - Certificate Authority
- `server.key` / `server.crt` - Server certificates
- `client.key` / `client.crt` - Client certificates

## Key Features

- **Workflow Orchestration**: Uses Cadence workflow to coordinate TLS testing
- **Automated Certificate Generation**: No manual certificate setup required
- **Comprehensive Testing**: Tests multiple TLS scenarios in sequence
- **Error Handling**: Graceful handling of connection failures
- **Clean Resource Management**: Automatic cleanup of generated certificates
- **Production Patterns**: Demonstrates proper Cadence workflow and activity patterns

## Code Structure

### Main Components
- `main.go` - Worker/trigger entry point following Cadence patterns
- `tls_workflow.go` - Workflow and activity definitions

### Key Functions
- `tlsWorkflow()` - Main workflow that orchestrates TLS testing
- `setupCertificatesActivity()` - Activity to create test certificates
- `testTLSConnectionActivity()` - Activity to test TLS connections
- `testStandardConnectionActivity()` - Activity to test non-TLS connections
- `cleanupCertificatesActivity()` - Activity to clean up certificates
- `createTLSClient()` - Helper to create TLS-enabled Cadence client

This recipe demonstrates how to integrate TLS configuration testing into Cadence workflows, making it suitable for production environments where TLS verification is part of automated testing or deployment processes.