package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"strings"
	"tkw/pkg/config"
	"tkw/pkg/template"
	"tkw/pkg/vsphere"
)

const (
	VsphereTlsThumbprint = "VSPHERE_TLS_THUMBPRINT"
	VsphereUsername      = "VSPHERE_USERNAME"
	VspherePassword      = "VSPHERE_PASSWORD"
	VsphereServer        = "VSPHERE_SERVER"
	VsphereDataCenter    = "VSPHERE_DATACENTER"
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
		mapper, err := loadConfig()
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Connecting vSphere server with configuration.
		var client vsphere.Client
		message := fmt.Sprintf("Connecting to the vSphere server... %s", mapper.Get(VsphereServer))
		tmpStyle := template.BaseStyle.Copy()
		fmt.Println(tmpStyle.Padding(3, 2, 3, 2).Render(message))
		if client, err = connectVCAndLogin(mapper); err != nil {
			config.ExplodeGraceful(err)
		}

		// Search for existent Datacenter.
		dcMOID, err := mapper.FilterDatacenter(ctx, client, mapper.Get(VsphereDataCenter))
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Get templates from vSphere and DC.
		vms, err := client.GetImportedVirtualMachinesImages(ctx, dcMOID)
		if err != nil {
			config.ExplodeGraceful(err)
		}

		// Iterate on VMS and print table by VM
		for _, vm := range vms {
			title := fmt.Sprintf("Template: %s", vm.Name)
			properties := client.GetVMMetadata(&vm)
			if err = template.RenderTable(properties, title); err != nil {
				config.ExplodeGraceful(err)
			}
		}
	},
}

// connectVCAndLogin returns the logged client.
func connectVCAndLogin(mapper *template.Mapper) (vsphere.Client, error) {
	var ctx = context.Background()
	host := mapper.Get(VsphereServer)
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}

	vc, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	vc.Path = "/sdk"
	vcClient, err := vsphere.NewClient(vc, mapper.Get(VsphereTlsThumbprint), false)
	if err != nil {
		return nil, err
	}
	_, err = vcClient.Login(ctx, mapper.Get(VsphereUsername), mapper.Get(VspherePassword))
	if err != nil {
		return nil, err
	}

	return vcClient, nil
}

// loadConfig returns the mapper object from config
func loadConfig() (mapper *template.Mapper, err error) {
	viperConfig := viper.GetString("config")
	if mapper, err = config.ConvertConfigIntoMap(viperConfig); err != nil {
		return nil, err
	}
	return
}
