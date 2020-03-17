package model

import (
	"sync"
	// "fmt"
	"context"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithFields(logrus.Fields{"package": "model"})
)

type Worker interface {
	// Start the worker with the given context
	Start(context.Context) error
	// Stop the worker
	Stop() error
	//   // Perform a job as soon as possibly
	//   Perform(Job) error
	//   // PerformAt performs a job at a particular time
	//   PerformAt(Job, time.Time) error
	//   // PerformIn performs a job after waiting for a specified amount of time
	//   PerformIn(Job, time.Duration) error
	//   // Register a Handler
	//   Register(string, Handler) error
}

type BaseWorker struct {
	startOnce  sync.Once
	listenerId string
	stop       chan bool
	stopped    chan bool
	jobs       chan Job
}
