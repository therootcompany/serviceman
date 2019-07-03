package runner

import (
	"os/exec"
	"syscall"
)

func backgroundCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
