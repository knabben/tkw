package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// windowsCmd represents the vsphere command
var templateUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "",
	Long:  `This command parses vsphere and enable subactions for your vsphere cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		klog.Info("vsphere template upload command")
	},
}

func init() {
	templateCmd.AddCommand(templateUploadCmd)
}
