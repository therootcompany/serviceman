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
	serviceDir := "/etc/systemd/system/"

	// Check paths first
	serviceName := c.Name + ".service"
	if !c.System {
		// Not sure which of these it's supposed to be...
		// * ~/.local/share/systemd/user/watchdog.service
		// * ~/.config/systemd/user/watchdog.service
		// https://wiki.archlinux.org/index.php/Systemd/User
		serviceDir = filepath.Join(c.Home, ".local/share/systemd/user")
		err := os.MkdirAll(filepath.Dir(serviceDir), 0755)
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
	servicePath := filepath.Join(serviceDir, serviceName)
	if err := ioutil.WriteFile(servicePath, rw.Bytes(), 0644); err != nil {
		return fmt.Errorf("ioutil.WriteFile error: %v", err)
	}

	// TODO template this as well?
	userspace := ""
	sudo := "sudo "
	if !c.System {
		userspace = "--user "
		sudo = ""
	}
	fmt.Printf("System service installed as '%s'.\n", servicePath)
	fmt.Printf("Run the following to start '%s':\n", c.Name)
	fmt.Printf("\t" + sudo + "systemctl " + userspace + "daemon-reload\n")
	fmt.Printf("\t"+sudo+"systemctl "+userspace+"restart %s.service\n", c.Name)
	fmt.Printf("\t"+sudo+"journalctl "+userspace+"-xefu %s\n", c.Name)
	return nil
}
