//go:build wails2_legacy && windows

package cmd

import (
	"log/slog"
	"os"

	"github.com/minjejeon/convert4share/windows"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the application to the Windows context menu.",
	Long: `Adds a 'Convert with Convert4Share' option to the context menu
for .mov and .heic files. This command must be run with administrator privileges.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !windows.IsElevated() {
			windows.RunAsAdmin()
			return
		}
		if err := windows.RegisterContextMenu(); err != nil {
			slog.Error("Failed to install context menu", "error", err)
			os.Exit(1)
		}
		slog.Info("Context menu installed successfully for .mov and .heic files.")
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
