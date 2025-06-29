package utils

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func IsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func RunAsAdmin() {
	verb := "runas"
	exe, _ := os.Executable()
	args := strings.Join(os.Args[1:], " ")
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	argsPtr, _ := syscall.UTF16PtrFromString(args)
	cwdPtr, _ := syscall.UTF16PtrFromString("")
	windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, 1)
}

func CheckAndRequestAdmin() {
	if !IsAdmin() {
		fmt.Println("Requesting administrator privileges...")
		RunAsAdmin()
		os.Exit(0)
	}
}
