package parse

import (
	"encoding/json"
	"fmt"
	"strings"

	containertypes "github.com/docker/engine-api/types/container"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/devices"
	"github.com/opencontainers/runc/libcontainer/label"
	"github.com/opencontainers/runc/libcontainer/user"
	"github.com/opencontainers/specs"
)

func parseDevices(config *specs.LinuxRuntimeSpec, hc *containertypes.HostConfig) error {
	if hc.Privileged {
		hostDevices, err := devices.HostDevices()
		if err != nil {
			return fmt.Errorf("getting host devices for privileged mode failed: %v", err)
		}
		for _, d := range hostDevices {
			config.Linux.Devices = append(config.Linux.Devices, specs.Device{
				Type:        d.Type,
				Path:        d.Path,
				Major:       d.Major,
				Minor:       d.Minor,
				Permissions: d.Permissions,
				FileMode:    d.FileMode,
				UID:         d.Uid,
				GID:         d.Gid,
			})
		}

		return nil
	}

	var userSpecifiedDevices []specs.Device
	for _, deviceMapping := range hc.Devices {
		devs, err := getDevicesFromPath(deviceMapping)
		if err != nil {
			return err
		}

		userSpecifiedDevices = append(userSpecifiedDevices, devs...)
	}

	config.Linux.Devices = mergeDevices(configs.DefaultSimpleDevices, userSpecifiedDevices)

	return nil
}

func parseMappings(config *specs.LinuxRuntimeSpec, hc *containertypes.HostConfig) error {
	for _, g := range hc.GroupAdd {
		var newGidMap = []specs.IDMapping{}
		group, err := user.LookupGroup(g)
		if err != nil {
			return fmt.Errorf("looking up group %s failed: %v", g, err)
		}
		gid := uint32(group.Gid)

		for _, gm := range config.Linux.GIDMappings {
			if (gm.ContainerID+gm.Size) >= gid && gm.ContainerID <= gid {
				size := gm.Size
				// split the config.Linux.GIDMappingsping up so we can map to the additional group
				gm.Size = gid - gm.ContainerID - 1

				// add the gid maps for the additional groups
				newGidMap = append(newGidMap, specs.IDMapping{
					ContainerID: gid,
					HostID:      gid,
					Size:        1,
				})

				// add the other side of the split
				newGidMap = append(newGidMap, specs.IDMapping{
					ContainerID: gid + 1,
					HostID:      gm.HostID + gid - 1,
					Size:        size - gid - 1,
				})
			}
			// add back original gm
			newGidMap = append(newGidMap, gm)
		}
		config.Linux.GIDMappings = newGidMap
	}

	return nil
}

func parseSecurityOpt(config *specs.LinuxRuntimeSpec, hc *containertypes.HostConfig) error {
	var (
		labelOpts []string
		err       error
	)

	var customSeccompProfile bool
	for _, opt := range hc.SecurityOpt {
		con := strings.SplitN(opt, ":", 2)
		if len(con) == 1 {
			return fmt.Errorf("invalid --security-opt: %q", opt)
		}
		switch con[0] {
		case "label":
			labelOpts = append(labelOpts, con[1])
		case "apparmor":
			config.Linux.ApparmorProfile = con[1]
		case "seccomp":
			customSeccompProfile = true
			var seccomp specs.Seccomp
			if err := json.Unmarshal([]byte(con[1]), &seccomp); err != nil {
				return fmt.Errorf("parsing seccomp profile failed: %v", err)
			}
			config.Linux.Seccomp = seccomp
		default:
			return fmt.Errorf("invalid security-opt: %q", opt)
		}
	}

	// set default apparmor profile if possible
	if config.Linux.ApparmorProfile == "" && !hc.Privileged {
		config.Linux.ApparmorProfile = DefaultApparmorProfile
	}
	if config.Linux.ApparmorProfile == "" && hc.Privileged {
		config.Linux.ApparmorProfile = "unconfined"
	}

	// set default seccomp profile if the user did not pass a custom profile
	if !customSeccompProfile && !hc.Privileged {
		config.Linux.Seccomp = defaultSeccompProfile
	}

	config.Linux.SelinuxProcessLabel, _, err = label.InitLabels(labelOpts)
	return err
}
