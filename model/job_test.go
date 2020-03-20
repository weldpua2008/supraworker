package model

import (
	"fmt"
	"github.com/weldpua2008/supraworker/model/cmdtest"
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
		// logrus.SetLevel(logrus.InfoLevel)
	}()
	done := make(chan bool, 1)
	// logrus.SetLevel(logrus.TraceLevel)

	job := NewJob(fmt.Sprintf("job-TestExecuteJobCancel"), fmt.Sprintf("echo v && sleep 100 && exit 0"))
	go func() {
		job.TTR = 10000000
		err := job.Run()
		if err == nil {
			t.Errorf("Expected  error, got %v", err)
		}
		defer func() { done <- true }()
	}()
	time.Sleep(500 * time.Millisecond)
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
