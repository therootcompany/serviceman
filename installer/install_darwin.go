package installer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"git.rootprojects.org/root/go-serviceman/installer/static"
)

func install(c *Config) error {
	// Darwin-specific config options
	if c.PrivilegedPorts {
		if !c.System {
			return fmt.Errorf("You must use root-owned LaunchDaemons (not user-owned LaunchAgents) to use priveleged ports on OS X")
		}
	}
	plistDir := "/Library/LaunchDaemons/"
	if !c.System {
		plistDir = filepath.Join(c.home, "Library/LaunchAgents")
	}

	// Check paths first
	err := os.MkdirAll(filepath.Dir(plistDir), 0750)
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
	plistName := c.Name + ".plist"
	plistPath := filepath.Join(plistDir, plistName)
	if err := ioutil.WriteFile(plistPath, rw.Bytes(), 0644); err != nil {
		fmt.Println("Use 'sudo' to install as a privileged system service.")
		fmt.Println("Use '--userspace' to install as an user service.")
		return fmt.Errorf("ioutil.WriteFile error: %v", err)
	}
	fmt.Printf("Installed. To start '%s' run the following:\n", c.Name)
	// TODO template config file
	fmt.Printf("\tlaunchctl load -w %s\n", strings.Replace(plistPath, c.home, "~", 1))

	return nil
}
