# manic-go

## Usage
> Download accelerator

manic-go

## Description

```
Download accelerator for fast HTTP downloads with SHA256 sum verification
```

## Flags
|Flag|Usage|
|----|-----|
|`--debug`|enable debug messages|
|`--disable-update-checks`|disables update checks|
|`--raw`|print unstyled raw output (set it if output is written to a file)|

## Commands
|Command|Usage|
|-------|-----|
|`manic-go completion`|Generate the autocompletion script for the specified shell|
|`manic-go download`|Download file over HTTP|
|`manic-go help`|Help about any command|
|`manic-go tests`|Testing downloading|
# ... completion
`manic-go completion`

## Usage
> Generate the autocompletion script for the specified shell

manic-go completion

## Description

```
Generate the autocompletion script for manic-go for the specified shell.
See each sub-command's help for details on how to use the generated script.

```

## Commands
|Command|Usage|
|-------|-----|
|`manic-go completion bash`|Generate the autocompletion script for bash|
|`manic-go completion fish`|Generate the autocompletion script for fish|
|`manic-go completion powershell`|Generate the autocompletion script for powershell|
|`manic-go completion zsh`|Generate the autocompletion script for zsh|
# ... completion bash
`manic-go completion bash`

## Usage
> Generate the autocompletion script for bash

manic-go completion bash

## Description

```
Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(manic-go completion bash)

To load completions for every new session, execute once:

#### Linux:

	manic-go completion bash > /etc/bash_completion.d/manic-go

#### macOS:

	manic-go completion bash > $(brew --prefix)/etc/bash_completion.d/manic-go

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion fish
`manic-go completion fish`

## Usage
> Generate the autocompletion script for fish

manic-go completion fish

## Description

```
Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	manic-go completion fish | source

To load completions for every new session, execute once:

	manic-go completion fish > ~/.config/fish/completions/manic-go.fish

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion powershell
`manic-go completion powershell`

## Usage
> Generate the autocompletion script for powershell

manic-go completion powershell

## Description

```
Generate the autocompletion script for powershell.

To load completions in your current shell session:

	manic-go completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion zsh
`manic-go completion zsh`

## Usage
> Generate the autocompletion script for zsh

manic-go completion zsh

## Description

```
Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(manic-go completion zsh); compdef _manic-go manic-go

To load completions for every new session, execute once:

#### Linux:

	manic-go completion zsh > "${fpath[1]}/_manic-go"

#### macOS:

	manic-go completion zsh > $(brew --prefix)/share/zsh/site-functions/_manic-go

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... download
`manic-go download`

## Usage
> Download file over HTTP

manic-go download [url]

## Flags
|Flag|Usage|
|----|-----|
|`-c, --check string`|Compare to a sha256sum|
|`-o, --output string`|Save to file|
|`-p, --progress`|Progress bar|
|`--proxy string`|Proxy servers to use|
|`-t, --threads int`|Maximum amount of threads (default 2)|
|`-T, --timeout int`|Set I/O and connection timeout (default 30)|
|`-w, --workers int`|amount of concurrent workers (default 3)|
# ... help
`manic-go help`

## Usage
> Help about any command

manic-go help [command]

## Description

```
Help provides help for any command in the application.
Simply type manic-go help [path to command] for full details.
```
# ... tests
`manic-go tests`

## Usage
> Testing downloading

manic-go tests

## Description

```
Command used for testing the program.
	To use it, pass it the url and optionally workers and a sha256sum to compare with
	By default amount of workers is 2
```

## Flags
|Flag|Usage|
|----|-----|
|`-c, --check string`|Compare to a sha256sum|
|`-o, --output string`|Save to file|
|`-p, --progress`|Progress bar|
|`-t, --threads int`|Maximum amount of threads (default 2)|
|`-w, --workers int`|amount of concurrent workers (default 3)|


---
> **Documentation automatically generated with [PTerm](https://github.com/pterm/cli-template) on 05 April 2023**
