## Running the Sample

### Step 1: Download Certificates
```bash
cd new_samples/client_samples/helloworld_tls/credentials
chmod +x download-certs.sh
./download-certs.sh
cd ..
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

