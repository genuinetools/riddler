# riddler

[![make-all](https://github.com/genuinetools/riddler/workflows/make%20all/badge.svg)](https://github.com/genuinetools/riddler/actions?query=workflow%3A%22make+all%22)
[![make-image](https://github.com/genuinetools/riddler/workflows/make%20image/badge.svg)](https://github.com/genuinetools/riddler/actions?query=workflow%3A%22make+image%22)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/riddler)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/riddler/total.svg?style=for-the-badge)](https://github.com/genuinetools/riddler/releases)


A tool to convert `docker inspect` to the
[opencontainers/specs](https://github.com/opencontainers/specs)
and [opencontainers/runc](https://github.com/opencontainers/runc).

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [NOTE](#note)
- [Installation](#installation)
    - [Binaries](#binaries)
    - [Via Go](#via-go)
- [Usage](#usage)
- [Installation](#installation-1)
  - [TODO](#todo)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## NOTE

This project is no longer maintained. If you are using a version of
docker greater than 1.11 then you can just copy the config from
`/var/run/docker/libcontainerd` like so:

```console
$ docker ps -a
CONTAINER ID    CREATED             STATUS              PORTS               NAMES
d4da95779a3c    3 minutes ago       Up 3 minutes        80/tcp              modest_meitner

$ sudo tree /var/run/docker/libcontainerd -L 1
/var/run/docker/libcontainerd
├── containerd
├── d4da95779a3c287b28b421194f04374b6330e6ff10f5ca1a99d03828d84f1635
├── docker-containerd.pid
├── docker-containerd.sock
└── event.ts

$ sudo tree /var/run/docker/libcontainerd/d4da95779a3c.../
/var/run/docker/libcontainerd/d4da95779a3c.../
├── config.json
├── init-stderr
├── init-stdin
└── init-stdout

$ sudo file /var/run/docker/libcontainerd/d4da95779a3c.../config.json
/var/run/docker/libcontainerd/d4da95779a3c.../config.json: ASCII text, with very long lines
```

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/genuinetools/riddler/releases).

#### Via Go

```console
$ go get github.com/genuinetools/riddler
```

## Usage

```console
$ riddler -h
riddler -  A tool to convert docker inspect to the opencontainers runc spec.

Usage: riddler <command>

Flags:

  --host       Docker Daemon socket(s) to connect to (default: unix:///var/run/docker.sock)
  --idlen      Length of UID/GID ID space ranges for user namespaces (default: 0)
  --idroot     Root UID/GID for user namespaces (default: 0)
  --bundle     Path to the root of the bundle directory (default: <none>)
  -d           enable debug logging (default: false)
  -f, --force  force overwrite existing files (default: false)
  --hook       Hooks to prefill into spec file. (ex. --hook prestart:netns) (default: [])

Commands:

  version  Show the version information.
```

## Installation

For seccomp and apparmor support you will need:

- `sys/apparmor.h`
- `seccomp.h`

**OR** to compile without those run:

```console
$ make build BUILDTAGS=""
```


**example**

```console
# just pass the container name or id on run

$ riddler chrome
config.json has been saved.
```

### TODO

- fixup various todos (mostly runtime config parsing)
- add more unit tests for each field
