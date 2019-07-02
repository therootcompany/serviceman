//go:generate go run -mod=vendor github.com/UnnoTed/fileb0x b0x.toml

package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config should describe the service well-enough for it to
// run on Mac, Linux, and Windows.
//
// 	&Config{
// 		// A human-friendy name
// 		Title: "Foobar App",
// 		// A computer-friendly name
// 		Name: "foobar-app",
// 		// A name for OS X plist
// 		ReverseDNS: "com.example.foobar-app",
// 		// A human-friendly description
// 		Desc: "Foobar App",
// 		// The app /service homepage
// 		URL: "https://example.com/foobar-app/",
// 		// The full path of the interpreter, if any (ruby, python, node, etc)
// 		Interpreter: "/opt/node/bin/node",
// 		// The name of the executable (or script)
// 		Exec: "foobar-app.js",
// 		// An array of arguments
// 		Argv: []string{"-c", "/path/to/config.json"},
// 		// A map of Environment variables that should be set
// 		Envs: map[string]string{
// 			PORT: "8080",
// 			ENV: "development",
// 		},
// 		// The user (Linux & Mac only).
// 		// This does not apply to userspace services.
// 		// There may be special considerations
// 		User: "www-data",
// 		// If different from User
// 		Group: "",
// 		// Whether to install as a system or user service
// 		System: false,
// 		// Whether or not the service may need privileged ports
// 		PrivilegedPorts: false,
// 	}
//
// Note that some fields are exported for templating,
// but not intended to be set by you.
// These are documented as omitted from JSON.
// Try to stick to what's outlined above.
type Config struct {
	Title               string            `json:"title"`
	Name                string            `json:"name"`
	Desc                string            `json:"desc"`
	URL                 string            `json:"url"`
	ReverseDNS          string            `json:"reverse_dns"` // i.e. com.example.foo-app
	Interpreter         string            `json:"interpreter"` // i.e. node, python
	Exec                string            `json:"exec"`
	Argv                []string          `json:"argv"`
	Workdir             string            `json:"workdir"`
	Envs                map[string]string `json:"envs"`
	User                string            `json:"user"`
	Group               string            `json:"group"`
	home                string            `json:"-"`
	Local               string            `json:"-"`
	Logdir              string            `json:"-"`
	System              bool              `json:"system"`
	Restart             bool              `json:"restart"`
	Production          bool              `json:"production"`
	PrivilegedPorts     bool              `json:"privileged_ports"`
	MultiuserProtection bool              `json:"multiuser_protection"`
}

// Install will do a best-effort attempt to install a start-on-startup
// user or system service via systemd, launchd, or reg.exe
func Install(c *Config) error {
	if "" == c.Exec {
		c.Exec = c.Name
	}

	if !c.System {
		home, err := os.UserHomeDir()
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unrecoverable Error: %s", err)
			os.Exit(4)
			return err
		} else {
			c.home = home
		}
	}

	err := install(c)
	if nil != err {
		return err
	}

	err = os.MkdirAll(c.Logdir, 0755)
	if nil != err {
		return err
	}

	return nil
}

// Returns true if we suspect that the current user (or process) will be able
// to write to system folders, bind to privileged ports, and otherwise
// successfully run a system service.
func IsPrivileged() bool {
	return isPrivileged()
}

func WhereIs(exec string) (string, error) {
	exec = filepath.ToSlash(exec)
	if strings.Contains(exec, "/") {
		// it's a path (so we don't allow filenames with slashes)
		stat, err := os.Stat(exec)
		if nil != err {
			return "", err
		}
		if stat.IsDir() {
			return "", fmt.Errorf("'%s' is not an executable file", exec)
		}
		return filepath.Abs(exec)
	}
	return whereIs(exec)
}
