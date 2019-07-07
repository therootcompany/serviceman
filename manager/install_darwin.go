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

const (
	srvExt      = ".plist"
	srvSysPath  = "/Library/LaunchDaemons"
	srvUserPath = "Library/LaunchAgents"
)

var srvLen int

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

	cmds := []Runnable{
		Runnable{
			Exec: "launchctl",
			Args: []string{"unload", "-w", service},
			Must: false,
		},
		Runnable{
			Exec: "launchctl",
			Args: []string{"load", "-w", service},
			Must: true,
		},
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
	// Darwin-specific config options
	if c.PrivilegedPorts {
		if !c.System {
			return fmt.Errorf("You must use root-owned LaunchDaemons (not user-owned LaunchAgents) to use priveleged ports on OS X")
		}
	}
	plistDir := srvSysPath
	if !c.System {
		plistDir = filepath.Join(c.Home, srvUserPath)
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
		return fmt.Errorf("Error writing %s: %v", plistPath, err)
	}

	// TODO --no-start
	err = start(c.System, c.Home, c.ReverseDNS)
	if nil != err {
		fmt.Printf("If things don't go well you should be able to get additional logging from launchctl:\n")
		fmt.Printf("\tsudo launchctl log level debug\n")
		fmt.Printf("\ttail -f /var/log/system.log\n")
		return err
	}

	fmt.Printf("Added and started '%s' as a launchctl service.\n", c.Name)
	return nil
}
