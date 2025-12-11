//go:build !windows

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the application from the Windows context menu (Windows only).",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Context menu uninstallation is only supported on Windows.")
	},
}

func init() {
	RootCmd.AddCommand(uninstallCmd)
}
