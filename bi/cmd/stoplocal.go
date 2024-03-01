/*
Copyright © 2024 Elliott Clark @ Batteries Included
*/
package cmd

import (
	"bi/pkg/local"
	"fmt"

	"github.com/spf13/cobra"
)

// startLocalCmd represents the startLocal command
var stopLocalCmd = &cobra.Command{
	Use:   "stoplocal",
	Short: "Stop all locally running batteries included clusters",
	Long: `Batteries Included is built on top of
Kubernetes; this stops all kubernetes clusters locally
with just docker as a dependency.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stop local called")
		kp, err := local.NewKindClusterProvider("bi")
		cobra.CheckErr(err)

		err = kp.EnsureDeleted()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(stopLocalCmd)
}
