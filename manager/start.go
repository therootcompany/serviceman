package manager

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

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
		return fmt.Errorf("Failed to run %s %s\n%s\n", x.Exec, strings.Join(x.Args, " "), str)
	}

	return nil
}

func (x Runnable) String() string {
	var comment string
	var must = "true"

	if x.Must {
		must = "exit"
		if len(x.Keywords) > 0 {
			comment += "# output must match all of:\n"
			comment += "\t" + strings.Join(x.Keywords, "#\t \n") + "\n"
		}
		if len(x.Badwords) > 0 {
			comment += "# output must not match any of:\n"
			comment += "\t" + strings.Join(x.Keywords, "#\t \n") + "\n"
		}
	}

	return strings.TrimSpace(fmt.Sprintf(
		"%s %s || %s\n%s",
		x.Exec,
		strings.Join(x.Args, " "),
		must,
		comment,
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
	dir := filepath.Join(home, srvUserPath)
	return getSrvs(dir)
}

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
	service := getExactSrvMatch(user, name)
	if "" != service {
		return service, nil
	}

	var errstr string
	// system service was wanted
	n := len(sys)
	switch {
	case 0 == n:
		errstr += fmt.Sprintf("Didn't find user service matching %q\n", name)
		if 0 != len(user) {
			errstr += fmt.Sprintf("Did you intend to run a user service instead?\n\t%s\n", strings.Join(user, "\n\t"))
		}
	case n > 1:
		errstr += fmt.Sprintf("Found more than one matching service:\n\t%s\n", strings.Join(sys, "\n\t"))
	default:
		service = filepath.Join(srvSysPath, sys[0])
	}

	if "" != errstr {
		return "", fmt.Errorf(errstr)
	}

	return service, nil
}

func getOneUserSrv(home string, sys []string, user []string, name string) (string, error) {
	service := getExactSrvMatch(user, name)
	if "" != service {
		return service, nil
	}

	var errstr string
	// user service was wanted
	n := len(user)
	switch {
	case 0 == n:
		errstr += fmt.Sprintf("Didn't find user service matching %q\n", name)
		if 0 != len(sys) {
			errstr += fmt.Sprintf("Did you intend to run a system service instead?\n\t%s\n", strings.Join(sys, "\n\t"))
		}
	case n > 1:
		errstr += fmt.Sprintf("Found more than one matching service:\n\t%s\n", strings.Join(user, "\n\t"))
	default:
		service = filepath.Join(home, srvUserPath, user[0]+srvExt)
	}

	if "" != errstr {
		return "", fmt.Errorf(errstr)
	}

	return service, nil
}
