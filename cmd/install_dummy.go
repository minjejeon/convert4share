//go:build wails2_legacy && !windows

package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the application to the Windows context menu (Windows only).",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Context menu installation is only supported on Windows.")
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
