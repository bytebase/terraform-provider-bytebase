# Kerberos Keytab (Mock)

This is a **mock Kerberos keytab file for example purposes only**. Do not use in production.

## Files

- `hive.keytab` - Kerberos keytab for Hive authentication

## Usage

Replace this file with your actual Kerberos keytab:

```bash
# Generate keytab using kadmin
kadmin -q "xst -k hive.keytab hive/hive.example.com@EXAMPLE.COM"

# Or copy from your Kerberos admin
cp /etc/security/keytabs/hive.keytab ./hive.keytab
chmod 600 ./hive.keytab
```

## Kerberos Configuration

The keytab must match the principal configured in the Terraform resource:

```terraform
sasl_config {
  kerberos {
    primary  = "hive"                    # Service name
    instance = "hive.example.com"        # Hostname
    realm    = "EXAMPLE.COM"             # Kerberos realm
    keytab   = base64encode(file("./kerberos/hive.keytab"))
    kdc_host = "kdc.example.com"         # KDC server
    kdc_port = "88"                      # KDC port
  }
}
```

## Security Note

Never commit real keytab files to version control. Consider using:
- External secret management
- Terraform variables with sensitive flag
- Encrypted storage
