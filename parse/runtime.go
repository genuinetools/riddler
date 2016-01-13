package parse

import (
	"github.com/docker/engine-api/types"
	"github.com/opencontainers/runc/libcontainer/apparmor"
	"github.com/opencontainers/specs"
)

// RuntimeConfig takes ContainerJSON and converts it into the opencontainers runtime spec.
func RuntimeConfig(c types.ContainerJSON) (*specs.LinuxRuntimeSpec, error) {
	config := &specs.LinuxRuntimeSpec{
		RuntimeSpec: specs.RuntimeSpec{
			Mounts: map[string]specs.Mount{
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
					Options: []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
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
					Options: []string{"nosuid", "noexec", "nodev", "relatime", "ro"},
				},
			},
			Hooks: specs.Hooks{
			// TODO: do hooks
			},
		},
		Linux: specs.LinuxRuntime{
			Namespaces: []specs.Namespace{
				{
					Type: "pid",
				},
				{
					Type: "network",
				},
				{
					Type: "ipc",
				},
				{
					Type: "uts",
				},
				{
					Type: "mount",
				},
				{
					Type: "user",
				},
			},
			UIDMappings: []specs.IDMapping{
				{
					HostID:      886432,
					ContainerID: 0,
					Size:        65535,
				},
			},
			GIDMappings: []specs.IDMapping{
				{
					HostID:      886432,
					ContainerID: 0,
					Size:        65535,
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

	// parse devices
	if err := parseDevices(config, c.HostConfig); err != nil {
		return nil, err
	}

	// parse security opt
	if err := parseSecurityOpt(config, c.HostConfig); err != nil {
		return nil, err
	}

	// set default apparmor profile if possible
	if config.Linux.ApparmorProfile == "" && apparmor.IsEnabled() && !c.HostConfig.Privileged {
		config.Linux.ApparmorProfile = DefaultApparmorProfile
	}

	// TODO: set default seccomp profile if possible

	return config, nil
}
