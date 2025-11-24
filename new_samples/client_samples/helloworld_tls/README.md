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
To enable mTLS in Cadence server, you need to configure TLS settings and start the server with the appropriate environment configuration.
Starting the Server with TLS
Use the --zone flag to specify the TLS configuration when starting the Cadence server:

./cadence-server --env development --zone tls start

This will load [config/development.yaml](https://github.com/cadence-workflow/cadence/blob/e1267de12f8bc670fc84fab456d3495c8fc2f8a8/config/development.yaml) + [config/development_tls.yaml](https://github.com/cadence-workflow/cadence/blob/e1267de12f8bc670fc84fab456d3495c8fc2f8a8/config/development_tls.yaml). 
See [CONTRIBUTING.md](https://github.com/cadence-workflow/cadence/blob/e1267de12f8bc670fc84fab456d3495c8fc2f8a8/CONTRIBUTING.md#4-run) for more details. 

## Running the Sample

### Step 1: Download Certificates
Download certificates from config/credentials of cadence server and place them in below folder
Or follow below steps

```bash
mkdir -p new_samples/client_samples/helloworld_tls/credentials

$ curl -s -O https://raw.githubusercontent.com/cadence-workflow/cadence/master/config/credentials/client.crt
$ curl -s -O https://raw.githubusercontent.com/cadence-workflow/cadence/master/config/credentials/client.key
$ curl -s -O https://raw.githubusercontent.com/cadence-workflow/cadence/master/config/credentials/keytest.crt

```

### Step 2: Register the Domain
Before running workflows, you must register the "default" domain:

```bash
cd ..
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
go run hello_world_tls.go
```

## References

- [Cadence Official Certificates](https://github.com/cadence-workflow/cadence/tree/master/config/credentials)
- [Cadence Documentation](https://cadenceworkflow.io/)
- [Go TLS Package](https://pkg.go.dev/crypto/tls)

