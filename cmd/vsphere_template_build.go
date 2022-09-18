package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"tkw/pkg/config"
	"tkw/pkg/docker"
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
	Long: `This command buils the a new Windows OVA and exports on a predefined
vSphere server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 2. upload Windows iso and vmtools iso to datastore. | todo --
		// 3. create windows-resource-bundle in the cluster. and extract ips | todo --
		// 1. generate Windows json
		// 4. docker run..
		ctx := context.Background()

		// Loading configuration on a mapper object.
		mapper, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Populate Windows configuration and save on temp file
		var cfg = windows.WindowsConfiguration{}
		_, err = cfg.PopulateWindowsConfiguration(mapper, viper.GetString("isopath"), viper.GetString("vmtoolspath"))
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Create a new docker connection.
		cli, err := docker.NewDockerClient()
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Run the image-builder container.
		var containerID string
		if containerID, err = cli.Run(ctx); err != nil {
			config.ExplodeGraceful(err)
		}

		// Iterate on logs and print output, monitor for errors.
		if err := monitorOutput(cli, containerID); err != nil {
			config.ExplodeGraceful(err)
		}
	},
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
