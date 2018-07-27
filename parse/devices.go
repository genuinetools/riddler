package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/devices"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

func mergeDevices(defaultDevices []*configs.Device, userDevices []specs.LinuxDevice, userDeviceCgroup []specs.LinuxDeviceCgroup, hasTty bool) (devs []specs.LinuxDevice, dc []specs.LinuxDeviceCgroup) {
	paths := map[string]specs.LinuxDevice{}
	for _, d := range userDevices {
		paths[d.Path] = d
	}

	for _, d := range defaultDevices {
		if d.Path == "/dev/tty" && !hasTty {
			continue
		}
		if _, defined := paths[d.Path]; !defined {
			t := string(d.Type)
			devs = append(devs, specs.LinuxDevice{
				Type:     t,
				Path:     d.Path,
				Major:    d.Major,
				Minor:    d.Minor,
				FileMode: &d.FileMode,
				UID:      &d.Uid,
				GID:      &d.Gid,
			})
			dc = append(dc, specs.LinuxDeviceCgroup{
				Allow:  true,
				Type:   t,
				Major:  &d.Major,
				Minor:  &d.Minor,
				Access: d.Permissions,
			})
		}
	}
	return append(devs, userDevices...), append(dc, userDeviceCgroup...)
}

func uint64ptr(i int64) *uint64 {
	n := uint64(i)
	return &n
}

func int64ptr(i int64) *int64 {
	return &i
}

func uint64ptrfptr(i *int64) *uint64 {
        if (i == nil) {
		return nil;
        }
	return uint64ptr(*i);
}

func getDevicesFromPath(deviceMapping containertypes.DeviceMapping) (devs []specs.LinuxDevice, dc []specs.LinuxDeviceCgroup, err error) {
	device, deviceCgroup, err := deviceFromPath(deviceMapping.PathOnHost, deviceMapping.CgroupPermissions)
	// if there was no error, return the device
	if err == nil {
		device.Path = deviceMapping.PathInContainer
		return append(devs, *device), append(dc, *deviceCgroup), nil
	}

	// if the device is not a device node
	// try to see if it's a directory holding many devices
	if err == devices.ErrNotADevice {

		// check if it is a directory
		if src, e := os.Stat(deviceMapping.PathOnHost); e == nil && src.IsDir() {

			// mount the internal devices recursively
			filepath.Walk(deviceMapping.PathOnHost, func(dpath string, f os.FileInfo, e error) error {
				if e != nil {
					// If we already got an error, ignore the device
					return nil
				}
				childDevice, childDeviceCgroup, e := deviceFromPath(dpath, deviceMapping.CgroupPermissions)
				if e != nil {
					// ignore the device
					return nil
				}

				// add the device to userSpecified devices
				childDevice.Path = strings.Replace(dpath, deviceMapping.PathOnHost, deviceMapping.PathInContainer, 1)
				devs = append(devs, *childDevice)
				dc = append(dc, *childDeviceCgroup)

				return nil
			})
		}
	}

	if len(devs) > 0 {
		return devs, dc, nil
	}

	return devs, dc, fmt.Errorf("Gathering device information while adding custom device (%s) failed: %s", deviceMapping.PathOnHost, err)
}

// deviceFromPath takes the path to a device and it's cgroup_permissions(which cannot be easily queried) and looks up the information about a linux device.
func deviceFromPath(path, permissions string) (*specs.LinuxDevice, *specs.LinuxDeviceCgroup, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	var (
		devType                rune
		mode                   = fileInfo.Mode()
		fileModePermissionBits = os.FileMode.Perm(mode)
	)
	switch {
	case mode&os.ModeDevice == 0:
		return nil, nil, devices.ErrNotADevice
	case mode&os.ModeCharDevice != 0:
		fileModePermissionBits |= syscall.S_IFCHR
		devType = 'c'
	default:
		fileModePermissionBits |= syscall.S_IFBLK
		devType = 'b'
	}
	statt, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, nil, fmt.Errorf("cannot determine the device number for device %s", path)
	}
	devNumber := int(statt.Rdev)
	major := int64(unix.Major(uint64(devNumber)))
	minor := int64(unix.Minor(uint64(devNumber)))
	t := string(devType)
	dev := &specs.LinuxDevice{
		Type:     t,
		Path:     path,
		Major:    major,
		Minor:    minor,
		FileMode: &fileModePermissionBits,
		UID:      &statt.Uid,
		GID:      &statt.Gid,
	}
	dc := &specs.LinuxDeviceCgroup{
		Allow:  true,
		Type:   t,
		Major:  &major,
		Minor:  &minor,
		Access: permissions,
	}
	return dev, dc, nil
}
