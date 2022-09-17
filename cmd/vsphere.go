/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

// vsphereCmd represents the vsphere command
var vsphereCmd = &cobra.Command{
	Use:   "vsphere",
	Short: "Helper for vsphere actions",
	Long:  `This command parses vsphere and enable subactions for your vsphere cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		klog.Info("someData printed using InfoF")
		fmt.Println("vsphere called")
	},
}

func init() {
	vsphereCmd.PersistentFlags().String("config", "c", "Path for the vSphere configuration file.")
	vsphereCmd.PersistentFlags().String("kubeconfig", "k", "Path Kubernetes configuration file.")
	viper.BindPFlag("config", vsphereCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("kubeconfig", vsphereCmd.PersistentFlags().Lookup("kubeconfig"))
}
