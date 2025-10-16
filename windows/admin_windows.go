//go:build windows

package windows

import (
	"log"
	"os"
	"strings"
	"syscall"

	sys "golang.org/x/sys/windows"
)

// IsElevated checks if the process is running with administrator privileges.
func IsElevated() bool {
	var sid *sys.SID
	// Authority: {0, 0, 0, 0, 0, 5} (SECURITY_NT_AUTHORITY)
	// SubAuthority[0]: 32 (SECURITY_BUILTIN_DOMAIN_RID)
	// SubAuthority[1]: 544 (DOMAIN_ALIAS_RID_ADMINS)
	err := sys.AllocateAndInitializeSid(
		&sys.SECURITY_NT_AUTHORITY,
		2,
		sys.SECURITY_BUILTIN_DOMAIN_RID,
		sys.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		log.Printf("SID Error: %s", err)
		return false
	}
	defer sys.FreeSid(sid)

	token := sys.Token(0) // current process token
	member, err := token.IsMember(sid)
	if err != nil {
		log.Printf("Token Membership Error: %s", err)
		return false
	}
	return member
}

// RunAsAdmin re-launches the application with elevated privileges using a UAC prompt.
func RunAsAdmin() {
	verb, err := syscall.UTF16PtrFromString("runas")
	if err != nil {
		log.Fatalf("Failed to create verb for ShellExecute: %v", err)
	}

	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Could not get executable path: %v", err)
	}
	exePtr, err := syscall.UTF16PtrFromString(exe)
	if err != nil {
		log.Fatalf("Failed to create exe path for ShellExecute: %v", err)
	}

	// Properly quote each argument to handle spaces in paths.
	var args []string
	for _, s := range os.Args[1:] {
		args = append(args, syscall.EscapeArg(s))
	}
	argsStr := strings.Join(args, " ")
	argsPtr, err := syscall.UTF16PtrFromString(argsStr)
	if err != nil {
		log.Fatalf("Failed to create args for ShellExecute: %v", err)
	}

	// The function expects a null pointer for an unused parameter.
	sys.ShellExecute(0, verb, exePtr, argsPtr, nil, sys.SW_SHOWNORMAL)
}
