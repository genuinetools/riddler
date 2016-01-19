# riddler

[![Circle CI](https://circleci.com/gh/jfrazelle/riddler.svg?style=svg)](https://circleci.com/gh/jfrazelle/riddler)


`docker inspect` to
[opencontainers/specs](https://github.com/opencontainers/specs)
/[opencontainers/runc](https://github.com/opencontainers/runc) generator.

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

**example**

```console
# just pass the container name or id on run

$ riddler chrome
config.json and runtime.json have been saved.
```

### TODO

- fixup various todos (mostly runtime config parsing)
- add more unit tests for each field
