---
id: ung_security_enable
title: ung security enable
---

## ung security enable

Enable database encryption

### Synopsis

Enable encryption for the database. This will encrypt the database file at rest
using AES-256-GCM encryption with PBKDF2 key derivation.

The database will be encrypted with a password that you provide. You'll need to
enter this password each time you use ung, or set it via the UNG_DB_PASSWORD
environment variable.

Example:
  ung security enable
  export UNG_DB_PASSWORD="your-password"

```
ung security enable [flags]
```

### Options

```
  -h, --help   help for enable
```

### SEE ALSO

* [ung security](./ung_security)	 - Manage database security and encryption

