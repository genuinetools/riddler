package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/genuinetools/pkg/cli"
	"github.com/genuinetools/riddler/parse"
	"github.com/genuinetools/riddler/version"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

//
const (
	specConfig = "config.json"
)

var (
	bundle     string
	dockerHost string
	force      bool

	hooks     specs.Hooks
	hookflags stringSlice

	idroot, idlen       uint32
	idrootVar, idlenVar int

	debug bool
)

// stringSlice is a slice of strings
type stringSlice []string

// implement the flag interface for stringSlice
func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
func (s stringSlice) ParseHooks() (hooks specs.Hooks, err error) {
	for _, v := range s {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) <= 1 {
			return hooks, fmt.Errorf("parsing %s as hook_name:exec failed", v)
		}
		cmd := strings.Split(parts[1], " ")
		exec, err := exec.LookPath(cmd[0])
		if err != nil {
			return hooks, fmt.Errorf("looking up exec path for %s failed: %v", cmd[0], err)
		}
		hook := specs.Hook{
			Path: exec,
		}
		if len(cmd) > 1 {
			hook.Args = append(hook.Args, cmd...)
		}
		switch parts[0] {
		case "prestart":
			hooks.Prestart = append(hooks.Prestart, hook)
		case "poststart":
			hooks.Poststart = append(hooks.Poststart, hook)
		case "poststop":
			hooks.Poststop = append(hooks.Poststop, hook)
		default:
			return hooks, fmt.Errorf("%s is not a valid hook, try 'prestart', 'poststart', or 'poststop'", parts[0])
		}
	}
	return hooks, nil
}

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "riddler"
	p.Description = "A tool to convert docker inspect to the opencontainers runc spec"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.StringVar(&dockerHost, "host", "unix:///var/run/docker.sock", "Docker Daemon socket(s) to connect to")
	p.FlagSet.StringVar(&bundle, "bundle", "", "Path to the root of the bundle directory")
	p.FlagSet.Var(&hookflags, "hook", "Hooks to prefill into spec file. (ex. --hook prestart:netns)")

	p.FlagSet.IntVar(&idrootVar, "idroot", 0, "Root UID/GID for user namespaces")
	p.FlagSet.IntVar(&idlenVar, "idlen", 0, "Length of UID/GID ID space ranges for user namespaces")

	p.FlagSet.BoolVar(&force, "force", false, "force overwrite existing files")
	p.FlagSet.BoolVar(&force, "f", false, "force overwrite existing files")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		idroot = uint32(idrootVar)
		idlen = uint32(idlenVar)

		if p.FlagSet.NArg() < 1 {
			return errors.New("pass the container name or ID")
		}

		var err error
		hooks, err = hookflags.ParseHooks()
		return err
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		// On ^C, or SIGTERM handle exit.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		go func() {
			for sig := range c {
				logrus.Infof("Received %s, exiting.", sig.String())
				os.Exit(0)
			}
		}()

		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		cli, err := client.NewClient(dockerHost, "", nil, defaultHeaders)
		if err != nil {
			panic(err)
		}

		// get container info
		ctr, err := cli.ContainerInspect(ctx, args[0])
		if err != nil {
			logrus.Fatalf("inspecting container (%s) failed: %v", args[0], err)
		}

		spec, err := parse.Config(ctr, runtime.GOOS, runtime.GOARCH, defaultCapabilities(), idroot, idlen)
		if err != nil {
			logrus.Fatalf("Spec config conversion for %s failed: %v", args[0], err)
		}

		// fill in hooks, if passed through command line
		spec.Hooks = &hooks
		if err := writeConfig(spec); err != nil {
			logrus.Fatal(err)
		}

		fmt.Printf("%s has been saved.\n", specConfig)
		return nil
	}

	// Run our program.
	p.Run()
}

func checkNoFile(name string) error {
	_, err := os.Stat(name)
	if err == nil {
		return fmt.Errorf("file %s exists, remove it", name)
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeConfig(spec *specs.Spec) error {
	if bundle != "" {
		// change current working directory
		if err := os.Chdir(bundle); err != nil {
			return fmt.Errorf("change working directory to %s failed: %v", bundle, err)
		}
	}

	// make sure we don't already have files, we would not want to overwrite them
	if !force {
		if err := checkNoFile(specConfig); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(&spec, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(specConfig, data, 0666)
}

func defaultCapabilities() []string {
	return []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FSETID",
		"CAP_FOWNER",
		"CAP_MKNOD",
		"CAP_NET_RAW",
		"CAP_SETGID",
		"CAP_SETUID",
		"CAP_SETFCAP",
		"CAP_SETPCAP",
		"CAP_NET_BIND_SERVICE",
		"CAP_SYS_CHROOT",
		"CAP_KILL",
		"CAP_AUDIT_WRITE",
	}
}
