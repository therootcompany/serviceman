package manager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"git.rootprojects.org/root/serviceman/manager/static"
	"git.rootprojects.org/root/serviceman/service"
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

func start(conf *service.Service) error {
	system := conf.System
	home := conf.Home
	name := conf.ReverseDNS

	_, err := getService(system, home, name)
	if nil != err {
		return err
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
				Args:     []string{"enable", name + ".service"},
				Badwords: []string{"not found", "failed"},
				Must:     true,
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

	typ := "USER MODE"
	if system {
		typ = "SYSTEM"
	}
	fmt.Printf("Starting systemd %s service unit...\n\n", typ)
	for i := range cmds {
		exe := cmds[i]
		fmt.Println("\t" + exe.String())
		err := exe.Run()
		if nil != err {
			return err
		}
	}
	fmt.Println()

	return nil
}

func stop(conf *service.Service) error {
	system := conf.System
	home := conf.Home
	name := conf.ReverseDNS

	_, err := getService(system, home, name)
	if nil != err {
		return err
	}

	var cmds []Runnable
	badwords := []string{"Failed to stop"}
	if system {
		cmds = []Runnable{
			Runnable{
				Exec:     "systemctl",
				Args:     []string{"stop", name + ".service"},
				Must:     true,
				Badwords: badwords,
			},
		}
	} else {
		cmds = []Runnable{
			Runnable{
				Exec:     "systemctl",
				Args:     []string{"stop", "--user", name + ".service"},
				Must:     true,
				Badwords: badwords,
			},
		}
	}

	cmds = adjustPrivs(system, cmds)

	fmt.Println()
	typ := "USER MODE"
	if system {
		typ = "SYSTEM"
	}
	fmt.Printf("Stopping systemd %s service...\n", typ)
	for i := range cmds {
		exe := cmds[i]
		fmt.Println("\t" + exe.String())
		err := exe.Run()
		if nil != err {
			return err
		}
	}
	fmt.Println()

	return nil
}

// Render will create a systemd .service file using the simple internal template
func Render(c *service.Service) ([]byte, error) {
	defaultUserGroup(c)

	// Create service file from template
	b, err := static.ReadFile("dist/etc/systemd/system/_name_.service.tmpl")
	if err != nil {
		return nil, err
	}
	s := string(b)
	rw := &bytes.Buffer{}
	// not sure what the template name does, but whatever
	tmpl, err := template.New("service").Parse(s)
	if err != nil {
		return nil, err
	}
	err = tmpl.Execute(rw, c)
	if nil != err {
		return nil, err
	}

	return rw.Bytes(), nil
}

func install(c *service.Service) (string, error) {
	defaultUserGroup(c)

	// Check paths first
	serviceDir := srvSysPath
	if !c.System {
		serviceDir = filepath.Join(c.Home, srvUserPath)
		err := os.MkdirAll(serviceDir, 0755)
		if nil != err {
			return "", err
		}
	}

	b, err := Render(c)
	if nil != err {
		return "", err
	}

	// Write the file out
	serviceName := c.Name + ".service"
	servicePath := filepath.Join(serviceDir, serviceName)
	if err := ioutil.WriteFile(servicePath, b, 0644); err != nil {
		return "", fmt.Errorf("Error writing %s: %v", servicePath, err)
	}

	// TODO --no-start
	err = start(c)
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
		return "", err
	}

	return "systemd", nil
}

func defaultUserGroup(c *service.Service) {
	// Linux-specific config options
	if c.System {
		if "" == c.User {
			c.User = "root"
		}
	}
	if "" == c.Group {
		c.Group = c.User
	}
}
