// +build linux,!seccomp

package parse

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

var (
	defaultSeccompProfile specs.LinuxSeccomp
)
