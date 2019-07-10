package runner

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

func backgroundCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func kill(pid int) error {
	// Kill the whole processes tree (all children and grandchildren)
	cmd := exec.Command("taskkill", "/pid", strconv.Itoa(pid), "/T", "/F")
	b, err := cmd.CombinedOutput()
	if nil != err {
		return fmt.Errorf("%s: %s", err.Error(), string(b))
	}

	return nil
}
