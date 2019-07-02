//go:generate go run -mod=vendor git.rootprojects.org/root/go-gitver

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.rootprojects.org/root/go-serviceman/installer"
)

var GitRev = "000000000"
var GitVersion = "v0.0.0"
var GitTimestamp = time.Now().Format(time.RFC3339)

func main() {
	conf := &installer.Config{
		Restart: true,
	}

	args := []string{}
	for i := range os.Args {
		if "--" == os.Args[i] {
			if len(os.Args) > i+1 {
				args = os.Args[i+1:]
			}
			os.Args = os.Args[:i]
			break
		}
	}
	conf.Argv = args

	force := false
	forUser := false
	forSystem := false
	flag.StringVar(&conf.Title, "title", "", "a human-friendly name for the service")
	flag.StringVar(&conf.Desc, "desc", "", "a human-friendly description of the service (ex: Foo App)")
	flag.StringVar(&conf.Name, "name", "", "a computer-friendly name for the service (ex: foo-app)")
	flag.StringVar(&conf.URL, "url", "", "the documentation on home page of the service")
	//flag.StringVar(&conf.Workdir, "workdir", "", "the directory in which the service should be started")
	flag.StringVar(&conf.ReverseDNS, "rdns", "", "a plist-friendly Reverse DNS name for launchctl (ex: com.example.foo-app)")
	flag.BoolVar(&forSystem, "system", false, "attempt to install system service as an unprivileged/unelevated user")
	flag.BoolVar(&forUser, "user", false, "install user space / user mode service even when admin/root/sudo/elevated")
	flag.BoolVar(&force, "force", false, "if the interpreter or executable doesn't exist, or things don't make sense, try anyway")
	flag.StringVar(&conf.User, "username", "", "run the service as this user")
	flag.StringVar(&conf.Group, "groupname", "", "run the service as this group")
	flag.BoolVar(&conf.PrivilegedPorts, "cap-net-bind", false, "this service should have access to privileged ports")
	flag.Parse()
	args = flag.Args()

	if forUser && forSystem {
		fmt.Println("Pfff! You can't --user AND --system! What are you trying to pull?")
		os.Exit(1)
		return
	}
	if forUser {
		conf.System = false
	} else if forSystem {
		conf.System = true
	} else {
		conf.System = installer.IsPrivileged()
	}

	n := len(args)
	if 0 == n {
		fmt.Println("Usage: serviceman install ./foo-app -- --foo-arg")
		os.Exit(2)
		return
	}

	execpath, err := installer.WhereIs(args[0])
	if nil != err {
		fmt.Fprintf(os.Stderr, "Error: '%s' could not be found.\n", args[0])
		if !force {
			os.Exit(3)
			return
		}
	} else {
		args[0] = execpath
	}
	conf.Exec = args[0]
	args = args[1:]

	if n >= 2 {
		conf.Interpreter = conf.Exec
		conf.Exec = args[0]
		conf.Argv = append(args[1:], conf.Argv...)
	}

	if "" == conf.Name {
		ext := filepath.Ext(conf.Exec)
		base := filepath.Base(conf.Exec[:len(conf.Exec)-len(ext)])
		conf.Name = strings.ToLower(base)
	}
	if "" == conf.Title {
		conf.Title = conf.Name
	}
	if "" == conf.ReverseDNS {
		// technically should be something more like "com.example." + conf.Name,
		// but whatever
		conf.ReverseDNS = conf.Name
	}

	if !conf.System {
		home, err := os.UserHomeDir()
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unrecoverable Error: %s", err)
			os.Exit(4)
			return
		}
		conf.Local = filepath.Join(home, ".local")
		conf.Logdir = filepath.Join(home, ".local", "share", conf.Name, "var", "log")
	} else {
		conf.Logdir = "/var/log/" + conf.Name
	}

	// Check to see if Exec exists
	//   /whatever => must exist exactly
	//   ./whatever => must exist in current or WorkDir(TODO)
	//   whatever => may also exist in {{ .Local }}/opt/{{ .Name }}/{{ .Exec }}
	_, err = os.Stat(conf.Exec)
	if nil != err {
		bad := true
		if !strings.Contains(filepath.ToSlash(conf.Exec), "/") {
			optpath := filepath.Join(conf.Local, "/opt", conf.Name, conf.Exec)
			_, err := os.Stat(optpath)
			if nil == err {
				bad = false
				fmt.Fprintf(os.Stderr, "Using '%s' for '%s'\n", optpath, conf.Exec)
				conf.Exec = optpath
			}
		}

		if bad {
			// TODO look for it in WorkDir?
			fmt.Fprintf(os.Stderr, "Error: '%s' could not be found.\n", conf.Exec)
			if !force {
				os.Exit(5)
				return
			}
			execpath, err := filepath.Abs(conf.Exec)
			if nil == err {
				conf.Exec = execpath
			}
			fmt.Fprintf(os.Stderr, "Using '%s' anyway.\n", conf.Exec)
		}
	} else {
		execpath, err := filepath.Abs(conf.Exec)
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unrecoverable Error: %s", err)
			os.Exit(4)
		} else {
			conf.Exec = execpath
		}
	}

	fmt.Printf("\n%#v\n\n", conf)

	err = installer.Install(conf)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		fmt.Fprintf(os.Stderr, "Use 'sudo' to install as a privileged system service.\n")
		fmt.Fprintf(os.Stderr, "Use '--user' to install as an user service.\n")
	}
}
