package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.rootprojects.org/root/go-serviceman/runner"
	"git.rootprojects.org/root/go-serviceman/service"

	"golang.org/x/sys/windows/registry"
)

var (
	srvLen      int
	srvExt      = ".json"
	srvSysPath  = "/opt/serviceman/etc"
	srvUserPath = ".local/opt/serviceman/etc"
)

func init() {
	srvLen = len(srvExt)
}

// TODO nab some goodness from https://github.com/takama/daemon

// TODO system service requires elevated privileges
// See https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
func install(c *service.Service) (string, error) {
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

	// Try to stop before trying to copy the file
	_ = runner.Stop(c)

	args, err := installServiceman(c)
	if nil != err {
		return "", err
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

		// The final string ends up looking something like one of these:
		// `"C:\Users\aj\.local\opt\appname\appname.js" -p 8080`
		// `"C:\Program Files (x64)\nodejs\node.exe" C:\Users\aj\.local\opt\appname\appname.js -p 8080`
		regSZ := bin + setArgs + strings.Join(c.Argv, " ")
	*/

	regSZ := fmt.Sprintf(`"%s" %s`, args[0], strings.Join(args[1:], " "))
	if len(regSZ) > 260 {
		return "", fmt.Errorf("data value is too long for registry entry")
	}
	// In order for a windows gui program to not show a console,
	// it has to not output any messages?
	//fmt.Println("Set Registry Key:")
	//fmt.Println(autorunKey, c.Title, regSZ)
	k.SetStringValue(c.Title, regSZ)

	err = start(c)
	return "serviceman", err
}

func Render(c *service.Service) ([]byte, error) {
	b, err := json.Marshal(c)
	if nil != err {
		return nil, err
	}
	return b, nil
}

func start(conf *service.Service) error {
	args := getRunnerArgs(conf)
	args = append(args, "--daemon")
	return Run(args[0], args[1:]...)
}

func stop(conf *service.Service) error {
	return runner.Stop(conf)
}

func list(c *service.Service) ([]string, []string, []error) {
	var errs []error

	regs, err := listRegistry(c)
	if nil != err {
		errs = append(errs, err)
	}

	cfgs, errors := listConfigs(c)
	if 0 != len(errors) {
		errs = append(errs, errors...)
	}

	managed := []string{}
	for i := range cfgs {
		managed = append(managed, cfgs[i].Name)
	}

	others := []string{}
	for i := range regs {
		reg := regs[i]
		if 0 == len(cfgs) {
			others = append(others, reg)
			continue
		}

		var found bool
		for j := range cfgs {
			cfg := cfgs[j]
			// Registry Value Names are case-insensitive
			if strings.ToLower(reg) == strings.ToLower(cfg.Title) {
				found = true
			}
		}
		if !found {
			others = append(others, reg)
		}
	}

	return managed, others, errs
}

func getRunnerArgs(c *service.Service) []string {
	self := os.Args[0]
	debug := ""
	if strings.Contains(self, "debug.exe") {
		debug = "debug."
	}

	smdir := `\opt\serviceman`
	// TODO support service level services (which probably wouldn't need serviceman)
	smdir = filepath.Join(c.Home, ".local", smdir)
	// for now we'll scope the runner to the name of the application
	smbin := filepath.Join(smdir, `bin\serviceman.`+debug+c.Name+`.exe`)

	confpath := filepath.Join(smdir, `etc`)
	conffile := filepath.Join(confpath, c.Name+`.json`)

	return []string{
		smbin,
		"run",
		"--config",
		conffile,
	}
}

type winConf struct {
	Filename string `json:"-"`
	Name     string `json:"name"`
	Title    string `json:"title"`
}

func listConfigs(c *service.Service) ([]winConf, []error) {
	var errs []error

	smdir := `\opt\serviceman`
	if !c.System {
		smdir = filepath.Join(c.Home, ".local", smdir)
	}
	confpath := filepath.Join(smdir, `etc`)

	infos, err := ioutil.ReadDir(confpath)
	if nil != err {
		if os.IsNotExist(err) {
			return nil, nil
		}
		errs = append(errs, &ManageError{
			Name:   confpath,
			Hint:   "Read directory",
			Parent: err,
		})
		return nil, errs
	}

	// TODO report active status
	srvs := []winConf{}
	for i := range infos {
		filename := strings.ToLower(infos[i].Name())
		if len(filename) <= srvLen || !strings.HasSuffix(filename, srvExt) {
			continue
		}

		name := filename[:len(filename)-srvLen]
		b, err := ioutil.ReadFile(filepath.Join(confpath, filename))
		if nil != err {
			errs = append(errs, &ManageError{
				Name:   name,
				Hint:   "Read file",
				Parent: err,
			})
			continue
		}
		cfg := winConf{Filename: filename}
		err = json.Unmarshal(b, &cfg)
		if nil != err {
			errs = append(errs, &ManageError{
				Name:   name,
				Hint:   "Parse JSON",
				Parent: err,
			})
			continue
		}

		srvs = append(srvs, cfg)
	}

	return srvs, errs
}

func listRegistry(c *service.Service) ([]string, error) {
	autorunKey := `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		autorunKey,
		registry.QUERY_VALUE,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	return k.ReadValueNames(-1)
}

// copies self to install path and returns config path
func installServiceman(c *service.Service) ([]string, error) {
	// TODO check version and upgrade or dismiss
	self := os.Args[0]

	args := getRunnerArgs(c)
	smbin := args[0]
	conffile := args[len(args)-1]

	if smbin != self {
		err := os.MkdirAll(filepath.Dir(smbin), 0755)
		if nil != err {
			return nil, err
		}
		bin, err := ioutil.ReadFile(self)
		if nil != err {
			return nil, err
		}
		err = ioutil.WriteFile(smbin, bin, 0755)
		if nil != err {
			return nil, err
		}
	}

	b, err := Render(c)
	if nil != err {
		// this should be impossible, so we'll just panic
		panic(err)
	}
	err = os.MkdirAll(filepath.Dir(conffile), 0755)
	if nil != err {
		return nil, err
	}
	err = ioutil.WriteFile(conffile, b, 0640)
	if nil != err {
		return nil, err
	}

	return args, nil
}
