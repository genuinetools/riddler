// +build linux,!seccomp

package parse

import "github.com/opencontainers/specs/specs-go"

var (
	defaultSeccompProfile specs.Seccomp
)
