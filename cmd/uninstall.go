//go:build windows

package cmd

import (
	"fmt"
	"log"

	"github.com/minjejeon/convert4share/windows"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/registry"
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
		if err := unregisterContextMenu(); err != nil {
			log.Fatalf("Failed to uninstall context menu: %v. Please ensure you are running this command as an administrator.", err)
		}
		log.Println("Context menu uninstalled successfully.")
	},
}

func unregisterContextMenu() error {
	extensions := []string{".mov", ".heic"}
	for _, ext := range extensions {
		keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\Convert4Share`, ext)
		if err := registry.DeleteKey(registry.CLASSES_ROOT, keyPath); err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("could not delete key for %s: %w", ext, err)
		}
	}
	return nil
}

func init() {
	RootCmd.AddCommand(uninstallCmd)
}
