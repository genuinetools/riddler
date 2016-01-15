package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	native "github.com/docker/docker/daemon/execdriver/native/template"
	"github.com/docker/engine-api/client"
	"github.com/jfrazelle/riddler/parse"
	"github.com/opencontainers/specs"
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

	specConfig    = "config.json"
	runtimeConfig = "runtime.json"
)

var (
	arg        string
	bundle     string
	dockerHost string
	force      bool

	debug   bool
	version bool
)

func init() {
	// parse flags
	flag.StringVar(&dockerHost, "host", "unix:///var/run/docker.sock", "Docker Daemon socket(s) to connect to")
	flag.StringVar(&bundle, "bundle", "", "Path to the root of the bundle directory")

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

	if flag.NArg() < 1 {
		usageAndExit("Pass the container name or ID.", 1)
	}

	// parse the arg
	arg = flag.Args()[0]

	if arg == "help" {
		usageAndExit("", 0)
	}

	if version || arg == "version" {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient(dockerHost, "", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	// get container info
	c, err := cli.ContainerInspect(arg)
	if err != nil {
		logrus.Fatalf("inspecting container (%s) failed: %v", arg, err)
	}

	// get daemon info
	info, err := cli.Info()
	if err != nil {
		logrus.Fatalf("getting daemon info failed: %v", err)
	}

	t := native.New()
	spec, err := parse.Config(c, info, t.Capabilities)
	if err != nil {
		logrus.Fatalf("Spec config conversion for %s failed: %v", arg, err)
	}

	rspec, err := parse.RuntimeConfig(c)
	if err != nil {
		logrus.Fatalf("Spec runtime config conversion for %s failed: %v", arg, err)
	}

	if err := writeConfigs(spec, rspec); err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("%s and %s have been saved.", specConfig, runtimeConfig)
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

func writeConfigs(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec) error {
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
		if err := checkNoFile(runtimeConfig); err != nil {
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

	rdata, err := json.MarshalIndent(&rspec, "", "    ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(runtimeConfig, rdata, 0666); err != nil {
		return err
	}

	return nil
}
