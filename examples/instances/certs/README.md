# SSL/TLS Certificates (Mock)

These are **self-signed throwaway certificates for example purposes only**. Do not use in production.

Bytebase 3.17.1+ parses inline `ssl_ca` / `ssl_cert` / `ssl_key` PEM at save time, so the files committed here are real, parseable PEM. The CA and client material are issued under a single throwaway `CN=Bytebase Terraform Example Mock CA` that exists only for this example to apply cleanly.

## Files

- `ca.pem`, `ca-path.pem` - CA certificate (used by the inline-PEM and path-based example respectively)
- `client-cert.pem`, `client-cert-path.pem` - Client certificate
- `client-key.pem`, `client-key-path.pem` - Client private key

The two file sets ship identical content. They exist as separate files so the inline and path-based examples in `../main.tf` reference visibly distinct paths.

## Usage

Replace these files with your actual SSL/TLS certificates:

```bash
# Example: Copy your real certificates
cp /path/to/your/ca.pem ./ca.pem
cp /path/to/your/client-cert.pem ./client-cert.pem
cp /path/to/your/client-key.pem ./client-key.pem
```

## Generating Test Certificates

For testing purposes, you can generate self-signed certificates:

```bash
# Generate CA key and certificate
openssl genrsa -out ca-key.pem 2048
openssl req -new -x509 -days 365 -key ca-key.pem -out ca.pem -subj "/CN=Test CA"

# Generate client key and certificate
openssl genrsa -out client-key.pem 2048
openssl req -new -key client-key.pem -out client.csr -subj "/CN=client"
openssl x509 -req -days 365 -in client.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem
```
