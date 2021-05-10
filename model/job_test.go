package model

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/weldpua2008/supraworker/config"
	"github.com/weldpua2008/supraworker/model/cmdtest"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TODO:
// Add tests with different HTTP methods

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
	// want := "{\"job_uid\":\"job-testing.(*common).Name-fm\",\"run_uid\":\"1\",\"extra_run_id\":\"1\",\"msg\":\"'S'\\n\"}"
	want := "{\"job_uid\":\"job_uid\",\"msg\":\"'S'\\n\",\"run_uid\":\"1\"}"
	var got string
	notifyStdoutSent := make(chan bool)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "{}")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ReadAll %s", err)
		}
		got = fmt.Sprintf("%s", b)
		notifyStdoutSent <- true
	}))
	defer func() {
		srv.Close()
		// StreamingAPIURL = ""
		restoreLevel()
	}()
	// cmdtest.StartTrace()
	// StreamingAPIURL = srv.URL
	viper.SetConfigType("yaml")
	var yamlExample = []byte(`
    jobs:
      logstream: &update
        url: "` + srv.URL + `"
        method: post
        params:
          "job_uid": "job_uid"
          "run_uid": "1"
    `)
	if err := viper.ReadConfig(bytes.NewBuffer(yamlExample)); err != nil {
		t.Errorf("Can't read config: %v\n", err)
	}

	job := NewTestJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), cmdtest.CMDForTest("echo S&&exit 0"))
	job.StreamInterval = 1 * time.Millisecond
	job.ExtraRunUID = "1"
	job.RunUID = "1"
	err := job.Run()
	if err != nil {
		t.Errorf("Expected no error in %s, got '%v'\n", cmdtest.GetFunctionName(t.Name), err)
	}
	select {
	case <-notifyStdoutSent:
		log.Trace("notifyStdoutSent")
	case <-time.After(10 * time.Second):
		t.Errorf("timed out")
	}

	if job.GetStatus() != JOB_STATUS_RUN_OK {
		t.Errorf("Expected %s, got %s\n", JOB_STATUS_RUN_OK, job.Status)
	}

	if got != want {
		t.Errorf("want %s, got %v", want, got)
	}
}

// TODO: test on 655361
func TestLongLineStreamApi(t *testing.T) {
	repeats := 65531
	// repeats:= 655361
	want := fmt.Sprintf("{\"job_uid\":\"job_uid\",\"msg\":\"'%v'\\n\",\"run_uid\":\"1\"}", strings.Repeat("a", repeats))
	var got string
	notifyStdoutSent := make(chan bool)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "{}")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ReadAll %s", err)
		}
		got = string(fmt.Sprintf("%s", b))
		notifyStdoutSent <- true
	}))
	logrus.SetLevel(logrus.TraceLevel)
	defer func() {
		srv.Close()
		// StreamingAPIURL = ""
		restoreLevel()
	}()
	// cmdtest.StartTrace()
	// StreamingAPIURL = srv.URL
	viper.SetConfigType("yaml")
	var yamlExample = []byte(`
    jobs:
      logstream: &update
        url: "` + srv.URL + `"
        method: post
        params:
          "job_uid": "job_uid"
          "run_uid": "1"
    `)
	if err := viper.ReadConfig(bytes.NewBuffer(yamlExample)); err != nil {
		t.Errorf("Can't read config: %v\n", err)
	}

	job := NewTestJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), cmdtest.CMDForTest(fmt.Sprintf("generate %v&&exit 0", repeats)))
	job.StreamInterval = 1 * time.Millisecond
	job.ExtraRunUID = "1"
	job.RunUID = "1"
	job.ResetBackPressureTimer = time.Duration(450) * time.Millisecond
	err := job.Run()
	if err != nil {
		t.Errorf("Expected no error in %s, got '%v'\n", cmdtest.GetFunctionName(t.Name), err)
	}
	select {
	case <-notifyStdoutSent:
		log.Trace("notifyStdoutSent")
	case <-time.After(5 * time.Second):
		t.Errorf("timed out")
	}

	if job.GetStatus() != JOB_STATUS_RUN_OK {
		t.Errorf("Expected %s, got %s\n", JOB_STATUS_RUN_OK, job.Status)
	}

	if got != want {
		t.Errorf("want %s, got %v", want, got)
	}
}

func TestExecuteJobSuccess(t *testing.T) {
	job := NewTestJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), cmdtest.CMDForTest("echo  &&exit 0"))
	err := job.Run()

	if err != nil {
		t.Errorf("Expected no error in %s, got %v\n", cmdtest.GetFunctionName(t.Name), err)
	}
	if job.GetStatus() != JOB_STATUS_RUN_OK {
		t.Errorf("Expected %s, got %s", JOB_STATUS_RUN_OK, job.Status)
	}
}

func TestExecuteJobError(t *testing.T) {
	job := NewTestJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), cmdtest.CMDForTest("echo  &&exit 1"))
	err := job.Run()

	if err == nil {
		t.Errorf("Expected  error, got %v\n", err)
	}
	if job.GetStatus() != JOB_STATUS_RUN_FAILED {
		t.Errorf("Expected %s, got %s\n", JOB_STATUS_RUN_FAILED, job.Status)
	}

}
func TestExecuteJobCancel(t *testing.T) {
	done := make(chan bool, 1)
	started := make(chan bool, 1)
	job := NewTestJob("job-TestExecuteJobCancel", cmdtest.CMDForTest("echo v && sleep 100 && exit 0"))
	job.TTR = 10000000

	go func() {
		defer func() { done <- true }()
		started <- true
		err := job.Run()
		if err == nil {
			//done <- true
			t.Errorf("Expected  error for job %v\n, got %v\n", job, err)
		}
		// log.Info(err)
		// log.Info("TestExecuteJobCancel finished")
	}()
	<-started
	// time.Sleep(10 * time.Millisecond)
	for job.GetStatus() != JOB_STATUS_IN_PROGRESS {
		// t.Errorf("job.Status %v",job.Status)
		time.Sleep(10 * time.Millisecond)
	}
	if errUpdate := job.Cancel(); errUpdate != nil {
		log.Tracef("failed cancel %s status '%s'", job.Id, job.Status)
	}

	<-done
	//if job.GetStatus() != JOB_STATUS_CANCELED {
	if job.GetStatus() != JOB_STATUS_CANCELED {

		t.Errorf("Expected %s, got %s\n", JOB_STATUS_CANCELED, job.Status)
	}
}

func TestJobFailed(t *testing.T) {
	job := NewTestJob("echo", "echo")
	if job.GetStatus() == JOB_STATUS_ERROR {
		t.Errorf("job.Status '%s' same '%s'\n", job.Status, JOB_STATUS_ERROR)
	}
	if errUpdate := job.Failed(); errUpdate != nil {
		log.Tracef("failed Failed %s status '%s'", job.Id, job.Status)
	}
	got := job.Status
	want := JOB_STATUS_ERROR

	if got != want {
		t.Errorf("got '%s', want '%s'\n", got, want)
	}
}

func TestJobFinished(t *testing.T) {
	job := NewTestJob("echo", "echo")
	if job.GetStatus() == JOB_STATUS_SUCCESS {
		t.Errorf("job.Status '%s' same '%s'\n", job.Status, JOB_STATUS_SUCCESS)
	}

	if errUpdate := job.Finish(); errUpdate != nil {
		log.Tracef("failed Finish %s status '%s'", job.Id, job.Status)
	}

	got := job.Status
	want := JOB_STATUS_SUCCESS

	if got != want {
		t.Errorf("got '%s', want '%s'\n", got, want)
	}
}

func TestJobCancel(t *testing.T) {
	job := NewTestJob("echo", "echo")
	if job.GetStatus() == JOB_STATUS_CANCELED {
		t.Errorf("job.Status '%s' same '%s'\n", job.Status, JOB_STATUS_CANCELED)
	}
	if errUpdate := job.Cancel(); errUpdate != nil {
		log.Tracef("failed Cancel %s status '%s'", job.Id, job.Status)
	}

	got := job.Status
	want := JOB_STATUS_CANCELED

	if got != want {
		t.Errorf("got '%s', want '%s'\n", got, want)
	}
}

func TestJobUpdateActivity(t *testing.T) {
	job := NewTestJob("echo", "echo")
	got := job.LastActivityAt
	job.updatelastActivity()
	want := job.LastActivityAt

	if got == want {
		t.Errorf("got '%s' == want '%s'\n", got, want)
	}
}

func TestJobUpdateStatus(t *testing.T) {
	job := NewTestJob("echo", "echo")
	if job.GetStatus() == JOB_STATUS_SUCCESS {
		t.Errorf("job.Status '%s' same '%s'\n", job.Status, JOB_STATUS_PENDING)
	}
	_ = job.updateStatus(JOB_STATUS_SUCCESS)
	got := job.Status

	want := JOB_STATUS_SUCCESS

	if got != want {
		t.Errorf("got '%s', want '%s'\n", got, want)
	}
}

// Test Job status updates via HTTP Communicator
func TestJobStatusCommunicators(t *testing.T) {
	isFlakyNow := make(chan bool, 1)
	out := make(chan string, 1)
	cmdtest.StartTrace()
	srv := config.NewFlakyTestServer(t, out, isFlakyNow, 15*time.Second)
	defer func() {
		srv.Close()
		close(out)
		cmdtest.RestoreLevel()
	}()

	var yamlString = []byte(`
    ClientId: "Test"
    version: "1.0"
    jobs:
      run: &run
        communicator: &comm
          url: "` + srv.URL + `/"
          method: get
          params:
            status: "{{ .Status }}"
            job_id: "{{ .JobId }}"
            run_id: "{{ .RunUID}}"
            extra_run_id: "{{.ExtraRunUID}}"
            previous_job_status: "{{.PreviousStatus}}"
          codes:
          - 200
          - 201
          backoff:
            maxelapsedtime: 30s
            maxinterval: 20s
            initialinterval: 0s
      failed: &failed
        communicator:
          <<: *comm
      finish: &failed
        communicator:
          <<: *comm
      timeout: &timeout
        communicator:
          <<: *comm
      cancel: &cancel
        communicator:
          <<: *comm
    `)

	C, tmpC := config.StringToCfgForTests(t, yamlString)
	config.C = C
	defer func() {
		config.C = tmpC
	}()
	outRun := `{ "status": "` + JOB_STATUS_IN_PROGRESS + `" }`
	cases := []struct {
		section string
		in      string
		out     string
		cmd     string
	}{
		{
			out:     `{ "status": "` + JOB_STATUS_ERROR + `" }`,
			section: JOB_STATUS_ERROR,
			cmd:     "echo S&&exit 1",
		},
		{
			out:     `{ "status": "` + JOB_STATUS_SUCCESS + `" }`,
			section: JOB_STATUS_SUCCESS,
			cmd:     "echo S&&exit 0",
		},
		{
			out:     `{ "status": "` + JOB_STATUS_CANCELED + `" }`,
			section: JOB_STATUS_CANCELED,
			cmd:     "echo S&&exit 0",
		},
		{
			out:     `{ "status": "` + JOB_STATUS_TIMEOUT + `" }`,
			section: JOB_STATUS_TIMEOUT,
			cmd:     "echo S&&exit 0",
		},
	}
	for _, tc := range cases {
		isFlakyNow <- true
		job := NewTestJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), cmdtest.CMDForTest("echo S&&exit 0"))
		job.StreamInterval = 1 * time.Millisecond
		job.ExtraRunUID = "1"
		job.RunUID = "1"
		out <- outRun
		err := job.Run()
		if err != nil {
			t.Fatalf("Expected no error in %s, got '%v'\n", cmdtest.GetFunctionName(t.Name), err)
		}
		isFlakyNow <- true

		out <- tc.out
		switch tc.section {
		case JOB_STATUS_ERROR:
			if errFail := job.Failed(); errFail != nil {
				t.Fatalf("[Failed()] got: %v ", errFail)
			}
		case JOB_STATUS_SUCCESS:
			if errFail := job.Finish(); errFail != nil {
				t.Fatalf("[Finish()] got: %v ", errFail)
			}
		case JOB_STATUS_CANCELED:
			if errFail := job.Cancel(); errFail != nil {
				t.Fatalf("[Cancel()] got: %v ", errFail)
			}
		case JOB_STATUS_TIMEOUT:
			if errFail := job.Timeout(); errFail != nil {
				t.Fatalf("[Timeout()] got: %v ", errFail)
			}
		}
	}
}
