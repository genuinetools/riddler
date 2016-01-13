package parse

import (
	"fmt"
	"syscall"

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
				User: specs.User{
					UID: uint32(syscall.Getuid()),
					GID: uint32(syscall.Getgid()),
				},
				Args: append([]string{c.Path}, c.Args...),
				Env:  c.Config.Env,
				Cwd:  c.Config.WorkingDir,
			},
			Root: specs.Root{
				Path:     "rootfs",
				Readonly: c.HostConfig.ReadonlyRootfs,
			},
			Hostname: c.Config.Hostname,
			Mounts: []specs.MountPoint{
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
			},
		},
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

	// get mounts
	for _, mount := range c.Mounts {
		config.Mounts = append(config.Mounts, specs.MountPoint{
			Name: mount.Destination,
			Path: mount.Destination,
		})
	}

	// get the capabilities
	config.Linux.Capabilities, err = execdriver.TweakCapabilities(capabilities, c.HostConfig.CapAdd.Slice(), c.HostConfig.CapDrop.Slice())
	if err != nil {
		return nil, fmt.Errorf("setting capabilities failed: %v", err)
	}

	return config, nil
}
