package model

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

var osGetEnv = os.Getenv
var execCommand = exec.Command
var execCommandContext = exec.CommandContext


func SwitchExecCommand( f func(command string, args ...string) *exec.Cmd  ) {
    execCommand = f
}
func SwitchExecCommandContext( f func(ctx context.Context, command string, args ...string) *exec.Cmd  ) {
    execCommandContext = f
}

// Jobber defines a job interface.
type Jobber interface {
	Run() error
	Cancel() error
	Finish() error
}

const (
	JOB_STATUS_PENDING          = "pending"
	JOB_STATUS_IN_PROGRESS      = "in_progress"
	JOB_STATUS_SUCCESS          = "success"
	JOB_STATUS_ERROR            = "error"
	JOB_STATUS_CANCEL_REQUESTED = "cancel_requested"
	JOB_STATUS_CANCELED         = "canceled"
)

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
	Priority       int64
	CreateAt       time.Time
	StartAt        time.Time
	LastActivityAt time.Time
	Status         string
	MaxAttempts    int    // Absoulute max num of attempts.
	MaxFails       int    // Absolute max number of failures.
	TTL            uint64 // max time to live in Millisecond
	TTR            uint64 // Time-to-run in seconds
	CMD            string // Comamand
	StreamInterval time.Duration
	mu             sync.RWMutex
	exitError      error
	cmd            *exec.Cmd
	ctx            context.Context
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
	if !IsTerminalStatus(j.Status) {
		log.Trace(fmt.Sprintf("Call Canceled for Job %s", j.Id))
		if j.cmd != nil && j.cmd.Process != nil {
			if err := j.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill process: %s", err)
			}
		}
		j.mu.Lock()
		defer j.mu.Unlock()
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

// Cancel job
// update your API
func (j *Job) SendLogStream(logStream []string) error {
	for _, oneStream := range logStream {
		fmt.Printf("%s", oneStream)
	}
	return nil

}

// runcmd executes command
func (j *Job) runcmd() error {
	// in case we have time limitation or context
	if (j.TTR > 0) || (j.ctx != nil) {
		var ctx context.Context
		var cancel context.CancelFunc
		if j.ctx != nil {
			ctx, cancel = context.WithTimeout(j.ctx, time.Duration(j.TTR)*time.Millisecond)

		} else {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(j.TTR)*time.Millisecond)
		}
		defer cancel()
		j.cmd = execCommandContext(ctx, "bash", "-c", j.CMD)
	} else {
		j.cmd = execCommand("bash", "-c", j.CMD)
	}

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
	if err != nil {
		return fmt.Errorf("cmd.Start, %s", err)
	}
	jobs := make(chan string)
	done := make(chan bool)
	logSent := make(chan bool)
	// parse stdout & stderr
	go func() {
		merged := io.MultiReader(stderr, stdout)
		scanner := bufio.NewScanner(merged)
		defer func() {
			done <- true
			logSent <- true
		}()

		for scanner.Scan() {
			msg := scanner.Text()
			jobs <- fmt.Sprintf("%s\n", msg)
		}

	}()

	// send logs to streaming API
	go func() {
		var logsCache []string

		ticker := time.NewTicker(j.StreamInterval)
		defer ticker.Stop()
		for {
			select {
			case msg := <-jobs:
				logsCache = append(logsCache, msg)
				if len(logsCache) > 10 {
					j.SendLogStream(logsCache)
					logsCache = nil
				}
			case <-done:
				// TODO: catch error
				j.SendLogStream(logsCache)
				return
			case <-ticker.C:
				if len(logsCache) > 0 {
					// TODO: catch error
					j.SendLogStream(logsCache)
					logsCache = nil
				}
			}
		}

	}()
	//
	//
	// buf := bufio.NewReader(stdout) // Notice that this is not in a loop
	// num := 1
	// for {
	// 	// line, _, _ := buf.ReadLine()
	//     line, err := buf.ReadString('\n')
	//     if err == io.EOF {
	//         break
	//     }
	//     if err != nil && err != io.EOF {
	//           return err
	//     }
	//
	// 	num += 1
	// 	fmt.Println(string(line))
	// }
	// if err = cmd.Wait(); err != nil {
	// 	log.Info(fmt.Errorf("cmd.Wait, %s",err))
	// }
	err = j.cmd.Wait()
	<-logSent
	return err
}

// Run job
// return error in case we have exit code greater then 0
func (j *Job) Run() error {

	j.StartAt = time.Now()
	j.updatelastActivity()
	j.updateStatus(JOB_STATUS_IN_PROGRESS)
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

func NewJob(id string, cmd string) *Job {
	return &Job{
		Id:             id,
		CreateAt:       time.Now(),
		StartAt:        time.Now(),
		LastActivityAt: time.Now(),
		Status:         JOB_STATUS_PENDING,
		MaxFails:       1,
		MaxAttempts:    1,
		CMD:            cmd,
		TTL:            1,
		StreamInterval: time.Duration(5) * time.Second,
	}
}
