package parse

import (
	"encoding/json"
	"fmt"
	"strings"

	containertypes "github.com/docker/engine-api/types/container"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/label"
	"github.com/opencontainers/specs"
)

func parseDevices(config *specs.LinuxRuntimeSpec, hc *containertypes.HostConfig) error {
	var userSpecifiedDevices []specs.Device
	for _, deviceMapping := range hc.Devices {
		devs, err := getDevicesFromPath(deviceMapping)
		if err != nil {
			return err
		}

		userSpecifiedDevices = append(userSpecifiedDevices, devs...)
	}

	config.Linux.Devices = mergeDevices(configs.DefaultAllowedDevices, userSpecifiedDevices)

	return nil
}

func parseSecurityOpt(config *specs.LinuxRuntimeSpec, hc *containertypes.HostConfig) error {
	var (
		labelOpts []string
		err       error
	)

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
			var seccomp specs.Seccomp
			if err := json.Unmarshal([]byte(con[1]), &seccomp); err != nil {
				return fmt.Errorf("parsing seccomp profile failed: %v", err)
			}
			config.Linux.Seccomp = seccomp
		default:
			return fmt.Errorf("invalid security-opt: %q", opt)
		}
	}

	config.Linux.SelinuxProcessLabel, _, err = label.InitLabels(labelOpts)
	return err
}
