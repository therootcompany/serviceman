//go:generate go run -mod=vendor git.rootprojects.org/root/go-gitver

// main runs the things and does the stuff
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"git.rootprojects.org/root/go-serviceman/manager"
	"git.rootprojects.org/root/go-serviceman/runner"
	"git.rootprojects.org/root/go-serviceman/service"
)

var GitRev = "000000000"
var GitVersion = "v0.5.3-pre+dirty"
var GitTimestamp = time.Now().Format(time.RFC3339)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("\tserviceman <command> --help")
	fmt.Println("\tserviceman add ./foo-app -- --foo-arg")
	fmt.Println("\tserviceman run --config ./foo-app.json")
	fmt.Println("\tserviceman list --all")
	fmt.Println("\tserviceman start <name>")
	fmt.Println("\tserviceman stop <name>")
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
	case "version":
		fmt.Println(GitVersion, GitTimestamp, GitRev)
	case "run":
		run()
	case "add":
		add()
	case "start":
		start()
	case "stop":
		stop()
	case "list":
		list()
	default:
		fmt.Fprintf(os.Stderr, "Unknown argument %s\n", top)
		usage()
		os.Exit(1)
	}
}

func add() {
	conf := &service.Service{
		Restart: true,
	}

	force := false
	forUser := false
	forSystem := false
	dryrun := false
	pathEnv := ""
	flag.StringVar(&conf.Title, "title", "", "a human-friendly name for the service")
	flag.StringVar(&conf.Desc, "desc", "", "a human-friendly description of the service (ex: Foo App)")
	flag.StringVar(&conf.Name, "name", "", "a computer-friendly name for the service (ex: foo-app)")
	flag.StringVar(&conf.URL, "url", "", "the documentation on home page of the service")
	flag.StringVar(&conf.Workdir, "workdir", "", "the directory in which the service should be started (if supported)")
	flag.StringVar(&conf.ReverseDNS, "rdns", "", "a plist-friendly Reverse DNS name for launchctl (ex: com.example.foo-app)")
	flag.BoolVar(&forSystem, "system", false, "attempt to add system service as an unprivileged/unelevated user")
	flag.BoolVar(&forUser, "user", false, "add user space / user mode service even when admin/root/sudo/elevated")
	flag.BoolVar(&force, "force", false, "if the interpreter or executable doesn't exist, or things don't make sense, try anyway")
	flag.StringVar(&pathEnv, "path", "", "set the path for the resulting systemd service")
	flag.StringVar(&conf.User, "username", "", "run the service as this user")
	flag.StringVar(&conf.Group, "groupname", "", "run the service as this group")
	flag.BoolVar(&conf.PrivilegedPorts, "cap-net-bind", false, "this service should have access to privileged ports")
	flag.BoolVar(&dryrun, "dryrun", false, "output the service file without modifying anything on disk")
	flag.Parse()
	flagargs := flag.Args()

	// You must have something to run, duh
	n := len(flagargs)
	if 0 == n {
		fmt.Println("Usage: serviceman add ./foo-app --foo-arg")
		os.Exit(2)
		return
	}

	if forUser && forSystem {
		fmt.Println("Pfff! You can't --user AND --system! What are you trying to pull?")
		os.Exit(1)
		return
	}

	// There are three groups of flags
	// serviceman --flag1 arg1 non-flag-arg --child1 -- --raw1 -- --raw2
	//  serviceman --flag1 arg1   // these belong to serviceman
	//  non-flag-arg --child1     // these will be interpretted
	//  --                        // separator
	//  --raw1 -- --raw2          // after the separater (including additional separators) will be ignored
	rawargs := []string{}
	for i := range flagargs {
		if "--" == flagargs[i] {
			if len(flagargs) > i+1 {
				rawargs = flagargs[i+1:]
			}
			flagargs = flagargs[:i]
			break
		}
	}

	// Assumptions
	ass := []string{}
	if forUser {
		conf.System = false
	} else if forSystem {
		conf.System = true
	} else {
		conf.System = manager.IsPrivileged()
		if conf.System {
			ass = append(ass, "# Because you're a privileged user")
			ass = append(ass, "  --system")
			ass = append(ass, "")
		} else {
			ass = append(ass, "# Because you're a unprivileged user")
			ass = append(ass, "  --user")
			ass = append(ass, "")
		}
	}
	if "" == conf.Workdir {
		dir, _ := os.Getwd()
		conf.Workdir = dir
		ass = append(ass, "# Because this is your current working directory")
		ass = append(ass, fmt.Sprintf("  --workdir %s", conf.Workdir))
		ass = append(ass, "")
	}
	if "" == conf.Name {
		name, _ := os.Getwd()
		base := filepath.Base(name)
		ext := filepath.Ext(base)
		n := (len(base) - len(ext))
		name = base[:n]
		if "" == name {
			name = base
		}
		conf.Name = name
		ass = append(ass, "# Because this is the name of your current working directory")
		ass = append(ass, fmt.Sprintf("  --name %s", conf.Name))
		ass = append(ass, "")
	}
	if "" != pathEnv {
		conf.Envs = make(map[string]string)
		conf.Envs["PATH"] = pathEnv
	}

	exepath, err := findExec(flagargs[0], force)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(3)
		return
	}
	flagargs[0] = exepath

	exeargs, err := testScript(flagargs[0], force)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(3)
		return
	}

	flagargs = append(exeargs, flagargs...)
	// TODO
	for i := range flagargs {
		arg := flagargs[i]
		arg = filepath.ToSlash(arg)
		// Paths considered to be anything starting with ./, .\, /, \, C:
		if "." == arg || strings.Contains(arg, "/") {
			//if "." == arg || (len(arg) >= 2 && "./" == arg[:2] || '/' == arg[0] || "C:" == strings.ToUpper(arg[:1])) {
			var err error
			arg, err = filepath.Abs(arg)
			if nil == err {
				_, err = os.Stat(arg)
			}
			if nil != err {
				fmt.Printf("%q appears to be a file path, but %q could not be read\n", flagargs[i], arg)
				if !force {
					os.Exit(7)
					return
				}
				continue
			}

			if '\\' != os.PathSeparator {
				// Convert paths back to .\ for Windows
				arg = filepath.FromSlash(arg)
			}

			// Lookin' good
			flagargs[i] = arg
		}
	}

	// We won't bother with Interpreter here
	// (it's really just for documentation),
	// but we will add any and all unchecked args to the full slice
	conf.Exec = flagargs[0]
	conf.Argv = append(flagargs[1:], rawargs...)

	// TODO update docs: go to the work directory
	// TODO test with "npm start"

	conf.NormalizeWithoutPath()

	//fmt.Printf("\n%#v\n\n", conf)
	if conf.System && !manager.IsPrivileged() {
		fmt.Fprintf(os.Stderr, "Warning: You may need to use 'sudo' to add %q as a privileged system service.\n", conf.Name)
	}

	if len(ass) > 0 {
		fmt.Println("OPTIONS: Making some assumptions...\n")
		for i := range ass {
			fmt.Println("\t" + ass[i])
		}
	}

	// Find who this is running as
	// And pretty print the command to run
	runAs := conf.User
	var wasflag bool
	fmt.Printf("COMMAND: Service %q will be run like this (more or less):\n\n", conf.Title)
	if conf.System {
		if "" == runAs {
			runAs = "root"
		}
		fmt.Printf("\t# Starts on system boot, as %q\n", runAs)
	} else {
		u, _ := user.Current()
		runAs = u.Name
		if "" == runAs {
			runAs = u.Username
		}
		fmt.Printf("\t# Starts as %q, when %q logs in\n", runAs, u.Username)
	}
	//fmt.Printf("\tpushd %s\n", conf.Workdir)
	fmt.Printf("\t%s\n", conf.Exec)
	for i := range conf.Argv {
		arg := conf.Argv[i]
		if '-' == arg[0] {
			if wasflag {
				fmt.Println()
			}
			wasflag = true
			fmt.Printf("\t\t%s", arg)
		} else {
			if wasflag {
				fmt.Printf(" %s\n", arg)
			} else {
				fmt.Printf("\t\t%s\n", arg)
			}
			wasflag = false
		}
	}
	if wasflag {
		fmt.Println()
	}
	fmt.Println()

	// TODO output config without installing
	if dryrun {
		b, err := manager.Render(conf)
		if nil != err {
			fmt.Fprintf(os.Stderr, "Error rendering: %s\n", err)
			os.Exit(10)
		}
		fmt.Println(string(b))
		return
	}

	fmt.Printf("LAUNCHER: ")
	servicetype, err := manager.Install(conf)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(500)
		return
	}

	fmt.Printf("LOGS: ")
	printLogMessage(conf)
	fmt.Println()

	servicemode := "USER MODE"
	if conf.System {
		servicemode = "SYSTEM"
	}
	fmt.Printf(
		"SUCCESS:\n\n\t%q started as a %s %s service, running as %q\n",
		conf.Name,
		servicetype,
		servicemode,
		runAs,
	)
	fmt.Println()
}

func list() {
	var verbose bool
	forUser := false
	forSystem := false
	flag.BoolVar(&forSystem, "system", false, "attempt to add system service as an unprivileged/unelevated user")
	flag.BoolVar(&forUser, "user", false, "add user space / user mode service even when admin/root/sudo/elevated")
	flag.BoolVar(&verbose, "all", false, "show all services (even those not managed by serviceman)")
	flag.Parse()

	if forUser && forSystem {
		fmt.Println("Pfff! You can't --user AND --system! What are you trying to pull?")
		os.Exit(1)
		return
	}

	conf := &service.Service{}
	if forUser {
		conf.System = false
	} else if forSystem {
		conf.System = true
	} else {
		conf.System = manager.IsPrivileged()
	}

	// Pretty much just for HomeDir
	conf.NormalizeWithoutPath()

	managed, others, errs := manager.List(conf)
	for i := range errs {
		fmt.Fprintf(os.Stderr, "possible error: %s\n", errs[i])
	}
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Println("serviceman-managed services:\n")
	for i := range managed {
		fmt.Println("\t" + managed[i])
	}
	if 0 == len(managed) {
		fmt.Println("\t(none)")
	}
	fmt.Println("")

	if verbose {
		fmt.Println("other services:\n")
		for i := range others {
			fmt.Println("\t" + others[i])
		}
		if 0 == len(others) {
			fmt.Println("\t(none)")
		}
		fmt.Println("")
	}
}

func findExec(exe string, force bool) (string, error) {
	// ex: node => /usr/local/bin/node
	// ex: ./demo.js => /Users/aj/project/demo.js
	exepath, err := exec.LookPath(exe)
	if nil != err {
		var msg string
		if strings.Contains(filepath.ToSlash(exe), "/") {
			if _, err := os.Stat(exe); err != nil {
				msg = fmt.Sprintf("Error: '%s' could not be found in PATH or working directory.\n", exe)
			} else {
				msg = fmt.Sprintf("Error: '%s' is not an executable.\nYou may be able to fix that. Try running this:\n\tchmod a+x %s\n", exe, exe)
			}
		} else {
			if _, err := os.Stat(exe); err != nil {
				msg = fmt.Sprintf("Error: '%s' could not be found in PATH", exe)
			} else {
				msg = fmt.Sprintf("Error: '%s' could not be found in PATH, did you mean './%s'?\n", exe, exe)
			}
		}
		if !force {
			return "", fmt.Errorf(msg)
		}
		fmt.Fprintf(os.Stderr, "%s\n", msg)
		return exe, nil
	}

	// ex: \Users\aj\project\demo.js => /Users/aj/project/demo.js
	// Can't have an error here when lookpath succeeded
	exepath, _ = filepath.Abs(filepath.ToSlash(exepath))
	return exepath, nil
}

func testScript(exepath string, force bool) ([]string, error) {
	f, err := os.Open(exepath)
	b := make([]byte, 256)
	if nil == err {
		_, err = f.Read(b)
	}
	if nil != err || len(b) < len("#!/x") {
		msg := fmt.Sprintf("Error when testing if '%s' is a binary or script: could not read file: %s\n", exepath, err)
		if !force {
			return nil, fmt.Errorf(msg)
		}
		fmt.Fprintf(os.Stderr, "%s\n", msg)
		return nil, nil
	}

	// Nott sure if this is more readable and idiomatic as if else or switch
	// However, the order matters
	switch {
	case utf8.Valid(b):
		// Looks like an executable script
		if "#!/" == string(b[:3]) {
			break
		}

		msg := fmt.Sprintf("Error: %q looks like a script, but we don't know the interpreter.\nYou can probably fix this by...\n"+
			"\tExplicitly naming the interpreter (ex: 'python my-script.py' instead of just 'my-script.py')\n"+
			"\tPlacing a hashbang at the top of the script (ex: '#!/usr/bin/env python')", exepath)

		if !force {
			return nil, fmt.Errorf(msg)
		}
		return nil, nil
	case "#!/" != string(b[:3]):
		// Looks like a normal binary
		return nil, nil
	default:
		// Looks like a corrupt script file
		msg := "Error: It looks like you've specified a corrupt script file."
		if !force {
			return nil, fmt.Errorf(msg)
		}
		return nil, nil
	}

	// Deal with #!/whatever

	// Get that first line
	// "#!/usr/bin/env node" => ["/usr/bin/env", "node"]
	// "#!/usr/bin/node --harmony => ["/usr/bin/node", "--harmony"]
	s := string(b[2:]) // strip leading #!
	s = strings.Split(strings.Replace(s, "\r\n", "\n", -1), "\n")[0]
	allargs := strings.Split(strings.TrimSpace(s), " ")
	args := []string{}
	for i := range allargs {
		arg := strings.TrimSpace(allargs[i])
		if "" != arg {
			args = append(args, arg)
		}
	}
	if strings.HasSuffix(args[0], "/env") && len(args) > 1 {
		// TODO warn that "env" is probably not an executable if 1 = len(args)?
		args = args[1:]
	}
	exepath, err = findExec(args[0], force)
	if nil != err {
		return nil, err
	}
	args[0] = exepath

	return args, nil
}

func start() {
	forUser := false
	forSystem := false
	flag.BoolVar(&forSystem, "system", false, "attempt to add system service as an unprivileged/unelevated user")
	flag.BoolVar(&forUser, "user", false, "add user space / user mode service even when admin/root/sudo/elevated")
	flag.Parse()

	args := flag.Args()
	if 1 != len(args) {
		fmt.Println("Usage: serviceman start <name>")
		os.Exit(1)
	}

	if forUser && forSystem {
		fmt.Println("Pfff! You can't --user AND --system! What are you trying to pull?")
		os.Exit(1)
		return
	}

	conf := &service.Service{
		Name:    args[0],
		Restart: false,
	}
	if forUser {
		conf.System = false
	} else if forSystem {
		conf.System = true
	} else {
		conf.System = manager.IsPrivileged()
	}
	conf.NormalizeWithoutPath()

	err := manager.Start(conf)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(500)
		return
	}
}

func stop() {
	forUser := false
	forSystem := false
	flag.BoolVar(&forSystem, "system", false, "attempt to add system service as an unprivileged/unelevated user")
	flag.BoolVar(&forUser, "user", false, "add user space / user mode service even when admin/root/sudo/elevated")
	flag.Parse()

	args := flag.Args()
	if 1 != len(args) {
		fmt.Println("Usage: serviceman stop <name>")
		os.Exit(1)
	}

	if forUser && forSystem {
		fmt.Println("Pfff! You can't --user AND --system! What are you trying to pull?")
		os.Exit(1)
		return
	}

	conf := &service.Service{
		Name:    args[0],
		Restart: false,
	}
	if forUser {
		conf.System = false
	} else if forSystem {
		conf.System = true
	} else {
		conf.System = manager.IsPrivileged()
	}
	conf.NormalizeWithoutPath()

	if err := manager.Stop(conf); nil != err {
		fmt.Println(err)
		os.Exit(127)
	}
}

func run() {
	var confpath string
	var daemonize bool
	flag.StringVar(&confpath, "config", "", "path to a config file to run")
	flag.BoolVar(&daemonize, "daemon", false, "spawn a child process that lives in the background, and exit")
	flag.Parse()

	if "" == confpath {
		fmt.Fprintf(os.Stderr, "%s\n", strings.Join(flag.Args(), " "))
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

	force := false
	s.Normalize(force)
	fmt.Printf("All output will be directed to the logs at:\n\t%s\n", s.Logdir)
	err = os.MkdirAll(s.Logdir, 0755)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	if !daemonize {
		//fmt.Fprintf(os.Stdout, "Running %s %s %s\n", s.Interpreter, s.Exec, strings.Join(s.Argv, " "))
		if err := runner.Start(s); nil != err {
			fmt.Println("Error:", err)
		}
		return
	}

	manager.Run(os.Args[0], "run", "--config", confpath)
}
