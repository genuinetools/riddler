package testdata

import (
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/blkiodev"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-units"
)

var (
	defaultSwap int64 = -1

	// TestContainerJSON is a test struct for types.ContainerJSON.
	TestContainerJSON = types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:              "13b2fcc2316e03405ed50b4e67ae8710ff170d8bc307680c23d76f620b27fc46",
			Created:         "2016-01-13T02:59:39.01570113Z",
			Path:            "tor",
			Args:            []string{"-f", "/etc/tor/torrc.default"},
			Image:           "sha256:e1ff4c2f696249847bf5508cdcf73f425a816cc2963b4e67246dd050ee0d9215",
			ResolvConfPath:  "",
			HostnamePath:    "",
			HostsPath:       "",
			LogPath:         "",
			Name:            "/torproxy",
			RestartCount:    0,
			Driver:          "overlay",
			MountLabel:      "",
			ProcessLabel:    "",
			AppArmorProfile: "",
			ExecIDs:         []string{},
			HostConfig: &container.HostConfig{
				Binds:           []string{"/etc/localtime:/etc/localtime:ro"},
				ContainerIDFile: "",
				VolumeDriver:    "",
				CapAdd:          (*strslice.StrSlice)(nil),
				CapDrop:         (*strslice.StrSlice)(nil),
				GroupAdd:        []string{"audio", "video"},
				IpcMode:         "",
				OomScoreAdj:     0,
				PidMode:         "",
				Privileged:      false,
				ReadonlyRootfs:  false,
				SecurityOpt:     []string{},
				Resources: container.Resources{
					CPUShares:            0,
					CgroupParent:         "",
					BlkioWeight:          0x0,
					BlkioWeightDevice:    []*blkiodev.WeightDevice{},
					BlkioDeviceReadBps:   []*blkiodev.ThrottleDevice{},
					BlkioDeviceWriteBps:  []*blkiodev.ThrottleDevice{},
					BlkioDeviceReadIOps:  []*blkiodev.ThrottleDevice{},
					BlkioDeviceWriteIOps: []*blkiodev.ThrottleDevice{},
					CPUPeriod:            0,
					CPUQuota:             0,
					CpusetCpus:           "",
					CpusetMems:           "",
					Devices: []container.DeviceMapping{
						{
							PathOnHost:        "/dev/snd",
							PathInContainer:   "/dev/snd",
							CgroupPermissions: "rwm",
						},
						{
							PathOnHost:        "/dev/dri",
							PathInContainer:   "/dev/dri",
							CgroupPermissions: "rwm",
						},
						{
							PathOnHost:        "/dev/video0",
							PathInContainer:   "/dev/video0",
							CgroupPermissions: "rwm",
						},
						{
							PathOnHost:        "/dev/usb",
							PathInContainer:   "/dev/usb",
							CgroupPermissions: "rwm",
						},
						{
							PathOnHost:        "/dev/bus/usb",
							PathInContainer:   "/dev/bus/usb",
							CgroupPermissions: "rwm",
						},
					},
					KernelMemory:      0,
					Memory:            0,
					MemoryReservation: 0,
					MemorySwap:        int64(6442450944),
					MemorySwappiness:  &defaultSwap,
					PidsLimit:         0,
					Ulimits:           []*units.Ulimit{},
				},
			},
		},
		Mounts: []types.MountPoint{
			{
				Name:        "",
				Source:      "/etc/localtime",
				Destination: "/etc/localtime",
				Driver:      "",
				Mode:        "ro",
				RW:          false,
				Propagation: "rprivate",
			},
		},
		Config: &container.Config{
			Hostname:        "13b2fcc2316e",
			Domainname:      "",
			User:            "tor",
			AttachStdin:     false,
			AttachStdout:    false,
			AttachStderr:    false,
			PublishService:  "",
			Tty:             false,
			OpenStdin:       false,
			StdinOnce:       false,
			Env:             []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
			ArgsEscaped:     false,
			Image:           "jess/tor-proxy",
			Volumes:         map[string]struct{}(nil),
			WorkingDir:      "",
			NetworkDisabled: false,
			MacAddress:      "",
			OnBuild:         []string(nil),
			Labels:          map[string]string{},
			StopSignal:      "SIGTERM",
		},
	}

	// DaemonInfo is a test struct for types.Info.
	DaemonInfo = types.Info{
		ID:         "HID6:TAIG:SWCC:NFOG:KUDO:IOSY:LIY3:PJDB:CZJQ:IH6M:5HBP:6VUT",
		Containers: 5,
		Images:     84,
		Driver:     "overlay",
		DriverStatus: [][2]string{
			{"Backing Filesystem", "extfs"},
		},
		Plugins: types.PluginsInfo{
			Volume:        []string{"local"},
			Network:       []string{"null", "host", "bridge"},
			Authorization: []string{},
		},
		MemoryLimit:       true,
		SwapLimit:         true,
		CPUCfsPeriod:      true,
		CPUCfsQuota:       true,
		CPUShares:         true,
		CPUSet:            true,
		IPv4Forwarding:    true,
		BridgeNfIptables:  true,
		BridgeNfIP6tables: true,
		Debug:             true, NFd: 33,
		OomKillDisable:     true,
		NGoroutines:        48,
		SystemTime:         "2016-01-12T23:02:58.837259542-08:00",
		ExecutionDriver:    "native-0.2",
		LoggingDriver:      "json-file",
		NEventsListener:    0,
		KernelVersion:      "4.3.3-fsociety",
		OperatingSystem:    "Debian GNU/Linux stretch/sid",
		OSType:             "linux",
		Architecture:       "x86_64",
		IndexServerAddress: "https://index.docker.io/v1/",
		InitSha1:           "",
		InitPath:           "/usr/bin/docker",
		NCPU:               4,
		MemTotal:           7902105600,
		DockerRootDir:      "/var/lib/docker",
		HTTPProxy:          "",
		HTTPSProxy:         "",
		NoProxy:            "",
		Name:               "debian",
		Labels:             []string{},
		ExperimentalBuild:  false,
		ServerVersion:      "1.10.0-dev",
		ClusterStore:       "",
		ClusterAdvertise:   "",
	}

	// Caps are a default/test set of linux capabilities.
	Caps = []string{"CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP", "NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE"}
)
