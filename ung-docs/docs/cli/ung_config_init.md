---
id: ung_config_init
title: ung config init
---

## ung config init

Initialize workspace configuration

### Synopsis

Create a local .ung.yaml configuration file in the current directory.

This is useful for project-specific databases where you want to keep
your billing data separate from the global database.

Example workspace config:
  database_path: ./ung.db
  invoices_dir: ./invoices
  language: en

```
ung config init [flags]
```

### Options

```
  -g, --global   Create global config (~/.ung/config.yaml)
  -h, --help     help for init
      --icloud   Use iCloud Drive for storage (macOS only)
  -l, --local    Create local workspace config (.ung.yaml) (default true)
```

### SEE ALSO

* [ung config](./ung_config)	 - Manage ung configuration

