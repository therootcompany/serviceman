package manager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"git.rootprojects.org/root/go-serviceman/manager/static"
	"git.rootprojects.org/root/go-serviceman/service"
)

func install(c *service.Service) error {
	// Darwin-specific config options
	if c.PrivilegedPorts {
		if !c.System {
			return fmt.Errorf("You must use root-owned LaunchDaemons (not user-owned LaunchAgents) to use priveleged ports on OS X")
		}
	}
	plistDir := "/Library/LaunchDaemons/"
	if !c.System {
		plistDir = filepath.Join(c.Home, "Library/LaunchAgents")
	}

	// Check paths first
	err := os.MkdirAll(filepath.Dir(plistDir), 0755)
	if nil != err {
		return err
	}

	// Create service file from template
	b, err := static.ReadFile("dist/Library/LaunchDaemons/_rdns_.plist.tmpl")
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
	// TODO rdns
	plistName := c.ReverseDNS + ".plist"
	plistPath := filepath.Join(plistDir, plistName)
	if err := ioutil.WriteFile(plistPath, rw.Bytes(), 0644); err != nil {

		return fmt.Errorf("ioutil.WriteFile error: %v", err)
	}
	fmt.Printf("Installed. To start '%s' run the following:\n", c.Name)
	// TODO template config file
	if "" != c.Home {
		plistPath = strings.Replace(plistPath, c.Home, "~", 1)
	}
	sudo := ""
	if c.System {
		sudo = "sudo "
	}
	fmt.Printf("\t%slaunchctl load -w %s\n", sudo, plistPath)

	return nil
}
