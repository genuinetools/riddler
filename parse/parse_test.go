package parse

import (
	"reflect"
	"testing"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/opencontainers/runc/libcontainer/user"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type mappings struct {
	gidMap           []specs.LinuxIDMapping
	additionalGroups []string
	expected         []specs.LinuxIDMapping
}

func TestParseMappings(t *testing.T) {
	groupIDs := map[string]uint32{}

	groups := []string{"audio", "video"}

	for _, g := range groups {
		group, err := user.LookupGroup(g)
		if err != nil {
			t.Fatalf("looking up group %s failed: %v", g, err)
		}
		groupIDs[g] = uint32(group.Gid)
	}

	tests := []mappings{
		{
			gidMap: []specs.LinuxIDMapping{
				{
					ContainerID: 0,
					HostID:      87645,
					Size:        46578392,
				},
			},
			additionalGroups: []string{"audio"},
			expected: []specs.LinuxIDMapping{
				{
					ContainerID: groupIDs["audio"],
					HostID:      groupIDs["audio"],
					Size:        1,
				},
				{
					ContainerID: groupIDs["audio"] + 1,
					HostID:      87645 + groupIDs["audio"] - 1,
					Size:        46578392 - groupIDs["audio"] - 1,
				},
				{
					ContainerID: 0,
					HostID:      87645,
					Size:        groupIDs["audio"] - 1,
				},
			},
		},
		{
			gidMap: []specs.LinuxIDMapping{
				{
					ContainerID: 0,
					HostID:      87645,
					Size:        46578392,
				},
			},
			additionalGroups: []string{"audio", "video"},
			expected: []specs.LinuxIDMapping{
				{
					ContainerID: groupIDs["audio"],
					HostID:      groupIDs["audio"],
					Size:        1,
				},
				{
					ContainerID: groupIDs["video"],
					HostID:      groupIDs["video"],
					Size:        1,
				},
				{
					ContainerID: groupIDs["video"] + 1,
					HostID:      (87645 + groupIDs["audio"] - 1) + groupIDs["video"] - 1,
					Size:        46578392 - groupIDs["video"] - groupIDs["audio"] - 2,
				},
				{
					ContainerID: groupIDs["audio"] + 1,
					HostID:      87645 + groupIDs["audio"] - 1,
					Size:        groupIDs["video"] - groupIDs["audio"] - 2,
				},
				{
					ContainerID: 0,
					HostID:      87645,
					Size:        groupIDs["audio"] - 1,
				},
			},
		},
	}

	for _, test := range tests {
		// make config
		config := &specs.Spec{
			Linux: &specs.Linux{
				GIDMappings: test.gidMap,
			},
		}
		hostConfig := &containertypes.HostConfig{
			GroupAdd: test.additionalGroups,
		}

		if err := parseMappings(config, hostConfig); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(test.expected, config.Linux.GIDMappings) {
			t.Fatalf("expected:\n%#v\ngot:\n%#v", test.expected, config.Linux.GIDMappings)
		}
	}
}
