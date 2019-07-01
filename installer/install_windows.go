package installer

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// TODO system service requires elevated privileges
// See https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
func install(c *Config) error {
	//token := windows.Token(0)
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

	setArgs := ""
	args := c.Argv
	exec := filepath.Join(c.home, ".local", "opt", c.Name, c.Exec)
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
	if len(regSZ) > 260 {
		return fmt.Errorf("data value is too long for registry entry")
	}
	fmt.Println("Set Registry Key:")
	fmt.Println(autorunKey, c.Title, regSZ)
	k.SetStringValue(c.Title, regSZ)

	return nil
}

func whereIs(exe string) (string, error) {
	cmd := exec.Command("where.exe", exe)
	out, err := cmd.Output()
	if nil != err {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
