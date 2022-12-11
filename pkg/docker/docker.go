package docker

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Docker struct {
	WindowsFile string
	Client      *client.Client
}

const IMAGE_BUILDER = "projects.registry.vmware.com/tkg/image-builder:v0.1.12_vmware.2"

func NewDockerClient(windows string) (*Docker, error) {
	return connectDocker(windows)
}

func connectDocker(windows string) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Docker{Client: cli, WindowsFile: windows}, nil
}

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

// GetLogs extract the logs
func (d *Docker) GetLogs(ctx context.Context, containerID string) (string, error) {
	opts := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}
	src, err := d.Client.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return "", err
	}
	if src != nil {
		dst := &bytes.Buffer{}
		stdcopy.StdCopy(dst, dst, src)
		return dst.String(), nil
	}
	return "", nil
}
