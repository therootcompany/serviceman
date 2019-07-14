package manager

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

func getService(system bool, home string, name string) (string, error) {
	sys, user, err := getMatchingSrvs(home, name)
	if nil != err {
		return "", err
	}

	var service string
	if system {
		service, err = getOneSysSrv(sys, user, name)
		if nil != err {
			return "", err
		}
	} else {
		service, err = getOneUserSrv(home, sys, user, name)
		if nil != err {
			return "", err
		}
	}

	return service, nil
}

// Runnable defines a command to run, along with its arguments,
// and whether or not failing to exit successfully matters.
// It also defines whether certains words must exist (or not exist)
// in its output, apart from existing successfully, to determine
// whether or not it was actually successful.
type Runnable struct {
	Exec     string
	Args     []string
	Must     bool
	Keywords []string
	Badwords []string
}

func (x Runnable) Run() error {
	cmd := exec.Command(x.Exec, x.Args...)
	out, err := cmd.CombinedOutput()
	if !x.Must {
		return nil
	}

	good := true
	str := string(out)
	for j := range x.Keywords {
		if !strings.Contains(str, x.Keywords[j]) {
			good = false
			break
		}
	}
	if good && 0 != len(x.Badwords) {
		for j := range x.Badwords {
			if "" != x.Badwords[j] && !strings.Contains(str, x.Badwords[j]) {
				good = false
				break
			}
		}
	}
	if nil != err {
		var comment string
		if len(x.Keywords) > 0 {
			comment += "# output must match all of:\n"
			comment += "# \t" + strings.Join(x.Keywords, "\n#\t") + "\n"
		}
		if len(x.Badwords) > 0 {
			comment += "# output must not match any of:\n"
			comment += "# \t" + strings.Join(x.Badwords, "\n#\t") + "\n"
		}
		return fmt.Errorf("Failed to run %s %s\n%s\n%s\n", x.Exec, strings.Join(x.Args, " "), str, comment)
	}

	return nil
}

func (x Runnable) String() string {
	var must = "true"

	if x.Must {
		must = "exit"
	}

	return strings.TrimSpace(fmt.Sprintf(
		"%s %s || %s\n",
		x.Exec,
		strings.Join(x.Args, " "),
		must,
	))
}

func getSrvs(dir string) ([]string, error) {
	plists := []string{}

	infos, err := ioutil.ReadDir(dir)
	if nil != err {
		return nil, err
	}

	for i := range infos {
		x := infos[i]
		fname := strings.ToLower(x.Name())
		if strings.HasSuffix(fname, srvExt) {
			plists = append(plists, x.Name())
		}
	}

	return plists, nil
}

func getSystemSrvs() ([]string, error) {
	return getSrvs(srvSysPath)
}

func getUserSrvs(home string) ([]string, error) {
	return getSrvs(filepath.Join(home, srvUserPath))
}

// "come.example.foo.plist" matches "foo"
func filterMatchingSrvs(plists []string, name string) []string {
	filtered := []string{}

	for i := range plists {
		pname := plists[i]
		lname := strings.ToLower(pname)
		n := len(lname)
		if strings.HasSuffix(lname[:n-srvLen], strings.ToLower(name)) {
			filtered = append(filtered, pname)
		}
	}

	return filtered
}

func getMatchingSrvs(home string, name string) ([]string, []string, error) {
	sysPlists, err := getSystemSrvs()
	if nil != err {
		return nil, nil, err
	}

	var userPlists []string
	if "" != home {
		userPlists, err = getUserSrvs(home)
		if nil != err {
			return nil, nil, err
		}
	}

	return filterMatchingSrvs(sysPlists, name), filterMatchingSrvs(userPlists, name), nil
}

func getExactSrvMatch(srvs []string, name string) string {
	for i := range srvs {
		srv := srvs[i]
		n := len(srv)
		if srv[:n-srvLen] == strings.ToLower(name) {
			return srv
		}
	}

	return ""
}

func getOneSysSrv(sys []string, user []string, name string) (string, error) {
	if service := getExactSrvMatch(user, name); "" != service {
		return filepath.Join(srvSysPath, service), nil
	}

	n := len(sys)
	switch {
	case 0 == n:
		errstr := fmt.Sprintf("Didn't find user service matching %q\n", name)
		if 0 != len(user) {
			errstr += fmt.Sprintf("Did you intend to run a user service instead?\n\t%s\n", strings.Join(user, "\n\t"))
		}
		return "", fmt.Errorf(errstr)
	case n > 1:
		errstr := fmt.Sprintf("Found more than one matching service:\n\t%s\n", strings.Join(sys, "\n\t"))
		return "", fmt.Errorf(errstr)
	default:
		return filepath.Join(srvSysPath, sys[0]), nil
	}
}

func getOneUserSrv(home string, sys []string, user []string, name string) (string, error) {
	if service := getExactSrvMatch(user, name); "" != service {
		return filepath.Join(home, srvUserPath, service), nil
	}

	n := len(user)
	switch {
	case 0 == n:
		errstr := fmt.Sprintf("Didn't find user service matching %q\n", name)
		if 0 != len(sys) {
			errstr += fmt.Sprintf("Did you intend to run a system service instead?\n\t%s\n", strings.Join(sys, "\n\t"))
		}
		return "", fmt.Errorf(errstr)
	case n > 1:
		errstr := fmt.Sprintf("Found more than one matching service:\n\t%s\n", strings.Join(user, "\n\t"))
		return "", fmt.Errorf(errstr)
	default:
		return filepath.Join(home, srvUserPath, user[0]), nil
	}
}

func adjustPrivs(system bool, cmds []Runnable) []Runnable {
	if !system || isPrivileged() {
		return cmds
	}

	sudos := cmds
	cmds = []Runnable{}
	for i := range sudos {
		exe := sudos[i]
		exe.Args = append([]string{exe.Exec}, exe.Args...)
		exe.Exec = "sudo"
		cmds = append(cmds, exe)
	}

	return cmds
}

func Run(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	// for debugging
	/*
		out, err := cmd.CombinedOutput()
		if nil != err {
			fmt.Println(err)
		}
		fmt.Println(string(out))
	*/

	err := cmd.Start()
	if nil != err {
		return err
	}
	return nil
}
