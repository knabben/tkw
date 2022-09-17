package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// windowsCmd represents the vsphere command
var winCmd = &cobra.Command{
	Use:   "windows",
	Short: "",
	Long:  `This command parses vsphere and enable subactions for your vsphere cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		klog.Info("windows command")
	},
}

func init() {
	vsphereCmd.AddCommand(winCmd)
}
