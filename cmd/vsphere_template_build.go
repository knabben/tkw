package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"path/filepath"
	"tkw/pkg/config"
	"tkw/pkg/docker"
	"tkw/pkg/template"
	"tkw/pkg/vsphere"
	"tkw/pkg/windows"
)

func init() {
	templateCmd.AddCommand(templateBuildCmd)

	templateBuildCmd.PersistentFlags().String("isopath", "i", "The Windows iso file path.")
	templateBuildCmd.PersistentFlags().String("vmtoolspath", "v", "The vmware tools iso file path.")
	viper.BindPFlag("isopath", templateBuildCmd.PersistentFlags().Lookup("isopath"))
	viper.BindPFlag("vmtoolspath", templateBuildCmd.PersistentFlags().Lookup("vmtoolspath"))
}

// windowsCmd represents the vsphere command
var templateBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Windows OVA.",
	Long: `This command builds a new Windows OVA and exports on a predefined
vSphere server.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			nodeIP     string
			ctx        = context.Background()
			kubeconfig = viper.GetString("kubeconfig")
		)

		// Loading configuration on a mapper object.
		mapper, err := config.LoadConfig(viper.GetString("config"))
		config.ExplodeGraceful(err)

		// 1. Upload Windows installation ISO and VMTools ISO into the Datastore.
		msg := fmt.Sprintf("Uploading ISOs files in the cluster %s", mapper.Get(vsphere.VsphereDataStore))
		klog.Info(template.Info(msg))
		config.ExplodeGraceful(findOrUploadISOs(ctx, mapper))

		// 2. Create the windows-resource-bundle in the cluster. extract the Node IP
		client, err := windows.NewKubernetesClient(kubeconfig)
		config.ExplodeGraceful(err)

		msg = fmt.Sprintf("Creating Windows Image-Builder resources on %s default context", kubeconfig)))
		klog.Info(template.Info(msg))
		err = client.CreateWindowsResources(ctx)
		config.ExplodeGraceful(err)
		if nodeIP, err = client.GetFirstNodeIP(ctx); err != nil {
			config.ExplodeGraceful(err)
		}

		// 3. Populate Windows configuration and save on a temporary file
		klog.Info(template.Info("Generate windows.json file with parameters"))
		winSettings := windows.NewWindowsSettings(
			viper.GetString("isopath"),
			viper.GetString("vmtoolspath"),
			nodeIP,
		)

		// Manage the configuration based on mgmt parameters
		data, err := winSettings.GenerateJSONConfig(mapper)
		config.ExplodeGraceful(err)
		windowsFile, err := winSettings.SaveTempJSON(data)
		config.ExplodeGraceful(err)

		// 4. Image builder running on a docker
		klog.Info(template.Info("Running Docker container with Image builder, be ready!"))
		cli, err := docker.NewDockerClient(windowsFile)
		config.ExplodeGraceful(err)

		// Run the image-builder container.
		var containerID string
		containerID, err = cli.Run(ctx)
		config.ExplodeGraceful(err)

		// Iterate on logs and print output, monitor for errors.
		err = monitorOutput(cli, containerID)
		config.ExplodeGraceful(err)
	},
}

func findOrUploadISOs(ctx context.Context, mapper *config.Mapper) error {
	ds := mapper.Get(vsphere.VsphereDataStore)
	client, dc, err := vsphere.ConnectAndFilterDC(ctx, mapper)
	if err != nil {
		return err
	}

	obj, err := client.FindDatastore(ctx, dc.Name, ds)
	if err != nil {
		return err
	}

	// Upload vmtoolspath first.
	var vmtoolspath = viper.GetString("vmtoolspath")
	klog.Info(template.Info(fmt.Sprintf("Uploading vmtoolspath: %s", vmtoolspath)))
	err = client.Upload(ctx, vmtoolspath, filepath.Base(vmtoolspath), obj)
	if err != nil {
		return err
	}

	var isopath = viper.GetString("isopath")
	klog.Info(template.Info(fmt.Sprintf("Uploading isopath: %s", isopath)))
	err = client.Upload(ctx, isopath, filepath.Base(isopath), obj)
	if err != nil {
		return err
	}
}

func monitorOutput(cli *docker.Docker, containerID string) error {
	p, err := docker.NewProgram(cli, containerID)
	if err != nil {
		return err
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}
