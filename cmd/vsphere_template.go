package cmd

import (
	"github.com/spf13/cobra"
)

// templateCmd represents the vsphere command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "vSphere template actions go deeper",
	Long: `This command parses vsphere actions and enable subactions to 
be execute in the VSphere cluster.`,
}

func init() {
	vsphereCmd.AddCommand(templateCmd)
}
