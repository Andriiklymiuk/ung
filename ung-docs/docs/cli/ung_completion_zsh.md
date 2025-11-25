---
id: ung_completion_zsh
title: ung completion zsh
---

## ung completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

```bash
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

To load completions in your current shell session:

```bash
source <(ung completion zsh)
```

To load completions for every new session, execute once:

#### Linux:

```bash
ung completion zsh > "${fpath[1]}/_ung"
```

#### macOS:

```bash
ung completion zsh > $(brew --prefix)/share/zsh/site-functions/_ung
```

You will need to start a new shell for this setup to take effect.


```
ung completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [ung completion](./ung_completion)	 - Generate the autocompletion script for the specified shell

