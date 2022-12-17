package daemon

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const EnvName = "XW_DAEMON_IDX"

var runIdx = 0

type Daemon struct {
	LogFile     string
	MaxCount    int
	MaxError    int
	MinExitTime int64
}

func Background(logFile string, isExit bool) (*exec.Cmd, error) {
	runIdx++
	envIdx, err := strconv.Atoi(os.Getenv(EnvName))
	if err != nil {
		envIdx = 0
	}
	if runIdx <= envIdx {
		return nil, nil
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("%s=%d", EnvName, runIdx))

	cmd, err := startProc(os.Args, env, logFile)
	if err != nil {
		log.Println(os.Getpid(), " Start child process error:", err)
		return nil, err
	}
	log.Println(os.Getpid(), " Start child process success:", cmd.Process.Pid)

	if isExit {
		os.Exit(0)
	}

	return cmd, nil
}

func NewDaemon(logFile string) *Daemon {
	return &Daemon{
		LogFile:     logFile,
		MaxCount:    0,
		MaxError:    3,
		MinExitTime: 10,
	}
}

func (d *Daemon) Run() {
	_, _ = Background(d.LogFile, true)

	var t int64
	count := 1
	errNum := 0
	for {
		dInfo := fmt.Sprintf("daemon process(pid:%d; count:%d/%d; errNum:%d/%d):",
			os.Getpid(), count, d.MaxCount, errNum, d.MaxError)
		if errNum > d.MaxError {
			log.Println(dInfo, "Start child process error too many,exit")
			os.Exit(1)
		}
		if d.MaxCount > 0 && count > d.MaxCount {
			log.Println(dInfo, "Too many restarts")
			os.Exit(0)
		}
		count++

		t = time.Now().Unix()
		cmd, err := Background(d.LogFile, false)
		if err != nil {
			log.Println(dInfo, "Start child process err:", err)
			errNum++
			continue
		}

		if cmd == nil {
			log.Printf("child process pid=%d: start", os.Getpid())
			break
		}

		err = cmd.Wait()
		dat := time.Now().Unix() - t
		if dat < d.MinExitTime {
			errNum++
		} else {
			errNum = 0
		}
		log.Printf("%s child process(%d)exit, Ran for %d seconds: %v\n", dInfo, cmd.ProcessState.Pid(), dat, err)
	}
}

func startProc(args, env []string, logFile string) (*exec.Cmd, error) {
	cmd := &exec.Cmd{
		Path:        args[0],
		Args:        args,
		Env:         env,
		SysProcAttr: NewSysProcAttr(),
	}

	if logFile != "" {
		stdout, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666) //nolint:gofumpt
		if err != nil {
			log.Println(os.Getpid(), ": Open log file error", err)
			return nil, err
		}
		cmd.Stderr = stdout
		cmd.Stdout = stdout
	}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
