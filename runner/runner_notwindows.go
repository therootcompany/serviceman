// +build !windows

package runner

import (
	"os"
	"os/exec"
)

func backgroundCmd(cmd *exec.Cmd) {
}

func kill(pid int) error {
	p, err := os.FindProcess(pid)
	// already died
	if nil != err {
		return nil
	}
	return p.Kill()
}
