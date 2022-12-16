package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

// templateCmd represents the vsphere command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "vSphere template actions go deeper",
	Long: `This command parses vsphere actions and enable subactions to 
be execute in the VSphere cluster.`,
}

func init() {
	kubeLocal := ""
	if home := homedir.HomeDir(); home != "" {
		kubeLocal = filepath.Join(home, ".kube", "config")
	}

	templateCmd.PersistentFlags().String("config", "c", "Path for the vSphere configuration file.")
	templateCmd.PersistentFlags().StringP("kubeconfig", "k", kubeLocal, "Path Kubernetes configuration file.")
	viper.BindPFlag("config", templateCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("kubeconfig", templateCmd.PersistentFlags().Lookup("kubeconfig"))
}
