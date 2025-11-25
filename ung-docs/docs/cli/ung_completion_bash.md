---
id: ung_completion_bash
title: ung completion bash
---

## ung completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(ung completion bash)

To load completions for every new session, execute once:

#### Linux:

	ung completion bash > /etc/bash_completion.d/ung

#### macOS:

	ung completion bash > $(brew --prefix)/etc/bash_completion.d/ung

You will need to start a new shell for this setup to take effect.


```
ung completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [ung completion](./ung_completion)	 - Generate the autocompletion script for the specified shell

