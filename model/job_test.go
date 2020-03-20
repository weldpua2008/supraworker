package model

import (
	"testing"
    "fmt"
    "os"
    "context"
    "os/exec"
    "reflect"
"runtime"
"strings"
"regexp"
"time"
// "github.com/sirupsen/logrus"

)
type execFunc func(command string, args ...string) *exec.Cmd
// func init(){
//     logrus.SetLevel(logrus.TraceLevel)
// }
func getFakeExecCommand(validator func(string, ...string)) execFunc {
	return func(command string, args ...string) *exec.Cmd {
		validator(command, args...)
		return fakeExecCommand(command, args...)
	}
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func fakeExecCommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}


func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build…/…/_test/job_test.test -test.run=TestHelperProcess --
	// cmd, args := args[3], args[4:]
    cmd:=args[3]

    re := regexp.MustCompile(`echo (.+?)`)
	// Handle the case where args[0] is dir:...
    // TODO: Futher Validation
    if !strings.Contains(cmd, "bash") {
		fmt.Fprintf(os.Stderr, "Expected command to be 'bash'. Got: '%s' %s", cmd, args)
		os.Exit(2)
	}

    exitCode:=127
	switch {
	case strings.Contains(strings.Join(args, " "), "exit 0"):
        exitCode = 0
    }
    // some code here to check arguments perhaps?
    switch {
	case strings.Contains(strings.Join(args, " "), "sleep "):
        time.Sleep(10 * time.Second)
    }
    res := re.FindStringSubmatch(strings.Join(args, " "))
    out:=""
    if len(res) > 1 {
        out = res[1]
    }

    if (len(out) > 0) && (out != string(' ')) {
        fmt.Fprintf(os.Stdout, fmt.Sprintf("'%v'",out))
    }

	os.Exit(exitCode)
}


func GetFunctionName(i interface{}) string {
    return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
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
	// osGetEnv = func(variable string) (out string) { return "/bin/echo" }
    // Override exec.Command
	execCommand = fakeExecCommand
    execCommandContext = fakeExecCommandContext
	defer func() {
        execCommand = exec.Command
        execCommandContext = exec.CommandContext
    }()

    job := NewJob(fmt.Sprintf("job-%v",GetFunctionName(t.Name) ), fmt.Sprintf("echo  &&exit 0"))
	err := job.Run()

	if err != nil {
		t.Errorf("Expected no error in %s, got %v",  GetFunctionName(t.Name), err)
	}
    if job.Status != JOB_STATUS_SUCCESS {
		t.Errorf("Expected %s, got %s", JOB_STATUS_SUCCESS, job.Status)
	}
}

func TestExecuteJobError(t *testing.T) {
	// osGetEnv = func(variable string) (out string) { return "/bin/echo" }
// Override exec.Command
	execCommand = fakeExecCommand
    execCommandContext = fakeExecCommandContext
	defer func() {
        execCommand = exec.Command
        execCommandContext = exec.CommandContext
    }()

    job := NewJob(fmt.Sprintf("job-%v",GetFunctionName(t.Name) ), fmt.Sprintf("echo  &&exit 1"))
	err := job.Run()

	if err == nil {
		t.Errorf("Expected  error, got %v", err)
	}
    if job.Status != JOB_STATUS_ERROR {
        t.Errorf("Expected %s, got %s", JOB_STATUS_ERROR, job.Status)
    }

}
func TestExecuteJobCancel(t *testing.T) {
	// osGetEnv = func(variable string) (out string) { return "/bin/echo" }
// Override exec.Command
	execCommand = fakeExecCommand
    execCommandContext = fakeExecCommandContext
	defer func() {
        execCommand = exec.Command
        execCommandContext = exec.CommandContext
        // logrus.SetLevel(logrus.InfoLevel)

    }()
    done := make(chan bool, 1)
    // logrus.SetLevel(logrus.TraceLevel)

    job := NewJob(fmt.Sprintf("job-TestExecuteJobCancel" ), fmt.Sprintf("echo v && sleep 100 && exit 0"))
    go func() {
        job.TTR = 10000000
           err:=job.Run()
           if err == nil {
       		      t.Errorf("Expected  error, got %v", err)
       	    }
           defer func() { done <- true}()
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
