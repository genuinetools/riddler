package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	native "github.com/docker/docker/daemon/execdriver/native/template"
	"github.com/docker/docker/pkg/platform"
	"github.com/docker/engine-api/client"
	"github.com/jfrazelle/riddler/parse"
	specs "github.com/opencontainers/specs/specs-go"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `      _     _     _ _
 _ __(_) __| | __| | | ___ _ __
| '__| |/ _` + "`" + ` |/ _` + "`" + ` | |/ _ \ '__|
| |  | | (_| | (_| | |  __/ |
|_|  |_|\__,_|\__,_|_|\___|_|
 docker inspect to opencontainers runc spec generator.
 Version: %s

`
	// VERSION is the binary version.
	VERSION = "v0.1.0"

	specConfig = "config.json"
)

var (
	arg        string
	bundle     string
	dockerHost string
	hooks      specs.Hooks
	hookflags  stringSlice
	force      bool
	idroot     uint32
	idlen      uint32

	debug   bool
	version bool
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
			hook.Args = cmd[:1]
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

func init() {
	var idrootVar, idlenVar int
	// parse flags
	flag.StringVar(&dockerHost, "host", "unix:///var/run/docker.sock", "Docker Daemon socket(s) to connect to")
	flag.StringVar(&bundle, "bundle", "", "Path to the root of the bundle directory")
	flag.Var(&hookflags, "hook", "Hooks to prefill into spec file. (ex. --hook prestart:netns)")

	flag.IntVar(&idrootVar, "idroot", 0, "Root UID/GID for user namespaces")
	flag.IntVar(&idlenVar, "idlen", 0, "Length of UID/GID ID space ranges for user namespaces")

	flag.BoolVar(&force, "force", false, "force overwrite existing files")
	flag.BoolVar(&force, "f", false, "force overwrite existing files")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()
	idroot = uint32(idrootVar)
	idlen = uint32(idlenVar)

	if version {
		fmt.Printf("%s\n", VERSION)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		usageAndExit("Pass the container name or ID.", 1)
	}

	// parse the arg
	arg = flag.Args()[0]
	if arg == "help" {
		usageAndExit("", 0)
	}

	if arg == "version" {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var err error
	hooks, err = hookflags.ParseHooks()
	if err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient(dockerHost, "", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	// get container info
	c, err := cli.ContainerInspect(context.Background(), arg)
	if err != nil {
		logrus.Fatalf("inspecting container (%s) failed: %v", arg, err)
	}

	t := native.New()
	spec, err := parse.Config(c, platform.OSType, platform.Architecture, t.Capabilities, idroot, idlen)
	if err != nil {
		logrus.Fatalf("Spec config conversion for %s failed: %v", arg, err)
	}

	// fill in hooks, if passed through command line
	spec.Hooks = hooks
	if err := writeConfig(spec); err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("%s has been saved.\n", specConfig)
}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}

func checkNoFile(name string) error {
	_, err := os.Stat(name)
	if err == nil {
		return fmt.Errorf("File %s exists. Remove it first", name)
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
	if err := ioutil.WriteFile(specConfig, data, 0666); err != nil {
		return err
	}

	return nil
}
