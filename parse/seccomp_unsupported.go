// +build linux,!seccomp

package parse

import "github.com/opencontainers/specs"

var (
	defaultSeccompProfile specs.Seccomp
)
