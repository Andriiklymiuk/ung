---
id: ung_completion_powershell
title: ung completion powershell
---

## ung completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	ung completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
ung completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [ung completion](./ung_completion)	 - Generate the autocompletion script for the specified shell

