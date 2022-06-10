# manic-go

## Usage
> Program intended to be a port of my manic library but also with a cli

manic-go

## Description

```
Port of my manic library but with a cli, the same is planned for the rust version soon
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
|`manic-go gh`|A brief description of your command|
|`manic-go help`|Help about any command|
|`manic-go tests`|Testing downloading|
# ... download
`manic-go download`

## Usage
> Download file over HTTP

manic-go download

## Flags
|Flag|Usage|
|----|-----|
|`-c, --check string`|Compare to a sha256sum|
|`-o, --output string`|Save to file|
|`-p, --progress`|Progress bar|
|`-t, --threads int`|Maximum amount of threads (default 2)|
|`-T, --timeout int`|Set I/O and connection timeout (default 30)|
|`-w, --workers int`|amount of concurrent workers (default 3)|
# ... gh
`manic-go gh`

## Usage
> A brief description of your command

manic-go gh

## Description

```
A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.
```

## Flags
|Flag|Usage|
|----|-----|
|`-i, --interactive`|Choose the release to download via a menu|
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
> **Documentation automatically generated with [PTerm](https://github.com/pterm/cli-template) on 10 June 2022**
