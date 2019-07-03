//go:generate go run -mod=vendor git.rootprojects.org/root/go-gitver

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.rootprojects.org/root/go-serviceman/installer"
	"git.rootprojects.org/root/go-serviceman/runner"
	"git.rootprojects.org/root/go-serviceman/service"
)

var GitRev = "000000000"
var GitVersion = "v0.0.0"
var GitTimestamp = time.Now().Format(time.RFC3339)

func usage() {
	fmt.Println("Usage: serviceman install ./foo-app -- --foo-arg")
	fmt.Println("Usage: serviceman run --config ./foo-app.json")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Too few arguments: %s\n", strings.Join(os.Args, " "))
		usage()
		os.Exit(1)
	}

	top := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)
	switch top {
	case "install":
		install()
	case "run":
		run()
	default:
		fmt.Fprintf(os.Stderr, "Unknown argument %s\n", top)
		usage()
		os.Exit(1)
	}
}

func install() {
	conf := &service.Service{
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

	conf.Normalize(force)

	fmt.Printf("\n%#v\n\n", conf)

	err = installer.Install(conf)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		fmt.Fprintf(os.Stderr, "Use 'sudo' to install as a privileged system service.\n")
		fmt.Fprintf(os.Stderr, "Use '--user' to install as an user service.\n")
	}
}

func run() {
	var confpath string
	var daemonize bool
	flag.StringVar(&confpath, "config", "", "path to a config file to run")
	flag.BoolVar(&daemonize, "daemon", false, "spawn a child process that lives in the background, and exit")
	flag.Parse()

	if "" == confpath {
		fmt.Fprintf(os.Stderr, "%s", strings.Join(flag.Args(), " "))
		fmt.Fprintf(os.Stderr, "--config /path/to/config.json is required\n")
		usage()
		os.Exit(1)
	}

	b, err := ioutil.ReadFile(confpath)
	if nil != err {
		fmt.Fprintf(os.Stderr, "Couldn't read config file: %s\n", err)
		os.Exit(400)
	}

	s := &service.Service{}
	err = json.Unmarshal(b, s)
	if nil != err {
		fmt.Fprintf(os.Stderr, "Couldn't JSON parse config file: %s\n", err)
		os.Exit(400)
	}

	m := map[string]interface{}{}
	err = json.Unmarshal(b, &m)
	if nil != err {
		fmt.Fprintf(os.Stderr, "Couldn't JSON parse config file: %s\n", err)
		os.Exit(400)
	}

	// default Restart to true
	if _, ok := m["restart"]; !ok {
		s.Restart = true
	}

	if "" == s.Exec {
		fmt.Fprintf(os.Stderr, "Missing exec\n")
		os.Exit(400)
	}

	s.Normalize(false)
	fmt.Fprintf(os.Stdout, "Logdir: %s\n", s.Logdir)
	if !daemonize {
		fmt.Fprintf(os.Stdout, "Running %s %s %s\n", s.Interpreter, s.Exec, strings.Join(s.Argv, " "))
		runner.Run(s)
		return
	}

	cmd := exec.Command(os.Args[0], "run", "--config", confpath)
	// for debugging
	/*
		out, err := cmd.CombinedOutput()
		if nil != err {
			fmt.Println(err)
		}
		fmt.Println(string(out))
	*/

	err = cmd.Start()
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(500)
	}
}
