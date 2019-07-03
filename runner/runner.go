package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"git.rootprojects.org/root/go-serviceman/service"
)

// Filled in on init by runner_windows.go
var shellArgs = []string{}

// Notes on spawning a child process
// https://groups.google.com/forum/#!topic/golang-nuts/shST-SDqIp4

func Run(conf *service.Service) {
	originalBackoff := 1 * time.Second
	maxBackoff := 1 * time.Minute
	threshold := 5 * time.Second

	backoff := originalBackoff
	failures := 0
	logfile := filepath.Join(conf.Logdir, conf.Name+".log")

	binpath := conf.Exec
	args := []string{}
	if "" != conf.Interpreter {
		binpath = conf.Interpreter
		args = append(args, conf.Exec)
	}
	args = append(args, conf.Argv...)

	if !conf.System && 0 != len(shellArgs) {
		nargs := append(shellArgs[1:], binpath)
		args = append(nargs, args...)
		binpath = shellArgs[0]
	}

	for {
		// setup the log
		lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if nil != err {
			fmt.Fprintf(os.Stderr, "[%s] Could not open log file %q\n", time.Now(), logfile)
			lf = os.Stderr
		} else {
			defer lf.Close()
		}

		start := time.Now()
		cmd := exec.Command(binpath, args...)
		fmt.Fprintf(lf, "[%s] Starting %q %s \n", time.Now(), binpath, strings.Join(args, " "))

		cmd.Stdin = nil
		cmd.Stdout = lf
		cmd.Stderr = lf
		if "" != conf.Workdir {
			cmd.Dir = conf.Workdir
		}
		err = cmd.Start()
		if nil != err {
			fmt.Fprintf(lf, "[%s] Could not start %q process: %s\n", time.Now(), conf.Name, err)
		} else {
			err = cmd.Wait()
			if nil != err {
				fmt.Fprintf(lf, "[%s] Process %q failed with error: %s\n", time.Now(), conf.Name, err)
			} else {
				fmt.Fprintf(lf, "[%s] Process %q exited cleanly\n", time.Now(), conf.Name)
			}
		}

		// if this is a oneshot... so it is
		if !conf.Restart {
			fmt.Fprintf(lf, "Not restarting %q because `restart` set to `false`\n", conf.Name)
			break
		}

		end := time.Now()
		if end.Sub(start) > threshold {
			backoff = originalBackoff
			failures = 0
		} else {
			failures += 1
			fmt.Fprintf(lf, "Waiting %s to restart %q (%d consequtive immediate exits)\n", backoff, conf.Name, failures)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}
