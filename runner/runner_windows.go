package runner

import (
	"os/exec"
)

func init() {
	cmd, _ := exec.LookPath("cmd.exe")
	if "" != cmd {
	  shellArgs = []string{cmd, "/c"}
	}
}