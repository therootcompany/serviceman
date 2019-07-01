//go:generate go run -mod=vendor git.rootprojects.org/root/go-gitver

package main

import (
	"flag"
	"fmt"
	"log"
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
	conf.Args = strings.Join(conf.Argv, " ")

	forUser := false
	forSystem := false
	flag.StringVar(&conf.Title, "title", "", "a human-friendly name for the service")
	flag.StringVar(&conf.Desc, "desc", "", "a human-friendly description of the service (ex: Foo App)")
	flag.StringVar(&conf.Name, "name", "", "a computer-friendly name for the service (ex: foo-app)")
	flag.StringVar(&conf.URL, "url", "", "the documentation on home page of the service")
	flag.StringVar(&conf.ReverseDNS, "rdns", "", "a plist-friendly Reverse DNS name for launchctl (ex: com.example.foo-app)")
	flag.BoolVar(&forSystem, "system", false, "attempt to install system service as an unprivileged/unelevated user")
	flag.BoolVar(&forUser, "user", false, "install user space / user mode service even when admin/root/sudo/elevated")
	flag.StringVar(&conf.User, "username", "", "run the service as this user")
	flag.StringVar(&conf.Group, "groupname", "", "run the service as this group")
	flag.BoolVar(&conf.PrivilegedPorts, "cap-net-bind", false, "this service should have access to privileged ports")
	flag.Parse()
	args = flag.Args()

	if forUser && forSystem {
		fmt.Println("Pfff! You can't --user AND --system! What are you trying to pull?")
		os.Exit(1)
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
		os.Exit(1)
	}

	execpath, err := installer.WhereIs(args[0])
	if nil != err {
		fmt.Fprintf(os.Stderr, "Error: '%s' could not be found.", args[0])
		os.Exit(1)
	}
	args[0] = execpath
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
		conf.ReverseDNS = "com.example." + conf.Name
	}

	fmt.Printf("\n%#v\n\n", conf)

	err = installer.Install(conf)
	if nil != err {
		log.Fatal(err)
	}
}
