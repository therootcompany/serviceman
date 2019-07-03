package installer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.rootprojects.org/root/go-serviceman/service"

	"golang.org/x/sys/windows/registry"
)

// TODO nab some goodness from https://github.com/takama/daemon

// TODO system service requires elevated privileges
// See https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
func install(c *service.Service) error {
	/*
		// LEAVE THIS DOCUMENTATION HERE
		reg.exe
		/V <value name> - "Telebit"
		/T <data type> - "REG_SZ" - String
		/D <value data>
		/C - case sensitive
		/F <search data??> - not sure...

		// Special Note:
		"/c" is similar to -- (*nix), and required within the data string
		So instead of setting "do.exe --do-arg1 --do-arg2"
		you must set "do.exe /c --do-arg1 --do-arg2"

		vars.telebitNode += '.exe';
		var cmd = 'reg.exe add "HKCU\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run"'
		+ ' /V "Telebit" /t REG_SZ /D '
		+ '"' + things.argv[0] + ' /c '  // something like C:\Program Files (x64)\nodejs\node.exe
		+ [ path.join(__dirname, 'bin/telebitd.js')
			, 'daemon'
			, '--config'
			, path.join(os.homedir(), '.config/telebit/telebitd.yml')
			].join(' ')
		+ '" /F'
		;
	*/
	autorunKey := `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		autorunKey,
		registry.SET_VALUE,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	args, err := installServiceman(c)
	if nil != err {
		return err
	}

	/*
		setArgs := ""
		args := c.Argv
		exec := c.Exec
		bin := c.Interpreter
		if "" != bin {
			// If this is something like node or python,
			// the interpeter must be called as "the main thing"
			// and "the app" must be an argument
			args = append([]string{exec}, args...)
		} else {
			// Otherwise, if "the app" is a true binary,
			// it can be "the main thing"
			bin = exec
		}
		if 0 != len(args) {
			// On Windows the /c acts kinda like -- does on *nix,
			// at least for commands in the registry that have arguments
			setArgs = ` /c `
		}

		// The final string ends up looking something like one of these:
		// "C:\Users\aj\.local\opt\appname\appname.js /c -p 8080"
		// "C:\Program Files (x64)\nodejs\node.exe /c C:\Users\aj\.local\opt\appname\appname.js -p 8080"
		regSZ := bin + setArgs + strings.Join(c.Argv, " ")
	*/

	regSZ := fmt.Sprintf("%s /c %s", args[0], strings.Join(args[1:], " "))
	if len(regSZ) > 260 {
		return fmt.Errorf("data value is too long for registry entry")
	}
	fmt.Println("Set Registry Key:")
	fmt.Println(autorunKey, c.Title, regSZ)
	k.SetStringValue(c.Title, regSZ)

	return nil
}

// copies self to install path and returns config path
func installServiceman(c *service.Service) ([]string, error) {
	// TODO check version and upgrade or dismiss
	self := os.Args[0]
	smdir := `\opt\serviceman`
	// TODO support service level services (which probably wouldn't need serviceman)
	smdir = filepath.Join(c.Home, ".local", smdir)
	// for now we'll scope the runner to the name of the application
	smbin := filepath.Join(smdir, `bin\serviceman.%s`, c.Name)

	if smbin != self {
		err := os.MkdirAll(filepath.Dir(smbin))
		if nil != err {
			return "", err
		}
		bin, err := ioutil.ReadFile(self)
		if nil != err {
			return "", err
		}
		err := ioutil.WriteFile(smbin, bin, 0755)
		if nil != err {
			return "", err
		}
	}

	b, err := json.Marshal(c)
	if nil != err {
		// this should be impossible, so we'll just panic
		panic(err)
	}
	confpath := filepath.Join(smpath, `etc`, conf.Name+`.json`)
	err := ioutil.WriteFile(confpath, b, 0640)
	if nil != err {
		return err
	}

	return []string{
		smbin,
		"run",
		"--config",
		confpath,
	}, nil
}

func whereIs(exe string) (string, error) {
	// TODO use exec.LookPath instead
	cmd := exec.Command("where.exe", exe)
	out, err := cmd.Output()
	if nil != err {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
