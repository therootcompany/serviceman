package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestEmptyUserServicePath(t *testing.T) {
	srvs, err := getUserSrvs("/tmp/fakeuser")
	if nil != err {
		t.Fatal(err)
	}
	if len(srvs) > 0 {
		t.Fatal(fmt.Errorf("sanity fail: shouldn't get services from empty directory"))
	}

	dirs, err := ioutil.ReadDir(filepath.Join("/tmp/fakeuser", srvUserPath))
	if nil != err {
		t.Fatal(err)
	}
	if len(dirs) > 0 {
		t.Fatal(fmt.Errorf("sanity fail: shouldn't get listing from empty directory"))
	}

	err = os.RemoveAll("/tmp/fakeuser")
	if nil != err {
		panic("couldn't remove /tmp/fakeuser")
	}
}
