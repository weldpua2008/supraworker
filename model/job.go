package model

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/mitchellh/go-ps"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// IsTerminalStatus returns true if status is terminal:
// - Failed
// - Canceled
// - Successful
func IsTerminalStatus(status string) bool {
	switch status {
	case JOB_STATUS_ERROR, JOB_STATUS_CANCELED, JOB_STATUS_SUCCESS:
		return true
	}
	return false
}

// StoreKey returns Job unique store key
func StoreKey(Id string, RunUID string, ExtraRunUID string) string {
	return fmt.Sprintf("%s:%s:%s", Id, RunUID, ExtraRunUID)
}

// Job public structure
type Job struct {
	Id                     string        // Identification for Job
	RunUID                 string        // Running identification
	ExtraRunUID            string        // Extra identification
	Priority               int64         // Priority for a Job
	CreateAt               time.Time     // When Job was created
	StartAt                time.Time     // When command started
	LastActivityAt         time.Time     // When job metadata last changed
	Status                 string        // Currently status
	MaxAttempts            int           // Absolute max num of attempts.
	MaxFails               int           // Absolute max number of failures.
	TTR                    uint64        // Time-to-run in Millisecond
	CMD                    string        // Command
	CmdENV                 []string      // Command
	RunAs                  string        // RunAs defines user
	ResetBackPressureTimer time.Duration // how often we will dump the logs
	StreamInterval         time.Duration
	mu                     sync.RWMutex
	exitError              error
	ExitCode               int // Exit code
	cmd                    *exec.Cmd
	ctx                    context.Context

	// params got from your API
	RawParams []map[string]interface{}
	// stream interface
	elements          uint
	notify            chan interface{}
	notifyStopStreams chan interface{}
	notifyLogSent     chan interface{}
	streamsMu         sync.Mutex
	counter           uint
	timeQuote         bool
	// If we should use shell and wrap the command
	UseSHELL   bool
	streamsBuf []string
}

// StoreKey returns StoreKey
func (j *Job) StoreKey() string {
	return StoreKey(j.Id, j.RunUID, j.ExtraRunUID)
}

// GetStatus get job status.
func (j *Job) GetStatus() string {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.Status
}

// updatelastActivity for the Job
func (j *Job) updatelastActivity() {
	j.LastActivityAt = time.Now()
}

// updateStatus job status
func (j *Job) updateStatus(status string) error {
	log.Trace(fmt.Sprintf("Job %s status %s -> %s", j.Id, j.Status, status))
	j.Status = status
	return nil
}

// GetRawParams from all previous calls
func (j *Job) GetRawParams() []map[string]interface{} {

	return j.RawParams
}

// PutRawParams for all next calls
func (j *Job) PutRawParams(params []map[string]interface{}) error {
	j.RawParams = params
	return nil
}

// GetAPIParams for stage from all previous calls
func (j *Job) GetAPIParams(stage string) map[string]string {
	c := make(map[string]string)
	params := GetParamsFromSection(stage, "params")
	for k, v := range params {
		c[k] = v
	}
	resendParamsKeys := GetSliceParamsFromSection(stage, "resend-params")
	// log.Tracef(" GetAPIParams(%s) params params %v\nresend-params %v\n", stage,params,resendParamsKeys)

	for _, resendParamKey := range resendParamsKeys {
		for _, rawVal := range j.RawParams {
			if val, ok := rawVal[resendParamKey]; ok {
				c[resendParamKey] = fmt.Sprintf("%s", val)
			}
		}
	}
	// log.Tracef("GetAPIParams(%s ) c:  %v \n",stage,c)

	return c
}

// Stops the job process and kills all children processes.
func (j *Job) stopProcess() (cancelError error) {
	var processChildren []int
	if j.cmd != nil && j.cmd.Process != nil {
		log.Tracef("[Job %s] Killing main process %v", j.Id, j.cmd.Process.Pid)
		processTree, errTree := NewProcessTree()
		if errTree == nil {
			processChildren = processTree.Get(j.cmd.Process.Pid)
		} else {
			log.Warnf("Can't fetch process tree, got %v", errTree)
		}
		if err := j.cmd.Process.Kill(); err != nil {
			status := j.cmd.ProcessState.Sys().(syscall.WaitStatus)
			exitStatus := status.ExitStatus()
			signaled := status.Signaled()
			signal := status.Signal()
			//cancelError = fmt.Errorf("failed to kill process: %s", err)
			if !signaled && exitStatus == 0 {
				cancelError = fmt.Errorf("unexpected: err %v, exitStatus was %v + signal %s, while running: %s", err, exitStatus, signal, j.CMD)
			}
		}
		if processList, err := ps.Processes(); err == nil {
			for aux := range processList {
				process := processList[aux]
				if ContainsIntInIntSlice(processChildren, process.Pid()) {
					errKill := syscall.Kill(process.Pid(), syscall.SIGTERM)
					log.Tracef("[Job %s] Killing PID: %d --> Name: %s --> ParentPID: %d [%v]", j.Id, process.Pid(), process.Executable(), process.PPid(), errKill)

				}
			}
		}
	}
	return cancelError
}

// Cancel job
// It triggers an update for the your API if it's configured
func (j *Job) Cancel() (cancelError error) {
	j.mu.Lock()

	defer func() {
		if !IsTerminalStatus(j.Status) {
			if errUpdate := j.updateStatus(JOB_STATUS_CANCELED); errUpdate != nil {
				log.Tracef("failed to change job %s status '%s' -> '%s'", j.Id, j.Status, JOB_STATUS_CANCELED)
			}
			j.updatelastActivity()
			stage := "jobs.cancel"
			params := j.GetAPIParams(stage)
			if err, result := DoApiCall(j.ctx, params, stage); err != nil {
				log.Tracef("failed to update api, got: %s and %s", result, err)
			}
		}
		// Fix race condition
		j.mu.Unlock()
	}()

	cancelError = j.stopProcess()

	return cancelError
}

// Failed job flow
// update your API
func (j *Job) Failed() (cancelError error) {
	j.mu.Lock()
	defer func() {
		msg := fmt.Sprintf("[FAILED] Job '%s' is already in terminal state %s", j.Id, j.Status)
		if !IsTerminalStatus(j.Status) {
			msg = fmt.Sprintf("[FAILED] Job '%s' moved to state %s", j.Id, j.Status)
			if errUpdate := j.updateStatus(JOB_STATUS_ERROR); errUpdate != nil {
				msg = fmt.Sprintf("failed to change job %s status '%s' -> '%s'", j.Id, j.Status, JOB_STATUS_ERROR)
			}
			j.updatelastActivity()
		}
		log.Trace(msg)
		stage := "jobs.failed"
		params := j.GetAPIParams(stage)
		if err, result := DoApiCall(j.ctx, params, stage); err != nil {
			log.Tracef("failed to update api, got: %s and %s", result, err)
		}
		// Fix race condition
		j.mu.Unlock()
	}()

	return j.stopProcess()
}

// Appends log stream to the buffer.
// The content of the buffer will be uploaded to API after:
//  - high volume log producers - after j.elements
//	- after buffer is full
//	- after slow log interval
func (j *Job) AppendLogStream(logStream []string) (err error) {
	if j.quotaHit() {
		<-j.notify
		err = j.doSendSteamBuf()
	}
	j.incrementCounter()
	j.streamsMu.Lock()
	j.streamsBuf = append(j.streamsBuf, logStream...)
	j.streamsMu.Unlock()
	return err
}

// count next element
func (j *Job) incrementCounter() {
	j.streamsMu.Lock()
	defer j.streamsMu.Unlock()
	j.counter++
}
// Checks quota for the buffer
// True - need to send
// False - can wait
func (j *Job) quotaHit() bool {
	return (j.counter >= j.elements) || (len(j.streamsBuf) > int(j.elements)) || (j.timeQuote)
}

// Flushes buffer state and resets state of counters.
func (j *Job) resetCounterLoop(ctx context.Context, after time.Duration) {
	ticker := time.NewTicker(after)
	tickerTimeInterval := time.NewTicker(2 * after)
	tickerSlowLogsInterval := time.NewTicker(10 * after)
	defer func() {
		ticker.Stop()
		tickerTimeInterval.Stop()
		tickerSlowLogsInterval.Stop()
	}()
	for {
		select {
		case <-ctx.Done():
			_ = j.doSendSteamBuf()
			return
		case <-j.notifyStopStreams:
			_ = j.doSendSteamBuf()
			return
		case <-ticker.C:
			j.streamsMu.Lock()
			if j.quotaHit() {
				j.timeQuote = false
				j.doNotify()
			}
			j.counter = 0
			j.streamsMu.Unlock()
		case <-tickerTimeInterval.C:
			j.streamsMu.Lock()
			j.timeQuote = true
			j.streamsMu.Unlock()
		// flush Buffer for slow logs
		case <-tickerSlowLogsInterval.C:
			_ = j.doSendSteamBuf()
		}
	}
}

func (j *Job) doNotify() {
	select {
	case j.notify <- struct{}{}:
	default:
	}
}

// FlushSteamsBuffer - empty current job's streams lines
func (j *Job) FlushSteamsBuffer() error {
	return j.doSendSteamBuf()
}

// doSendSteamBuf low-level functions which sends streams to the remote API
// Send stream only if there is something
func (j *Job) doSendSteamBuf() error {
	j.streamsMu.Lock()
	defer j.streamsMu.Unlock()
	if len(j.streamsBuf) > 0 {
		// log.Tracef("doSendSteamBuf for '%v' len '%v' %v\n ", j.Id, len(j.streamsBuf),j.streamsBuf)

		streamsReader := strings.NewReader(strings.Join(j.streamsBuf, ""))
		// update API
		stage := "jobs.logstream"
		params := j.GetAPIParams(stage)
		if urlProvided(stage) {
			// log.Tracef("Using DoApiCall for Streaming")
			params["msg"] = strings.Join(j.streamsBuf, "")
			if errApi, result := DoApiCall(j.ctx, params, stage); errApi != nil {
				log.Tracef("failed to update api, got: %s and %s\n", result, errApi)
			}

		} else {
			var buf bytes.Buffer
			if _, errReadFrom := buf.ReadFrom(streamsReader); errReadFrom != nil {
				log.Tracef("buf.ReadFrom error %v\n", errReadFrom)
			}
			fmt.Printf("Job '%s': %s\n", j.Id, buf.String())
		}
		j.streamsBuf = nil

	}
	return nil
}

// runcmd executes command
// returns error
// supports cancellation
func (j *Job) runcmd() error {
	j.mu.Lock()
	j.StartAt = time.Now()
	j.updatelastActivity()
	ctx, cancel := prepareContext(j.ctx, j.TTR)
	defer cancel()
	// Use shell wrapper
	shell, args := CmdWrapper(j.RunAs, j.UseSHELL, j.CMD)
	j.cmd = execCommandContext(ctx, shell, args...)
	j.cmd.Env = MergeEnvVars(j.CmdENV)
	j.mu.Unlock()

	stdout, err := j.cmd.StdoutPipe()
	if err != nil {
		msg:=fmt.Sprintf("Cannot initial stdout %s\n", err)
		_ = j.AppendLogStream([]string{msg})
		return fmt.Errorf(msg)
	}

	stderr, err := j.cmd.StderrPipe()
	if err != nil {
		msg:=fmt.Sprintf("Cannot initial stderr %s\n", err)
		_ = j.AppendLogStream([]string{msg})
		return fmt.Errorf(msg)
	}

	j.mu.Lock()
	err = j.cmd.Start()
	if errUpdate := j.updateStatus(JOB_STATUS_IN_PROGRESS); errUpdate != nil {
		log.Tracef("failed to change job %s status '%s' -> '%s'", j.Id, j.Status, JOB_STATUS_IN_PROGRESS)
	}
	j.mu.Unlock()
	if err != nil && j.cmd.Process != nil {
		log.Tracef("Run cmd: %v [%v]\n", j.cmd, j.cmd.Process.Pid)
	} else {
		log.Tracef("Run cmd: %v\n", j.cmd)
	}
	// update API
	stage := "jobs.run"
	if errApi, result := DoApiCall(j.ctx, j.GetAPIParams(stage), stage); errApi != nil {
		log.Tracef("failed to update api, got: %s and %s\n", result, errApi)
	}
	if err != nil {
		_ = j.AppendLogStream([]string{fmt.Sprintf("cmd.Start %s\n", err)})
		return fmt.Errorf("cmd.Start, %s", err)
	}
	notifyStdoutSent := make(chan bool, 1)
	notifyStderrSent := make(chan bool, 1)

	// reset backpressure counter
	per := 5 * time.Second
	if j.ResetBackPressureTimer.Nanoseconds() > 0 {
		per = j.ResetBackPressureTimer
	}
	resetCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go j.resetCounterLoop(resetCtx, per)

	// copies stdout/stderr to to streaming API
	copyStd := func(data *io.ReadCloser, processed chan <- bool ) {
		defer func() {
			processed <- true
		}()
		if data == nil {
			return
		}
		stdOutBuf := bufio.NewReader(*data)
		scanner := bufio.NewScanner(stdOutBuf)
		scanner.Split(bufio.ScanLines)

		buf := make([]byte, 0, 64*1024)
		// The second argument to scanner.Buffer() sets the maximum token size.
		// We will be able to scan the stdout as long as none of the lines is
		// larger than 1MB.
		scanner.Buffer(buf, bufio.MaxScanTokenSize)

		for scanner.Scan() {
			if errScan := scanner.Err(); errScan != nil {
				stdOutBuf.Reset(*data)
			}

			msg := scanner.Text()
			_ = j.AppendLogStream([]string{msg, "\n"})
		}

		if scanner.Err() != nil {
			b, err := ioutil.ReadAll(*data)
			if err == nil {
				_ = j.AppendLogStream([]string{string(b), "\n"})
			} else {
				log.Tracef("[Job  %v] Scanner got unexpected failure: %v", j.Id, err)
			}
		}
	}

	// send stdout to streaming API
	go copyStd(&stdout, notifyStdoutSent)

	// send stderr to streaming API
	go copyStd(&stderr, notifyStderrSent)

	<-notifyStdoutSent
	<-notifyStderrSent

	// The returned error is nil if the command runs, has
	// no problems copying stdin, stdout, and stderr,
	// and exits with a zero exit status.
	err = j.cmd.Wait()
	if err != nil {
		status := j.cmd.ProcessState.Sys().(syscall.WaitStatus)
		signaled := status.Signaled()

		if !signaled {
			log.Tracef("cmd.Wait for '%v' returned error: %v", j.Id, err)
		} /* else {
			log.Tracef("Got Signal: %v, while running: %s", status.Signal(), j.CMD)
		}
		*/
	}

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
		_ = j.AppendLogStream([]string{fmt.Sprintf("%s\n", err)})
		j.exitError = err
	}
	if err == nil {
		signaled := ws.Signaled()
		signal := ws.Signal()
		if signaled {
			err = fmt.Errorf("Signal: %v", signal)
			j.exitError = err
		} else if j.Status == JOB_STATUS_CANCELED {
			err = fmt.Errorf("return error for Canceled Job")
		}
	}

	// log.Tracef("The number of goroutines that currently exist.: %v", runtime.NumGoroutine())
	return err
}

// Run job
// return error in case we have exit code greater then 0
func (j *Job) Run() error {
	j.mu.Lock()
	alreadyRunning := j.Status == JOB_STATUS_IN_PROGRESS || IsTerminalStatus(j.Status)
	j.mu.Unlock()
	if alreadyRunning {
		return fmt.Errorf("Cannot start Job %v with status '%s' ", j.Id, j.Status)
	}
	err := j.runcmd()
	j.mu.Lock()
	defer j.mu.Unlock()
	if err!=nil &&j.exitError == nil {
		j.exitError = err
	}
	j.updatelastActivity()
	if !IsTerminalStatus(j.Status) {
		finalStatus := JOB_STATUS_ERROR
		if err == nil {
			finalStatus = JOB_STATUS_SUCCESS
		}

		if errUpdate := j.updateStatus(finalStatus); errUpdate != nil {
			log.Tracef("failed to change job %s status '%s' -> '%s'", j.Id, j.Status, finalStatus)
		}
		log.Tracef("[RUN] Job '%s' is moved to state %s", j.Id, j.Status)
	} else {
		log.Tracef("[RUN] Job '%s' is already in terminal state %s", j.Id, j.Status)
	}
	return err
}

// Finish is triggered when execution is successful.
func (j *Job) Finish() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.updatelastActivity()
	if errUpdate := j.updateStatus(JOB_STATUS_SUCCESS); errUpdate != nil {
		log.Tracef("failed to change job %s status '%s' -> '%s'", j.Id, j.Status, JOB_STATUS_SUCCESS)
	}
	stage := "jobs.finish"
	params := j.GetAPIParams(stage)
	if err, result := DoApiCall(j.ctx, params, stage); err != nil {
		log.Tracef("failed to update api, got: %s and %s", result, err)
	}

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
		RawParams:         make([]map[string]interface{}, 0),
		counter:           0,
		elements:          100,
		UseSHELL:          true,
		RunAs:             "",
		StreamInterval:    time.Duration(5) * time.Second,
	}
}

// NewTestJob return Job with defaults for test
func NewTestJob(id string, cmd string) *Job {
	j := NewJob(id, cmd)
	j.CmdENV = []string{"GO_WANT_HELPER_PROCESS=1"}
	return j
}
