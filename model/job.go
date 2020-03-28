package model

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// func SwitchExecCommand(f func(command string, args ...string) *exec.Cmd) {
// 	execCommand = f
// }
// func SwitchExecCommandContext(f func(ctx context.Context, command string, args ...string) *exec.Cmd) {
// 	execCommandContext = f
// }

// IsTerminalStatus returns true if status is terminal:
// - Failed
// - Canceled
// - Successfull
func IsTerminalStatus(status string) bool {
	switch status {
	case JOB_STATUS_ERROR, JOB_STATUS_CANCELED, JOB_STATUS_SUCCESS:
		log.Tracef("IsTerminalStatus %s true", status)
		return true
	}
	log.Tracef("IsTerminalStatus %s false", status)
	return false
}

type Job struct {
	Id             string
	RunUID         string
	ExtraRunUID    string
	Priority       int64
	CreateAt       time.Time
	StartAt        time.Time
	LastActivityAt time.Time
	Status         string
	MaxAttempts    int      // Absoulute max num of attempts.
	MaxFails       int      // Absolute max number of failures.
	TTR            uint64   // Time-to-run in Millisecond
	CMD            string   // Comamand
	CmdENV         []string // Comamand

	StreamInterval time.Duration
	mu             sync.RWMutex
	exitError      error
	ExitCode       int
	cmd            *exec.Cmd
	ctx            context.Context

	// params
	RawParams []map[string]interface{}
	// steram interface
	elements          uint
	notify            chan interface{}
	notifyStopStreams chan interface{}
	notifyLogSent     chan interface{}
	stremMu           sync.Mutex
	counter           uint
	timeQuote         bool
	UseSHELL          bool
	streamsBuf        []string
}

func (j *Job) updatelastActivity() {
	j.LastActivityAt = time.Now()
}

func (j *Job) updateStatus(status string) error {
	log.Trace(fmt.Sprintf("Job %s status %s -> %s", j.Id, j.Status, status))
	j.Status = status
	return nil
}

// Cancel job
// update your API
func (j *Job) Cancel() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if !IsTerminalStatus(j.Status) {
		log.Trace(fmt.Sprintf("Call Canceled for Job %s", j.Id))
		if j.cmd != nil && j.cmd.Process != nil {
			if err := j.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill process: %s", err)
			}
		}

		j.updateStatus(JOB_STATUS_CANCELED)
		j.updatelastActivity()
	} else {
		log.Trace(fmt.Sprintf("Job %s in terminal '%s' status ", j.Id, j.Status))
	}
	return nil
}

// Cancel job
// update your API
func (j *Job) Failed() error {
	if !IsTerminalStatus(j.Status) {
		log.Trace(fmt.Sprintf("Call Failed for Job %s", j.Id))

		if j.cmd != nil && j.cmd.Process != nil {
			if err := j.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill process: %s", err)
			}
		}
		j.mu.Lock()
		defer j.mu.Unlock()
		j.updateStatus(JOB_STATUS_ERROR)
		j.updatelastActivity()
	}
	return nil
}

// AppendLogStream for job
// update your API
func (j *Job) AppendLogStream(logStream []string) error {
	if j.quotaHit() {
		<-j.notify
		j.doSendSteamBuf()
	}
	j.incrementCounter()
	j.stremMu.Lock()
	j.streamsBuf = append(j.streamsBuf, logStream...)
	j.stremMu.Unlock()
	return nil
}

//count next element
func (j *Job) incrementCounter() {
	j.stremMu.Lock()
	defer j.stremMu.Unlock()
	j.counter++
}

func (j *Job) quotaHit() bool {
	return (j.counter >= j.elements) || (len(j.streamsBuf) > int(j.elements)) || (j.timeQuote)
}

//scheduled elements counter refresher
func (j *Job) resetCounterLoop(after time.Duration) {
	ticker := time.NewTicker(after)
	tickerTimeInterval := time.NewTicker(2 * after)
	defer func() {
		ticker.Stop()
		tickerTimeInterval.Stop()
		close(j.notifyLogSent)
	}()
	for {
		select {
		case <-j.notifyStopStreams:
			j.doSendSteamBuf()
			j.notifyLogSent <- struct{}{}

			log.Tracef("resetCounterLoop finished for '%v'", j.Id)
			return
		case <-ticker.C:

			j.stremMu.Lock()
			if j.quotaHit() {
				// log.Tracef("doNotify for '%v'", j.Id)
				j.timeQuote = false
				j.doNotify()

			}
			j.counter = 0
			j.stremMu.Unlock()
		case <-tickerTimeInterval.C:

			j.stremMu.Lock()
			j.timeQuote = true
			j.stremMu.Unlock()
		}
	}
}

func (j *Job) doNotify() {
	select {
	case j.notify <- struct{}{}:
	default:
	}
}

func (j *Job) doSendSteamBuf() error {
	log.Tracef("doSendSteamBuf for '%v' len %v", j.Id, len(j.streamsBuf))
	if len(j.streamsBuf) > 0 {
		j.stremMu.Lock()
		defer j.stremMu.Unlock()
		streamsReader := strings.NewReader(strings.Join(j.streamsBuf, ""))
		if len(StreamingAPIURL) > 0 {
			c := struct {
				Job_uid      string `json:"job_uid"`
				Run_uid      string `json:"run_uid"`
				Extra_run_id string `json:"extra_run_id"`
				Msg          string `json:"msg"`
			}{
				Job_uid:      j.Id,
				Run_uid:      j.RunUID,
				Extra_run_id: j.ExtraRunUID,
				Msg:          strings.Join(j.streamsBuf, ""),
			}

			jsonStr, err := json.Marshal(&c)
			if err != nil {
				return fmt.Errorf("Failed to marshal for '%v' due %s", j.Id, err)
			}
			log.Tracef(string(jsonStr))
			req, err := http.NewRequest(StreamingAPIMethod,
				StreamingAPIURL,
				bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			if err != nil {
				return fmt.Errorf("Failed to prepare request '%v' due %s", j.Id, err)
			}
			log.Trace(fmt.Sprintf("New Streaming request %s  to %s from %s", StreamingAPIMethod, StreamingAPIURL, j.Id))

			client := &http.Client{Timeout: time.Duration(15 * time.Second)}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("Failed to sendfor '%v' len %v due %s", j.Id, len(j.streamsBuf), err)
			}
			defer resp.Body.Close()
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				log.Tracef("Response for '%v' %s", j.Id, body)
			}

		} else {
			var buf bytes.Buffer
			buf.ReadFrom(streamsReader)
			fmt.Println(buf.String())
		}
		j.streamsBuf = nil

	}
	return nil
}

// runcmd executes command
func (j *Job) runcmd() error {
	var ctx context.Context
	var cancel context.CancelFunc
	// in case we have time limitation or context
	if (j.TTR > 0) && (j.ctx != nil) {
		ctx, cancel = context.WithTimeout(j.ctx, time.Duration(j.TTR)*time.Millisecond)
	} else if (j.TTR > 0) && (j.ctx == nil) {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(j.TTR)*time.Millisecond)
	} else if j.ctx != nil {
		ctx, cancel = context.WithCancel(j.ctx)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	// Use shell wrapper
	j.mu.Lock()
	if j.UseSHELL {
		shell := "bash"
		args := []string{"-c", j.CMD}
		switch runtime.GOOS {
		case "windows":
			shell = "powershell.exe"
			ps, err := exec.LookPath("powershell.exe")
			if err != nil {
				log.Tracef("Can't fetch powershell %s", err)
				shell = ps
			}
			args = []string{"-NoProfile", "-NonInteractive", j.CMD}
		}
		j.cmd = execCommandContext(ctx, shell, args...)
	} else {
		j.cmd = execCommandContext(ctx, strings.Fields(j.CMD)[0], strings.Fields(j.CMD)[1:]...)
	}
	if len(j.CmdENV) > 0 {
		j.cmd.Env = j.CmdENV
	}
	j.mu.Unlock()

	log.Trace(fmt.Sprintf("Run cmd: %v\n", j.cmd))
	stdout, err := j.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cmd.StdoutPipe, %s", err)
	}

	stderr, err := j.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("cmd.StderrPipe, %s", err)
	}
	err = j.cmd.Start()
	j.mu.Lock()
	j.updateStatus(JOB_STATUS_IN_PROGRESS)
	j.mu.Unlock()
	if err != nil {
		return fmt.Errorf("cmd.Start, %s", err)
	}
	notifyStdoutSent := make(chan bool)
	notifyStderrSent := make(chan bool)

	// reset backpresure counter
	per := 5 * time.Second
	go j.resetCounterLoop(per)

	// parse stdout
	// send logs to streaming API
	go func() {
		defer func() {
			notifyStdoutSent <- true
		}()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			msg := scanner.Text()
			j.AppendLogStream([]string{fmt.Sprintf("%s\n", msg)})
		}
	}()
	// parse stderr
	// send logs to streaming API
	go func() {
		defer func() {
			notifyStderrSent <- true
		}()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			msg := scanner.Text()
			j.AppendLogStream([]string{fmt.Sprintf("%s\n", msg)})
		}
	}()

	// The returned error is nil if the command runs, has
	// no problems copying stdin, stdout, and stderr,
	// and exits with a zero exit status.
	err = j.cmd.Wait()
	if err != nil {
		log.Tracef("cmd.Wait for '%v' returned error: %v", j.Id, err)
	}

	<-notifyStdoutSent
	<-notifyStderrSent
	// signal that we've read all logs
	j.notifyStopStreams <- struct{}{}
	status := j.cmd.ProcessState.Sys()
	ws, ok := status.(syscall.WaitStatus)
	if !ok {
		err = fmt.Errorf("process state Sys() was a %T; want a syscall.WaitStatus", status)
		j.exitError = err
	}
	exitCode := ws.ExitStatus()
	j.ExitCode = exitCode
	if exitCode < 0 {
		err = fmt.Errorf("invalid negative exit status %d", exitCode)
		j.exitError = err
	}
	if exitCode != 0 {
		err = fmt.Errorf("exit code '%d'", exitCode)
		j.exitError = err
	}
	if err == nil {
		signaled := ws.Signaled()
		signal := ws.Signal()
		log.Tracef("Error: %v", err)
		if signaled {
			log.Tracef("Signal: %v", signal)
			err = fmt.Errorf("Signal: %v", signal)
			j.exitError = err
		}
	}
	if err == nil && j.Status == JOB_STATUS_CANCELED {
		err = fmt.Errorf("return error for Canceled Job")
	}
	log.Tracef("The number of goroutines that currently exist.: %v", runtime.NumGoroutine())
	return err
}

// Run job
// return error in case we have exit code greater then 0
func (j *Job) Run() error {
	j.mu.Lock()
	j.StartAt = time.Now()
	j.updatelastActivity()
	j.mu.Unlock()
	err := j.runcmd()
	j.mu.Lock()
	defer j.mu.Unlock()
	j.exitError = err
	j.updatelastActivity()
	if !IsTerminalStatus(j.Status) {
		if err == nil {
			j.updateStatus(JOB_STATUS_SUCCESS)
		} else {
			j.updateStatus(JOB_STATUS_ERROR)
		}
	}
	// <-j.notifyLogSent
	return err
}

// Finish sucessfull job
// update your API
func (j *Job) Finish() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.updatelastActivity()
	j.updateStatus(JOB_STATUS_SUCCESS)
	return nil
}

// SetContext for job
// in case there is time limit for example
func (j *Job) SetContext(ctx context.Context) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.ctx = ctx
}

// NewJob return Job with defaults
func NewJob(id string, cmd string) *Job {
	return &Job{
		Id:                id,
		CreateAt:          time.Now(),
		StartAt:           time.Now(),
		LastActivityAt:    time.Now(),
		Status:            JOB_STATUS_PENDING,
		MaxFails:          1,
		MaxAttempts:       1,
		CMD:               cmd,
		CmdENV:            []string{},
		TTR:               0,
		notify:            make(chan interface{}),
		notifyStopStreams: make(chan interface{}),
		notifyLogSent:     make(chan interface{}),
		counter:           0,
		elements:          100,
		UseSHELL:          true,
		StreamInterval:    time.Duration(5) * time.Second,
	}
}

// NewTestJob return Job with defaults for test
func NewTestJob(id string, cmd string) *Job {
	j := NewJob(id, cmd)
	j.CmdENV = []string{"GO_WANT_HELPER_PROCESS=1"}
	return j
}
