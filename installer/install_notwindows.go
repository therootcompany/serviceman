// +build !windows

package installer

import (
	"os/exec"
	"strings"
)

func whereIs(exe string) (string, error) {
	cmd := exec.Command("command", "-v", exe)
	out, err := cmd.Output()
	if nil != err {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
