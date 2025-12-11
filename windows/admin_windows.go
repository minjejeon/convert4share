//go:build windows

package windows

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// RunAsAdmin relaunches the application with administrator privileges.
func RunAsAdmin() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return
	}

	verb := "runas"
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err = windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, showCmd)
	if err != nil {
		fmt.Println("Error running as admin:", err)
	}
}

// RunCommandAsAdmin runs the current executable with the specified arguments as administrator.
func RunCommandAsAdmin(args string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	verb := "runas"
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	return windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, showCmd)
}

// IsElevated checks if the current process is running with administrator privileges.
func IsElevated() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	return token.IsElevated()
}
