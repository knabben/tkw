package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

type Docker struct {
	Client *client.Client
}

const IMAGE_BUILDER = "projects.registry.vmware.com/tkg/image-builder:v0.1.12_vmware.2"

func NewDockerClient() (*Docker, error) {
	return connectDocker()
}

func connectDocker() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Docker{Client: cli}, nil
}

// docker run -it --rm
// --mount type=bind,source=$(pwd)/windows.json,target=/windows.json
// --mount type=bind,source=$(pwd)/autounattend.xml,target=/home/imagebuilder/packer/ova/windows/windows-2019/autounattend.xml
// -e -e
//-e PACKER_FLAGS='-force -on-error=ask' -t
//  build-node-ova-vsphere-windows-2019

func (d *Docker) Run(ctx context.Context) (string, error) {
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
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	klog.Info(filepath.Join(pwd, "windows.json"))
	hostConfig := container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Target: "/home/imagebuilder/windows.json",
				Source: filepath.Join(pwd, "windows.json"),
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
func (d *Docker) GetLogs(ctx context.Context, containerID string) ([]byte, error) {
	opts := types.ContainerLogsOptions{ShowStdout: true}
	out, err := d.Client.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return []byte{}, err
	}

	if out != nil {
		return io.ReadAll(out)
	}
	return []byte{}, nil
}
