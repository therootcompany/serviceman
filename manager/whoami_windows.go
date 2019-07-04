package manager

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func isPrivileged() bool {
	var sid *windows.SID

	// Although this looks scary, it is directly copied from the
	// official windows documentation. The Go API for this is a
	// direct wrap around the official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		// we don't believe this _can_ return an error with the given inputs
		// and if it does, the important info is still the false
		fmt.Fprintf(os.Stderr, "warning: Unexpected Windows UserID Error: %s\n", err)
		return false
	}

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	isAdmin, err := token.IsMember(sid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: Unexpected Windows Permission ID Error: %s\n", err)
		return false
	}

	return isAdmin || token.IsElevated()
}
