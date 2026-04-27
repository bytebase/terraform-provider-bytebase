# SSH Keys (Mock)

This is a **mock SSH private key for example purposes only**. Do not use in production.

## Files

- `id_rsa` - SSH private key for tunnel authentication

## Usage

Replace this file with your actual SSH private key:

```bash
# Example: Copy your real SSH key
cp ~/.ssh/id_rsa ./id_rsa
chmod 600 ./id_rsa
```

## Generating Test SSH Keys

For testing purposes, you can generate a new SSH key pair:

```bash
# Generate SSH key pair
ssh-keygen -t rsa -b 4096 -f ./id_rsa -N ""

# The public key (id_rsa.pub) should be added to the SSH server's authorized_keys
```

## Security Note

Never commit real SSH private keys to version control. Consider using:
- Environment variables
- Terraform variables with sensitive flag
- External secret management (Vault, AWS Secrets Manager, etc.)
