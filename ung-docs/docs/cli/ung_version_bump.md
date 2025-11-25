---
id: ung_version_bump
title: ung version bump
---

## ung version bump

Bump version and create git tag

### Synopsis

Bump the version number and create a git tag.

Examples:
  ung version bump patch  # 1.2.3 -> 1.2.4
  ung version bump minor  # 1.2.3 -> 1.3.0
  ung version bump major  # 1.2.3 -> 2.0.0

This command will:
1. Get the latest git tag
2. Increment the version number
3. Create a new git tag
4. Display the new version

```
ung version bump [patch|minor|major] [flags]
```

### Options

```
  -h, --help   help for bump
```

### SEE ALSO

* [ung version](./ung_version)	 - Display version information

