//go:generate go run -mod=vendor github.com/UnnoTed/fileb0x b0x.toml

package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.rootprojects.org/root/go-serviceman/service"
)

// Install will do a best-effort attempt to install a start-on-startup
// user or system service via systemd, launchd, or reg.exe
func Install(c *service.Service) error {
	if "" == c.Exec {
		c.Exec = c.Name
	}

	if !c.System {
		home, err := os.UserHomeDir()
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unrecoverable Error: %s", err)
			os.Exit(4)
			return err
		} else {
			c.Home = home
		}
	}

	err := install(c)
	if nil != err {
		return err
	}

	err = os.MkdirAll(c.Logdir, 0755)
	if nil != err {
		return err
	}

	return nil
}

func Start(conf *service.Service) error {
	return start(conf)
}

func Stop(conf *service.Service) error {
	return stop(conf)
}

// IsPrivileged returns true if we suspect that the current user (or process) will be able
// to write to system folders, bind to privileged ports, and otherwise
// successfully run a system service.
func IsPrivileged() bool {
	return isPrivileged()
}

// WhereIs uses exec.LookPath to return an absolute filepath with forward slashes
func WhereIs(exe string) (string, error) {
	exepath, err := exec.LookPath(exe)
	if nil != err {
		return "", err
	}
	return filepath.Abs(filepath.ToSlash(exepath))
}

type ErrDaemonize struct {
	DaemonArgs []string
	error      string
}

func (e *ErrDaemonize) Error() string {
	return e.error + "\nYou need to switch on ErrDaemonize, and use .DaemonArgs, which would run this:" + strings.Join(e.DaemonArgs, " ")
}
