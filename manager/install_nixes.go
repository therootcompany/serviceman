// +build !windows

package manager

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"git.rootprojects.org/root/go-serviceman/service"
)

// this code is shared between Mac and Linux, but may diverge in the future
func list(c *service.Service) ([]string, []string, []error) {
	confDir := srvSysPath
	if !c.System {
		confDir = filepath.Join(c.Home, srvUserPath)
	}

	// Enuser path exists
	err := os.MkdirAll(confDir, 0755)
	if nil != err {
		return nil, nil, []error{err}
	}

	fis, err := ioutil.ReadDir(confDir)
	if nil != err {
		return nil, nil, []error{err}
	}

	managed := []string{}
	others := []string{}
	errs := []error{}
	b := make([]byte, 256)
	for i := range fis {
		fi := fis[i]
		if !strings.HasSuffix(strings.ToLower(fi.Name()), srvExt) || len(fi.Name()) <= srvLen {
			continue
		}

		confFile := filepath.Join(confDir, fi.Name())
		r, err := os.Open(confFile)
		if nil != err {
			errs = append(errs, &ManageError{
				Name:   confFile,
				Hint:   "Open file",
				Parent: err,
			})
			continue
		}

		n, err := r.Read(b)
		if nil != err {
			errs = append(errs, &ManageError{
				Name:   confFile,
				Hint:   "Read file",
				Parent: err,
			})
			continue
		}
		b = b[:n]

		name := fi.Name()[:len(fi.Name())-srvLen]
		if bytes.Contains(b, []byte("for serviceman.")) {
			managed = append(managed, name)
		} else {
			others = append(others, name)
		}
	}

	return managed, others, errs
}
