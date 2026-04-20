//go:build windows

package cmd

import (
	"log/slog"
	"os"

	"github.com/minjejeon/convert4share/windows"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the application from the Windows context menu.",
	Long: `Removes the 'Convert with Convert4Share' option from the context menu.
This command must be run with administrator privileges.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !windows.IsElevated() {
			windows.RunAsAdmin()
			return
		}
		if err := windows.UnregisterContextMenu(); err != nil {
			slog.Error("Failed to uninstall context menu", "error", err)
			os.Exit(1)
		}
		slog.Info("Context menu uninstalled successfully.")
	},
}

func init() {
	RootCmd.AddCommand(uninstallCmd)
}
