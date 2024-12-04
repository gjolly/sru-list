# SRU list

A simple tool to list the Stable Release Updates processes currently in progress for all stable Ubuntu releases.

## Install

```bash
go install github.com/gjolly/sru-list@latest
```

Or

```bash
git clone https://github.com/gjolly/sru-list
go build
```

## Usage

A config file is required to run the tool. A few examples are provided in the [configs](./configs) folder.

```bash
./sru-list ./config/all.yaml
```

## Configuration

```yaml
# A list of packages to watch for.
# The keyword ALL can be used to list everything.
packages:
  - bash
  - systemd

# An optional list of regexp to watch for.
package_regexps:
  - grub-*
```

## How it works

This tool fetches this file https://ubuntu-archive-team.ubuntu.com/sru_report.yaml and parses the information in contains.
