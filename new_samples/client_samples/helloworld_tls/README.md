## Pre-requisites

Follow this document to start cadence server:
https://github.com/cadence-workflow/cadence/blob/e1267de12f8bc670fc84fab456d3495c8fc2f8a8/CONTRIBUTING.md#L1

1. **Build tools in cadence server**
   ```bash
   make bins
   ```

2. **Start cassandra**
   ```bash
   docker compose -f ./docker/dev/cassandra.yml up -d
   ```

3. **Install schema**
   ```bash
   make install-schema
   ```

4. **Start cadence server with TLS**
   ```bash
   ./cadence-server --env development --zone tls start
   ```

## Running the Sample

### Step 1: Download Certificates
Download certificates from config/credentials of cadence server and place them in below folder

```bash
new_samples/client_samples/helloworld_tls/credentials
```

### Step 2: Register the Domain
Before running workflows, you must register the "default" domain:

```bash
cd new_samples/client_samples/helloworld_tls
go run register_domain.go
```

Expected output:
```
Successfully registered domain  {"domain": "default"}
```

If the domain already exists, you'll see:
```
Domain already exists  {"domain": "default"}
```

### Step 3: Run the Sample
In another terminal:
```bash
cd new_samples/client_samples/helloworld_tls
go run hello_world_tls.go
```

## References

- [Cadence Official Certificates](https://github.com/cadence-workflow/cadence/tree/master/config/credentials)
- [Cadence Documentation](https://cadenceworkflow.io/)
- [Go TLS Package](https://pkg.go.dev/crypto/tls)

