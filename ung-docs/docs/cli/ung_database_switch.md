---
id: ung_database_switch
title: ung database switch
---

## ung database switch

Switch to a different database

### Synopsis

Switch to a different database by creating or updating the local workspace config.

The database path can be:
  - Relative: ./mydb.db or data/billing.db
  - Absolute: /full/path/to/db.db
  - Tilde: ~/.ung/production.db

Examples:
  ung db switch ./client-a.db
  ung db switch ~/.ung/production.db

```
ung database switch <database-path> [flags]
```

### Options

```
  -h, --help   help for switch
```

### SEE ALSO

* [ung database](./ung_database)	 - Manage multiple databases

