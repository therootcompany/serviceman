package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.rootprojects.org/root/go-serviceman/service"

	ps "github.com/mitchellh/go-ps"
)

// Filled in on init by runner_windows.go
var shellArgs = []string{}

// Notes on spawning a child process
// https://groups.google.com/forum/#!topic/golang-nuts/shST-SDqIp4

// Start will execute the service, and write the PID and logs out to the log directory
func Start(conf *service.Service) error {
	pid := os.Getpid()
	originalBackoff := 1 * time.Second
	maxBackoff := 1 * time.Minute
	threshold := 5 * time.Second

	backoff := originalBackoff
	failures := 0
	logfile := filepath.Join(conf.Logdir, conf.Name+".log")

	if oldPid, exename, err := getProcess(conf); nil == err {
		return fmt.Errorf("%q may already be running as %q (pid %d)", conf.Name, exename, oldPid)
	}

	go func() {
		for {
			maybeWritePidFile(pid, conf)
			time.Sleep(1 * time.Second)
		}
	}()

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
		backgroundCmd(cmd)
		fmt.Fprintf(lf, "[%s] Starting %q %s \n", time.Now(), binpath, strings.Join(args, " "))

		cmd.Stdin = nil
		cmd.Stdout = lf
		cmd.Stderr = lf
		if "" != conf.Workdir {
			cmd.Dir = conf.Workdir
		}
		if len(conf.Envs) > 0 {
			for k, v := range conf.Envs {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
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
			failures++
			fmt.Fprintf(lf, "Waiting %s to restart %q (%d consequtive immediate exits)\n", backoff, conf.Name, failures)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return nil
}

// Stop will find and stop another serviceman runner instance by it's PID
func Stop(conf *service.Service) error {
	i := 0
	var err error
	for {
		if i >= 3 {
			return err
		}
		i++
		oldPid, exename, err2 := getProcess(conf)
		err = err2
		switch err {
		case nil:
			fmt.Printf("killing old process %q with pid %d\n", exename, oldPid)
			err := kill(oldPid)
			if nil != err {
				return err
			}
			return waitForProcessToDie(oldPid)
		case ErrNoPidFile:
			return err
		case ErrNoProcess:
			return err
		case ErrInvalidPidFile:
			fallthrough
		default:
			// waiting a little bit since the PID is written every second
			time.Sleep(400 * time.Millisecond)
		}
	}

	return fmt.Errorf("unexpected error: %s", err)
}

// Restart calls Stop, ignoring any failure, and then Start, returning any failure
func Restart(conf *service.Service) error {
	_ = Stop(conf)
	return Start(conf)
}

var ErrNoPidFile = fmt.Errorf("no pid file")
var ErrInvalidPidFile = fmt.Errorf("malformed pid file")
var ErrNoProcess = fmt.Errorf("process not found by pid")

func waitForProcessToDie(pid int) error {
	exename := "unknown"
	for i := 0; i < 10; i++ {
		px, err := ps.FindProcess(pid)
		if nil != err {
			return nil
		}
		if nil == px {
			return nil
		}
		exename = px.Executable()
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("process %q (%d) just won't die", exename, pid)
}

func getProcess(conf *service.Service) (int, string, error) {
	// TODO make Pidfile() a property of conf?
	pidFile := filepath.Join(conf.Logdir, conf.Name+".pid")
	b, err := ioutil.ReadFile(pidFile)
	if nil != err {
		return 0, "", ErrNoPidFile
	}

	s := strings.TrimSpace(string(b))
	oldPid, err := strconv.Atoi(s)
	if nil != err {
		return 0, "", ErrInvalidPidFile
	}

	px, err := ps.FindProcess(oldPid)
	if nil != err {
		return 0, "", err
	}
	if nil == px {
		return 0, "", ErrNoProcess
	}

	_, err = os.FindProcess(oldPid)
	if nil != err {
		return 0, "", err
	}

	exename := px.Executable()
	return oldPid, exename, nil
}

// TODO error out if can't write to PID or log
func maybeWritePidFile(pid int, conf *service.Service) bool {
	newPid := []byte(strconv.Itoa(pid))

	// TODO use a specific PID dir? meh...
	pidFile := filepath.Join(conf.Logdir, conf.Name+".pid")
	b, err := ioutil.ReadFile(pidFile)
	if nil != err {
		ioutil.WriteFile(pidFile, newPid, 0644)
		return true
	}

	s := strings.TrimSpace(string(b))
	oldPid, err := strconv.Atoi(s)
	if nil != err {
		ioutil.WriteFile(pidFile, newPid, 0644)
		return true
	}

	if oldPid != pid {
		Stop(conf)
		ioutil.WriteFile(pidFile, newPid, 0644)
		return true
	}

	return false
}
