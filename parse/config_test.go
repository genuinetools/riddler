package parse

import (
	"reflect"
	"testing"

	"github.com/jfrazelle/riddler/parse/testdata"
	"github.com/opencontainers/specs"
)

func TestConfig(t *testing.T) {
	var expected = &specs.LinuxSpec{
		Spec: specs.Spec{
			Version: "0.2.0",
			Platform: specs.Platform{
				OS:   "linux",
				Arch: "x86_64",
			},
			Process: specs.Process{
				Terminal: false,
				User: specs.User{
					UID:            0x0,
					GID:            0x0,
					AdditionalGids: []uint32{0x1d, 0x2c},
				},
				Args: []string{"tor", "-f", "/etc/tor/torrc.default"},
				Env:  []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
				Cwd:  "",
			},
			Root: specs.Root{
				Path:     "rootfs",
				Readonly: false,
			},
			Hostname: "13b2fcc2316e",
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
				{
					Name: "/etc/localtime",
					Path: "/etc/localtime",
				},
				{
					Name: "/etc/hosts",
					Path: "/etc/hosts",
				},
				{
					Name: "/etc/resolv.conf",
					Path: "/etc/resolv.conf",
				},
			},
		},
		Linux: specs.Linux{
			Capabilities: []string{"CAP_CHOWN", "CAP_DAC_OVERRIDE", "CAP_FSETID", "CAP_FOWNER", "CAP_MKNOD", "CAP_NET_RAW", "CAP_SETGID", "CAP_SETUID", "CAP_SETFCAP", "CAP_SETPCAP", "CAP_NET_BIND_SERVICE", "CAP_SYS_CHROOT", "CAP_KILL", "CAP_AUDIT_WRITE"},
		},
	}

	config, err := Config(testdata.TestContainerJSON, testdata.DaemonInfo, testdata.Caps)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config, expected) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected, config)
	}
}
