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
|`manic-go download`|Download file over HTTP|
|`manic-go help`|Help about any command|
|`manic-go tests`|Testing downloading|
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
> **Documentation automatically generated with [PTerm](https://github.com/pterm/cli-template) on 11 August 2022**
