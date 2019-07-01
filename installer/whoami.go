// +build !windows

package installer

import "os/user"

func isPrivileged() bool {
	u, err := user.Current()
	if nil != err {
		return false
	}

	// not quite, but close enough for now
	return "0" == u.Uid
}
