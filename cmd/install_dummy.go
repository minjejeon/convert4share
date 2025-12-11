//go:build !windows

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the application to the Windows context menu (Windows only).",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Context menu installation is only supported on Windows.")
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
