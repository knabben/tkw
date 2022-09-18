/* Copyright © 2022 Amim Knabben */
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"path/filepath"
)

// vsphereCmd represents the vsphere command
var vsphereCmd = &cobra.Command{
	Use:   "vsphere",
	Short: "Helper for vsphere actions",
	Long:  `This command parses vsphere and enable subactions for your vsphere cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		klog.Info("Use a subcommand instead, nothing to see here.")
	},
}

func init() {
	kubeLocal := ""
	if home := homedir.HomeDir(); home != "" {
		kubeLocal = filepath.Join(home, ".kube", "config")
	}
	vsphereCmd.PersistentFlags().String("config", "c", "Path for the vSphere configuration file.")
	vsphereCmd.PersistentFlags().StringP("kubeconfig", "k", kubeLocal, "Path Kubernetes configuration file.")
	viper.BindPFlag("config", vsphereCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("kubeconfig", vsphereCmd.PersistentFlags().Lookup("kubeconfig"))
}
