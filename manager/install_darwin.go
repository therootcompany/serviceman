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

func start(conf *service.Service) error {
	system := conf.System
	home := conf.Home
	rdns := conf.ReverseDNS

	service, err := getService(system, home, rdns)
	if nil != err {
		return err
	}

	cmds := []Runnable{
		Runnable{
			Exec: "launchctl",
			Args: []string{"unload", "-w", service},
			Must: false,
		},
		Runnable{
			Exec:     "launchctl",
			Args:     []string{"load", "-w", service},
			Must:     true,
			Badwords: []string{"No such file or directory", "service already loaded"},
		},
	}

	cmds = adjustPrivs(system, cmds)

	typ := "USER"
	if system {
		typ = "SYSTEM"
	}
	fmt.Printf("Starting launchd %s service...\n\n", typ)
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
	rdns := conf.ReverseDNS

	service, err := getService(system, home, rdns)
	if nil != err {
		return err
	}

	cmds := []Runnable{
		Runnable{
			Exec:     "launchctl",
			Args:     []string{"unload", service},
			Must:     false,
			Badwords: []string{"No such file or directory", "Cound not find specified service"},
		},
	}

	cmds = adjustPrivs(system, cmds)

	fmt.Println()
	typ := "USER"
	if system {
		typ = "SYSTEM"
	}
	fmt.Printf("Stopping launchd %s service...\n", typ)
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

func Render(c *service.Service) ([]byte, error) {
	// Create service file from template
	b, err := static.ReadFile("dist/Library/LaunchDaemons/_rdns_.plist.tmpl")
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
	// Darwin-specific config options
	if c.PrivilegedPorts {
		if !c.System {
			return "", fmt.Errorf("You must use root-owned LaunchDaemons (not user-owned LaunchAgents) to use priveleged ports on OS X")
		}
	}
	plistDir := srvSysPath
	if !c.System {
		plistDir = filepath.Join(c.Home, srvUserPath)
	}

	// Check paths first
	err := os.MkdirAll(filepath.Dir(plistDir), 0755)
	if nil != err {
		return "", err
	}

	b, err := Render(c)
	if nil != err {
		return "", err
	}

	// Write the file out
	// TODO rdns
	plistName := c.ReverseDNS + ".plist"
	plistPath := filepath.Join(plistDir, plistName)
	if err := ioutil.WriteFile(plistPath, b, 0644); err != nil {
		return "", fmt.Errorf("Error writing %s: %v", plistPath, err)
	}

	// TODO --no-start
	err = start(c)
	if nil != err {
		fmt.Printf("If things don't go well you should be able to get additional logging from launchctl:\n")
		fmt.Printf("\tsudo launchctl log level debug\n")
		fmt.Printf("\ttail -f /var/log/system.log\n")
		return "", err
	}

	return "launchd", nil
}
