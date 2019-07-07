package manager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"git.rootprojects.org/root/go-serviceman/manager/static"
	"git.rootprojects.org/root/go-serviceman/service"
)

var (
	srvLen     int
	srvExt     = ".service"
	srvSysPath = "/etc/systemd/system"

	// Not sure which of these it's supposed to be...
	// * ~/.local/share/systemd/user/watchdog.service
	// * ~/.config/systemd/user/watchdog.service
	// https://wiki.archlinux.org/index.php/Systemd/User
	// This seems to work on Ubuntu
	srvUserPath = ".config/systemd/user"
)

func init() {
	srvLen = len(srvExt)
}

func start(system bool, home string, name string) error {
	sys, user, err := getMatchingSrvs(home, name)
	if nil != err {
		return err
	}

	var service string
	if system {
		service, err = getOneSysSrv(sys, user, name)
		if nil != err {
			return err
		}
		service = filepath.Join(srvSysPath, service)
	} else {
		service, err = getOneUserSrv(home, sys, user, name)
		if nil != err {
			return err
		}
		service = filepath.Join(home, srvUserPath, service)
	}

	var cmds []Runnable
	if system {
		cmds = []Runnable{
			Runnable{
				Exec: "systemctl",
				Args: []string{"daemon-reload"},
				Must: false,
			},
			Runnable{
				Exec: "systemctl",
				Args: []string{"stop", name + ".service"},
				Must: false,
			},
			Runnable{
				Exec:     "systemctl",
				Args:     []string{"start", name + ".service"},
				Badwords: []string{"not found", "failed"},
				Must:     true,
			},
		}
	} else {
		cmds = []Runnable{
			Runnable{
				Exec: "systemctl",
				Args: []string{"--user", "daemon-reload"},
				Must: false,
			},
			Runnable{
				Exec: "systemctl",
				Args: []string{"stop", "--user", name + ".service"},
				Must: false,
			},
			Runnable{
				Exec:     "systemctl",
				Args:     []string{"start", "--user", name + ".service"},
				Badwords: []string{"not found", "failed"},
				Must:     true,
			},
		}
	}

	cmds = adjustPrivs(system, cmds)

	fmt.Println()
	for i := range cmds {
		exe := cmds[i]
		fmt.Println(exe.String())
		err := exe.Run()
		if nil != err {
			return err
		}
	}
	fmt.Println()

	return nil
}

func install(c *service.Service) error {
	// Linux-specific config options
	if c.System {
		if "" == c.User {
			c.User = "root"
		}
	}
	if "" == c.Group {
		c.Group = c.User
	}

	// Check paths first
	serviceDir := srvSysPath
	if !c.System {
		serviceDir = filepath.Join(c.Home, srvUserPath)
		err := os.MkdirAll(serviceDir, 0755)
		if nil != err {
			return err
		}
	}

	// Create service file from template
	b, err := static.ReadFile("dist/etc/systemd/system/_name_.service.tmpl")
	if err != nil {
		return err
	}
	s := string(b)
	rw := &bytes.Buffer{}
	// not sure what the template name does, but whatever
	tmpl, err := template.New("service").Parse(s)
	if err != nil {
		return err
	}
	err = tmpl.Execute(rw, c)
	if nil != err {
		return err
	}

	// Write the file out
	serviceName := c.Name + ".service"
	servicePath := filepath.Join(serviceDir, serviceName)
	if err := ioutil.WriteFile(servicePath, rw.Bytes(), 0644); err != nil {
		return fmt.Errorf("Error writing %s: %v", servicePath, err)
	}

	// TODO --no-start
	err = start(c.System, c.Home, c.Name)
	if nil != err {
		sudo := ""
		// --user-unit rather than --user --unit for older systemd
		unit := "--user-unit"
		if c.System {
			sudo = "sudo "
			unit = "--unit"
		}
		fmt.Printf("If things don't go well you should be able to get additional logging from journalctl:\n")
		fmt.Printf("\t%sjournalctl -xe %s %s.service\n", sudo, unit, c.Name)
		return err
	}

	fmt.Printf("Added and started '%s' as a systemd service.\n", c.Name)
	return nil
}
