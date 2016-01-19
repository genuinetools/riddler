package parse

import (
	"path"

	"github.com/docker/engine-api/types"
	"github.com/opencontainers/specs"
)

const (
	// DefaultUserNSHostID is the default start mapped host id for userns.
	DefaultUserNSHostID = 886432
	// DefaultUserNSMapSize is the default size for the uid and gid mappings for userns.
	DefaultUserNSMapSize = 46578392
)

var (
	// DefaultMountpoints are the default mounts for the runtime.
	DefaultMountpoints = map[string]specs.Mount{
		"proc": {
			Type:    "proc",
			Source:  "proc",
			Options: nil,
		},
		"dev": {
			Type:    "tmpfs",
			Source:  "tmpfs",
			Options: []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
		},
		"devpts": {
			Type:    "devpts",
			Source:  "devpts",
			Options: []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620"},
		},
		"shm": {
			Type:    "tmpfs",
			Source:  "shm",
			Options: []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
		},
		"mqueue": {
			Type:    "mqueue",
			Source:  "mqueue",
			Options: []string{"nosuid", "noexec", "nodev"},
		},
		"sysfs": {
			Type:    "sysfs",
			Source:  "sysfs",
			Options: []string{"nosuid", "noexec", "nodev"},
		},
		"cgroup": {
			Type:    "cgroup",
			Source:  "cgroup",
			Options: []string{"nosuid", "noexec", "nodev", "relatime"},
		},
	}
)

// RuntimeConfig takes ContainerJSON and converts it into the opencontainers runtime spec.
func RuntimeConfig(c types.ContainerJSON) (*specs.LinuxRuntimeSpec, error) {
	config := &specs.LinuxRuntimeSpec{
		RuntimeSpec: specs.RuntimeSpec{
			Mounts: map[string]specs.Mount{},
			Hooks:  specs.Hooks{},
		},
		Linux: specs.LinuxRuntime{
			Namespaces: []specs.Namespace{
				{
					Type: "ipc",
				},
				{
					Type: "uts",
				},
				{
					Type: "mount",
				},
			},
			UIDMappings: []specs.IDMapping{
				{
					ContainerID: 0,
					HostID:      DefaultUserNSHostID,
					Size:        DefaultUserNSMapSize,
				},
			},
			GIDMappings: []specs.IDMapping{
				{
					ContainerID: 0,
					HostID:      DefaultUserNSHostID,
					Size:        DefaultUserNSMapSize,
				},
			},
			// TODO: add parsing of Ulimits
			Rlimits: []specs.Rlimit{
				{
					Type: "RLIMIT_NOFILE",
					Hard: uint64(1024),
					Soft: uint64(1024),
				},
			},
			Resources: &specs.Resources{
				DisableOOMKiller: c.HostConfig.Resources.OomKillDisable,
				OOMScoreAdj:      &c.HostConfig.OomScoreAdj,
				Memory: &specs.Memory{
					Limit:       uint64ptr(c.HostConfig.Resources.Memory),
					Reservation: uint64ptr(c.HostConfig.Resources.MemoryReservation),
					Swap:        uint64ptr(c.HostConfig.Resources.MemorySwap),
					Swappiness:  uint64ptr(*c.HostConfig.Resources.MemorySwappiness),
					Kernel:      uint64ptr(c.HostConfig.Resources.KernelMemory),
				},
				CPU: &specs.CPU{
					Shares: uint64ptr(c.HostConfig.Resources.CPUShares),
					Quota:  uint64ptr(c.HostConfig.Resources.CPUQuota),
					Period: uint64ptr(c.HostConfig.Resources.CPUPeriod),
					Cpus:   &c.HostConfig.Resources.CpusetCpus,
					Mems:   &c.HostConfig.Resources.CpusetMems,
				},
				Pids: &specs.Pids{
					Limit: &c.HostConfig.Resources.PidsLimit,
				},
				BlockIO: &specs.BlockIO{
					Weight: &c.HostConfig.Resources.BlkioWeight,
					// TODO: add parsing for Throttle/Weight Devices
				},
			},
			ApparmorProfile:   c.AppArmorProfile,
			RootfsPropagation: "",
		},
	}

	// check namespaces
	if !c.HostConfig.NetworkMode.IsHost() {
		config.Linux.Namespaces = append(config.Linux.Namespaces, specs.Namespace{
			Type: "network",
		})
	}
	if !c.HostConfig.PidMode.IsHost() {
		config.Linux.Namespaces = append(config.Linux.Namespaces, specs.Namespace{
			Type: "pid",
		})
	}
	if !c.HostConfig.NetworkMode.IsHost() && !c.HostConfig.PidMode.IsHost() && !c.HostConfig.Privileged {
		config.Linux.Namespaces = append(config.Linux.Namespaces, specs.Namespace{
			Type: "user",
		})
	} else {
		// reset uid and gid mappings
		config.Linux.UIDMappings = []specs.IDMapping{}
		config.Linux.GIDMappings = []specs.IDMapping{}
	}

	// get mounts
	for _, mount := range c.Mounts {
		var opt []string
		if mount.RW {
			opt = append(opt, "rw")
		}
		if mount.Mode != "" {
			opt = append(opt, mount.Mode)
		}
		opt = append(opt, "rbind")

		config.Mounts[mount.Destination] = specs.Mount{
			Type:    "bind",
			Source:  mount.Source,
			Options: opt,
		}
	}

	// if we aren't doing something crazy like mounting a default mount ourselves,
	// the we can mount it the default way
	for name, mount := range DefaultMountpoints {
		if _, ok := config.Mounts[path.Join("/", "dev", name)]; !ok {
			config.Mounts[name] = mount
		}
	}

	// fix default mounts for cgroups and devpts without user namespaces
	// see: https://github.com/opencontainers/runc/issues/225#issuecomment-136519577
	if len(config.Linux.UIDMappings) == 0 {
		if _, ok := config.Mounts["cgroup"]; ok {
			config.Mounts["cgroup"] = specs.Mount{
				Type:    "cgroup",
				Source:  "cgroup",
				Options: append(config.Mounts["cgroup"].Options, "ro"),
			}
		}
		if _, ok := config.Mounts["devpts"]; ok {
			config.Mounts["devpts"] = specs.Mount{
				Type:    "devpts",
				Source:  "devpts",
				Options: append(config.Mounts["devpts"].Options, "gid=5"),
			}
		}
	}

	// add /etc/hosts and /etc/resolv.conf if we should have networking
	if !c.HostConfig.NetworkMode.IsNone() && !c.HostConfig.NetworkMode.IsHost() {
		for _, nm := range NetworkMounts {
			config.Mounts[nm.Path] = specs.Mount{
				Type:    "bind",
				Source:  nm.Path,
				Options: []string{"rbind", "ro"},
			}
		}
	}

	// parse additional groups and add them to gid mappings
	if err := parseMappings(config, c.HostConfig); err != nil {
		return nil, err
	}

	// parse devices
	if err := parseDevices(config, c.HostConfig); err != nil {
		return nil, err
	}

	// parse security opt
	if err := parseSecurityOpt(config, c.HostConfig); err != nil {
		return nil, err
	}

	// set privileged
	if c.HostConfig.Privileged {
		if !c.HostConfig.ReadonlyRootfs {
			// clear readonly for cgroup
			config.Mounts["cgroup"] = DefaultMountpoints["cgroup"]
		}
	}

	return config, nil
}
