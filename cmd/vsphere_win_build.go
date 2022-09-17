package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// windowsCmd represents the vsphere command
var winBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "",
	Long:  `This command parses vsphere and enable subactions for your vsphere cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		klog.Info("windows build command")
	},
}

func init() {
	winCmd.AddCommand(winBuildCmd)
}
