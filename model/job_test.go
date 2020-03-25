package model

import (
	"fmt"
	"github.com/weldpua2008/supraworker/model/cmdtest"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T) {
	cmdtest.TestHelperProcess(t)
}

func TestTerminalStatus(t *testing.T) {
	for _, terminalStatus := range []string{
		JOB_STATUS_ERROR,
		JOB_STATUS_SUCCESS,
		JOB_STATUS_CANCELED,
	} {
		if !IsTerminalStatus(terminalStatus) {
			t.Errorf("Status %s expected to be terminal", terminalStatus)

		}
	}
}

func TestStreamApi(t *testing.T) {
	want := "{\"job_uid\":\"job-testing.(*common).Name-fm\",\"run_uid\":\"1\",\"extra_run_id\":\"1\",\"msg\":\"'S'\\n\"}"
	var got string
	notifyStdoutSent := make(chan bool)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ReadAll %s", err)
		}
		got = string(fmt.Sprintf("%s", b))
		notifyStdoutSent <- true
	}))
	defer func() {
		srv.Close()
		StreamingAPIURL = ""
		restoreLevel()
		execCommand = exec.Command
		execCommandContext = exec.CommandContext

	}()
	// startTrace()
	StreamingAPIURL = srv.URL
	log.Trace(fmt.Sprintf("StreamingAPIURL  %s", StreamingAPIURL))

	execCommand = cmdtest.FakeExecCommand
	execCommandContext = cmdtest.FakeExecCommandContext
	job := NewJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), fmt.Sprintf("echo S&&exit 0"))
	job.StreamInterval = 1 * time.Millisecond

	err := job.Run()
	if err != nil {
		t.Errorf("Expected no error in %s, got %v", cmdtest.GetFunctionName(t.Name), err)
	}
	select {
	case <-notifyStdoutSent:

	case <-time.After(10 * time.Second):
		t.Errorf("timed out")
	}

	// time.Sleep(100 * time.Millisecond)
	if job.Status != JOB_STATUS_SUCCESS {
		t.Errorf("Expected %s, got %s", JOB_STATUS_SUCCESS, job.Status)
	}

	if got != want {
		t.Errorf("want %s, got %v", want, got)
	}
}

func TestExecuteJobSuccess(t *testing.T) {
	// Override exec.Command
	execCommand = cmdtest.FakeExecCommand
	execCommandContext = cmdtest.FakeExecCommandContext
	defer func() {
		execCommand = exec.Command
		execCommandContext = exec.CommandContext
	}()

	job := NewJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), fmt.Sprintf("echo  &&exit 0"))
	err := job.Run()

	if err != nil {
		t.Errorf("Expected no error in %s, got %v", cmdtest.GetFunctionName(t.Name), err)
	}
	if job.Status != JOB_STATUS_SUCCESS {
		t.Errorf("Expected %s, got %s", JOB_STATUS_SUCCESS, job.Status)
	}
}

func TestExecuteJobError(t *testing.T) {
	// Override exec.Command
	execCommand = cmdtest.FakeExecCommand
	execCommandContext = cmdtest.FakeExecCommandContext
	defer func() {
		execCommand = exec.Command
		execCommandContext = exec.CommandContext
	}()

	job := NewJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), fmt.Sprintf("echo  &&exit 1"))
	err := job.Run()

	if err == nil {
		t.Errorf("Expected  error, got %v", err)
	}
	if job.Status != JOB_STATUS_ERROR {
		t.Errorf("Expected %s, got %s", JOB_STATUS_ERROR, job.Status)
	}

}
func TestExecuteJobCancel(t *testing.T) {
	// Override exec.Command
	execCommand = cmdtest.FakeExecCommand
	execCommandContext = cmdtest.FakeExecCommandContext
	defer func() {
		execCommand = exec.Command
		execCommandContext = exec.CommandContext
	}()
	done := make(chan bool, 1)
	started := make(chan bool, 1)

	// startTrace()
	// defer restoreLevel()
	job := NewJob(fmt.Sprintf("job-TestExecuteJobCancel"), fmt.Sprintf("echo v && sleep 100 && exit 0"))
	go func() {
		job.TTR = 10000000
		started <- true
		err := job.Run()
		if err == nil {

			t.Errorf("Expected  error for job %v\n, got %v", job, err)
		}
		defer func() { done <- true }()
	}()
	<-started
	if job.Status != JOB_STATUS_IN_PROGRESS {
		// t.Errorf("job.Status %v",job.Status)
		time.Sleep(100 * time.Millisecond)
	}
	job.Cancel()
	<-done
	if job.Status != JOB_STATUS_CANCELED {
		t.Errorf("Expected %s, got %s", JOB_STATUS_CANCELED, job.Status)
	}
}

func TestJobFailed(t *testing.T) {
	job := NewJob("echo", "echo")
	if job.Status == JOB_STATUS_ERROR {
		t.Errorf("job.Status '%s' same '%s'", job.Status, JOB_STATUS_ERROR)
	}
	job.Failed()
	got := job.Status
	want := JOB_STATUS_ERROR

	if got != want {
		t.Errorf("got '%s', want '%s'", got, want)
	}
}

func TestJobFinished(t *testing.T) {
	job := NewJob("echo", "echo")
	if job.Status == JOB_STATUS_SUCCESS {
		t.Errorf("job.Status '%s' same '%s'", job.Status, JOB_STATUS_SUCCESS)
	}

	job.Finish()
	got := job.Status
	want := JOB_STATUS_SUCCESS

	if got != want {
		t.Errorf("got '%s', want '%s'", got, want)
	}
}

func TestJobCancel(t *testing.T) {
	job := NewJob("echo", "echo")
	if job.Status == JOB_STATUS_CANCELED {
		t.Errorf("job.Status '%s' same '%s'", job.Status, JOB_STATUS_CANCELED)
	}
	job.Cancel()
	got := job.Status
	want := JOB_STATUS_CANCELED

	if got != want {
		t.Errorf("got '%s', want '%s'", got, want)
	}
}

func TestJobUpdateActivity(t *testing.T) {
	job := NewJob("echo", "echo")
	got := job.LastActivityAt
	job.updatelastActivity()
	want := job.LastActivityAt

	if got == want {
		t.Errorf("got '%s' == want '%s'", got, want)
	}
}

func TestJobUpdateStatus(t *testing.T) {
	job := NewJob("echo", "echo")
	if job.Status == JOB_STATUS_SUCCESS {
		t.Errorf("job.Status '%s' same '%s'", job.Status, JOB_STATUS_PENDING)
	}
	job.updateStatus(JOB_STATUS_SUCCESS)
	got := job.Status

	want := JOB_STATUS_SUCCESS

	if got != want {
		t.Errorf("got '%s', want '%s'", got, want)
	}
}
