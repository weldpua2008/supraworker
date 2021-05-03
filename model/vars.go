package model

import (
	"errors"
	"github.com/sirupsen/logrus"
	"os/exec"
)

const (
	CTX_REQUEST_TIMEOUT = "ctx_req_timeout"
)

var (
	ErrFailedSendRequest = errors.New("Failed to send request")

	execCommandContext = exec.CommandContext
	// FetchNewJobAPIURL is URL for pulling new jobs
	FetchNewJobAPIURL string
	// FetchNewJobAPIMethod is Http METHOD for fetch Jobs API
	FetchNewJobAPIMethod = "POST"
	// FetchNewJobAPIParams is used in eqch requesto for a new job
	FetchNewJobAPIParams = make(map[string]string)

	// StreamingAPIURL is URL for uploading log steams
	StreamingAPIURL string
	// StreamingAPIMethod is Http METHOD for streaming log API
	StreamingAPIMethod = "POST"

	log           = logrus.WithFields(logrus.Fields{"package": "model"})
	previousLevel logrus.Level
)

// Jobber defines a job interface.
type Jobber interface {
	Run() error
	Cancel() error
	Finish() error
}

const (
	JOB_STATUS_PENDING     = "PENDING"
	JOB_STATUS_IN_PROGRESS = "RUNNING"
	JOB_STATUS_SUCCESS     = "SUCCESS"
	JOB_STATUS_ERROR       = "ERROR"
	JOB_STATUS_CANCELED    = "CANCELED"
	JOB_STATUS_TIMEOUT     = "TIMEOUT"
)

func init() {
	previousLevel = logrus.GetLevel()
}
