package model

import (
	"testing"
	// "time"
)

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
