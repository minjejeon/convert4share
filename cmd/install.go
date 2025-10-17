package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/minjejeon/convert4share/windows"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/registry"
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
		if err := registerContextMenu(); err != nil {
			log.Fatalf("Failed to install context menu: %v. Please ensure you are running this command as an administrator.", err)
		}
		log.Println("Context menu installed successfully for .mov and .heic files.")
	},
}

func registerContextMenu() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}

	menuName := "Convert with Convert4Share"
	keyName := "Convert4Share"
	command := fmt.Sprintf(`"%s" "%%1"`, exePath) // Reverted to simple command
	extensions := []string{".mov", ".heic"}

	for _, ext := range extensions {
		// Create HKEY_CLASSES_ROOT\SystemFileAssociations\.ext\shell\Convert4Share
		keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\%s`, ext, keyName)
		key, _, err := registry.CreateKey(registry.CLASSES_ROOT, keyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("could not create shell key for %s: %w", ext, err)
		}
		// Set the menu display text and icon
		if err := key.SetStringValue("", menuName); err != nil {
			key.Close()
			return fmt.Errorf("could not set default value for %s: %w", ext, err)
		}
		if err := key.SetStringValue("Icon", `"`+exePath+`"`); err != nil {
			key.Close()
			return fmt.Errorf("could not set icon for %s: %w", ext, err)
		}

		key.Close()

		// Create HKEY_CLASSES_ROOT\SystemFileAssociations\.ext\shell\Convert4Share\command
		cmdKeyPath := fmt.Sprintf(`%s\command`, keyPath)
		cmdKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, cmdKeyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("could not create command key for %s: %w", ext, err)
		}
		if err := cmdKey.SetStringValue("", command); err != nil {
			cmdKey.Close()
			return fmt.Errorf("could not set command value for %s: %w", ext, err)
		}
		cmdKey.Close()
	}
	return nil
}

func init() {
	RootCmd.AddCommand(installCmd)
}
