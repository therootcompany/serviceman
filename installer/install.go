//go:generate go run -mod=vendor github.com/UnnoTed/fileb0x b0x.toml

package installer

import (
	"fmt"
	"os"
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

// Returns true if we suspect that the current user (or process) will be able
// to write to system folders, bind to privileged ports, and otherwise
// successfully run a system service.
func IsPrivileged() bool {
	return isPrivileged()
}

func WhereIs(exec string) (string, error) {
	// TODO use exec.LookPath instead
	exec = filepath.ToSlash(exec)
	if strings.Contains(exec, "/") {
		// it's a path (so we don't allow filenames with slashes)
		stat, err := os.Stat(exec)
		if nil != err {
			return "", err
		}
		if stat.IsDir() {
			return "", fmt.Errorf("'%s' is not an executable file", exec)
		}
		return filepath.Abs(exec)
	}
	return whereIs(exec)
}
