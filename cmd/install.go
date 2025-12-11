//go:build windows

package cmd

import (
	"log"

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
			log.Fatalf("Failed to install context menu: %v. Please ensure you are running this command as an administrator.", err)
		}
		log.Println("Context menu installed successfully for .mov and .heic files.")
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
