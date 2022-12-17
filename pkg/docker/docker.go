package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type Docker struct {
	WindowsFile string
	Client      *client.Client
}

const IMAGE_BUILDER = "projects.registry.vmware.com/tkg/image-builder:v0.1.12_vmware.2"

func (d *Docker) Run(ctx context.Context) (string, error) {
	// Configuration with image-builder command and debugging flags.
	config := container.Config{
		Image: IMAGE_BUILDER,
		Cmd:   []string{"build-node-ova-vsphere-windows-2019"},
		Env: []string{
			"PACKER_LOG=1",
			"PACKER_VAR_FILES=windows.json",
			"IB_OVFTOOL=1",
			"IB_OVFTOOL_ARGS='--skipManifestCheck'",
		},
	}
	// Volume mount the Windows json custom configuration.
	// The user can be able to change anything before using in the future.
	hostConfig := container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Target: "/home/imagebuilder/windows.json",
				Source: d.WindowsFile,
			},
		},
	}
	resp, err := d.Client.ContainerCreate(ctx, &config, &hostConfig, nil, nil, "image-builder-windows")
	if err != nil {
		return "", err
	}

	containerID := resp.ID
	if err = d.Client.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return containerID, nil
}
