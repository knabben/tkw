package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"tkw/pkg/config"
	"tkw/pkg/template"
	"tkw/pkg/vsphere"
)

func init() {
	templateCmd.AddCommand(templateListCmd)
}

// templateListCmd represents the Template listing command
var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates in the vSphere Server",
	Long:  `List all templates in the vSphere Server`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Loading configuration on a mapper object.
		mapper, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Connect and filter DataCenter.
		client, dc, err := vsphere.ConnectAndFilterDC(ctx, mapper)
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Get templates from vSphere and DC.
		vms, err := client.GetImportedVirtualMachinesImages(ctx, dc.Moid)
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Iterate on VMS and print table by VM
		for i, vm := range vms {
			title := fmt.Sprintf("[%d] Template: %s", i, vm.Name)
			properties := client.GetVMMetadata(&vm)
			if err = template.RenderTable(properties, title); err != nil {
				config.ExplodeGraceful(err)
			}
		}
	},
}
