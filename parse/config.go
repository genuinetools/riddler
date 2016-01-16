package parse

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
	"github.com/docker/engine-api/types"
	"github.com/opencontainers/runc/libcontainer/user"
	"github.com/opencontainers/specs"
)

const (
	// SpecVersion is the version of the opencontainers spec that will be created.
	SpecVersion = "0.2.0"

	// DefaultApparmorProfile is docker engine's default apparmor profile for containers.
	DefaultApparmorProfile = "docker-default"

	// DefaultCurrentWorkingDirectory defines the default directory the container
	// will enter in, if not already specified by the user.
	DefaultCurrentWorkingDirectory = "/"

	// DefaultTerminal is the default TERM for containers.
	DefaultTerminal = "xterm"
)

var (
	// DefaultTerminalEnv holds the minimum terminal env vars needed for an
	// interactive session.
	DefaultTerminalEnv = []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}

	// DefaultMounts are the default mounts for a container.
	DefaultMounts = []specs.MountPoint{
		{
			Name: "proc",
			Path: "/proc",
		},
		{
			Name: "dev",
			Path: "/dev",
		},
		{
			Name: "devpts",
			Path: "/dev/pts",
		},
		{
			Name: "shm",
			Path: "/dev/shm",
		},
		{
			Name: "mqueue",
			Path: "/dev/mqueue",
		},
		{
			Name: "sysfs",
			Path: "/sys",
		},
		{
			Name: "cgroup",
			Path: "/sys/fs/cgroup",
		},
	}

	// NetworkMounts are the mounts needed for default networking.
	NetworkMounts = []specs.MountPoint{
		{
			Name: "/etc/hosts",
			Path: "/etc/hosts",
		},
		{
			Name: "/etc/resolv.conf",
			Path: "/etc/resolv.conf",
		},
	}
)

// Config takes ContainerJSON and Daemon Info and converts it into the opencontainers spec.
func Config(c types.ContainerJSON, info types.Info, capabilities []string) (config *specs.LinuxSpec, err error) {
	config = &specs.LinuxSpec{
		Spec: specs.Spec{
			Version: SpecVersion,
			Platform: specs.Platform{
				OS:   info.OSType,
				Arch: info.Architecture,
			},
			Process: specs.Process{
				Terminal: c.Config.Tty,
				User:     specs.User{
				// TODO: user stuffs
				},
				Args: append([]string{c.Path}, c.Args...),
				Env:  c.Config.Env,
				Cwd:  c.Config.WorkingDir,
			},
			Root: specs.Root{
				Path:     "rootfs",
				Readonly: c.HostConfig.ReadonlyRootfs,
			},
			Mounts: []specs.MountPoint{},
		},
	}

	// make sure the current working directory is not blank
	if config.Process.Cwd == "" {
		config.Process.Cwd = DefaultCurrentWorkingDirectory
	}

	// get the user
	if c.Config.User != "" {
		u, err := user.LookupUser(c.Config.User)
		if err != nil {
			config.Spec.Process.User = specs.User{
				UID: uint32(u.Uid),
				GID: uint32(u.Gid),
			}
		} else {
			//return nil, fmt.Errorf("Looking up user (%s) failed: %v", c.Config.User, err)
			logrus.Warnf("Looking up user (%s) failed: %v", c.Config.User, err)
		}
	}
	// add the additional groups
	for _, group := range c.HostConfig.GroupAdd {
		g, err := user.LookupGroup(group)
		if err != nil {
			return nil, fmt.Errorf("Looking up group (%s) failed: %v", group, err)
		}
		config.Spec.Process.User.AdditionalGids = append(config.Spec.Process.User.AdditionalGids, uint32(g.Gid))
	}

	// get the hostname, if the hostname is the name as the first 12 characters of the id,
	// then set the hostname as the container name
	if c.ID[:12] == c.Config.Hostname {
		config.Hostname = strings.TrimPrefix(c.Name, "/")
	}

	// get mounts
	mounts := map[string]bool{}
	for _, mount := range c.Mounts {
		mounts[mount.Destination] = true
		config.Mounts = append(config.Mounts, specs.MountPoint{
			Name: mount.Destination,
			Path: mount.Destination,
		})
	}

	// add /etc/hosts and /etc/resolv.conf if we should have networking
	if c.HostConfig.NetworkMode != "none" && c.HostConfig.NetworkMode != "host" {
		DefaultMounts = append(DefaultMounts, NetworkMounts...)
	}

	// if we aren't doing something crazy like mounting a default mount ourselves,
	// the we can mount it the default way
	for _, mount := range DefaultMounts {
		if _, ok := mounts[mount.Path]; !ok {
			config.Mounts = append(config.Mounts, mount)
		}
	}

	// set privileged
	if c.HostConfig.Privileged {
		// allow all caps
		capabilities = execdriver.GetAllCapabilities()
	}

	// get the capabilities
	config.Linux.Capabilities, err = execdriver.TweakCapabilities(capabilities, c.HostConfig.CapAdd.Slice(), c.HostConfig.CapDrop.Slice())
	if err != nil {
		return nil, fmt.Errorf("setting capabilities failed: %v", err)
	}

	// add CAP_ prefix
	// TODO: this is awful
	for i, cap := range config.Linux.Capabilities {
		if !strings.HasPrefix(cap, "CAP_") {
			config.Linux.Capabilities[i] = fmt.Sprintf("CAP_%s", cap)
		}
	}

	// if we have a container that needs a terminal but no env vars, then set
	// default env vars for the terminal to function
	if config.Spec.Process.Terminal && len(config.Spec.Process.Env) <= 0 {
		config.Spec.Process.Env = DefaultTerminalEnv
	}
	if config.Spec.Process.Terminal {
		// make sure we have TERM set
		var termSet bool
		for _, env := range config.Spec.Process.Env {
			if strings.HasPrefix(env, "TERM=") {
				termSet = true
				break
			}
		}
		if !termSet {
			// set the term variable
			config.Spec.Process.Env = append(config.Spec.Process.Env, fmt.Sprintf("TERM=%s", DefaultTerminal))
		}
	}

	return config, nil
}
