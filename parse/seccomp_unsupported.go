// +build linux,!seccomp

package native

import "github.com/opencontainers/specs"

var (
	defaultSeccompProfile specs.Seccomp
)
