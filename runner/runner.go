package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"git.rootprojects.org/root/go-serviceman/service"
)

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

	for {
		// setup the log
		lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if nil != err {
			fmt.Fprintf(os.Stderr, "Could not open log file %q\n", logfile)
			lf = os.Stderr
		} else {
			defer lf.Close()
		}

		start := time.Now()
		cmd := exec.Command(binpath, args...)
		cmd.Stdin = nil
		cmd.Stdout = lf
		cmd.Stderr = lf
		if "" != conf.Workdir {
			cmd.Dir = conf.Workdir
		}
		err = cmd.Start()
		if nil != err {
			fmt.Fprintf(lf, "Could not start %q process: %s\n", conf.Name, err)
		} else {
			err = cmd.Wait()
			if nil != err {
				fmt.Fprintf(lf, "Process %q failed with error: %s\n", conf.Name, err)
			} else {
				fmt.Fprintf(lf, "Process %q exited cleanly\n", conf.Name)
				fmt.Printf("Process %q exited cleanly\n", conf.Name)
			}
		}

		// if this is a oneshot... so it is
		if !conf.Restart {
			fmt.Printf("Not restarting %q because `restart` set to `false`\n", conf.Name)
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
