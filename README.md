# riddler

[![Travis CI](https://travis-ci.org/jessfraz/riddler.svg?branch=master)](https://travis-ci.org/jessfraz/riddler)


`docker inspect` to
[opencontainers/specs](https://github.com/opencontainers/specs)
/[opencontainers/runc](https://github.com/opencontainers/runc) generator.

> **NOTE:** This project is no longer maintained. If you are using a version of
> docker greater than 1.11 then you can just copy the config from
> `/var/run/docker/libcontainerd` like so:

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

```console
$ riddler --help
      _     _     _ _
 _ __(_) __| | __| | | ___ _ __
| '__| |/ _` |/ _` | |/ _ \ '__|
| |  | | (_| | (_| | |  __/ |
|_|  |_|\__,_|\__,_|_|\___|_|
 docker inspect to opencontainers runc spec generator.
 Version: v0.1.0

 -bundle string
        Path to the root of the bundle directory
  -d    run in debug mode
  -f    force overwrite existing files
  -force
        force overwrite existing files
  -hook value
        Hooks to prefill into spec file. (ex. --hook prestart:netns) (default [])
  -host string
        Docker Daemon socket(s) to connect to (default "unix:///var/run/docker.sock")
  -v    print version and exit (shorthand)
  -version
        print version and exit
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
