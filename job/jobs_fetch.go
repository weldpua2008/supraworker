package job

import (
	// "sync"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	model "github.com/weldpua2008/supraworker/model"
	"time"
)

var (
	log          = logrus.WithFields(logrus.Fields{"package": "job"})
	JobsRegistry = model.NewRegistry()
)

// StartGenerateJobs gorutine for getting jobs from API with internal
// exists on kill
func StartGenerateJobs(jobs chan *model.Job, ctx context.Context, interval time.Duration) {
	doneNumJobs := make(chan int, 1)
	doneNumCancelJobs := make(chan int, 1)
	log.Info(fmt.Sprintf("Starting generate jobs with delay %v", interval))
	tickerCancelJobs := time.NewTicker(10 * time.Second)
	tickerGenerateJobs := time.NewTicker(interval)
	defer func() {
		tickerGenerateJobs.Stop()
		tickerCancelJobs.Stop()
	}()

	go func() {
		j := 0
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				doneNumJobs <- j
				if GracefullShutdown(jobs) {
					log.Debug("Jobs generation finished [ SUCESSFULLY ]")
				} else {
					log.Warn("Jobs generation finished [ FAILED ]")
				}

				return
			case <-tickerGenerateJobs.C:
				// example Job
				// job := model.NewJob(fmt.Sprintf("job-%v", j), fmt.Sprintf("echo %v;for i in {1..20};do date;done; for i in {1..5};do date;sleep 5 && echo $(date);done;exit 0", j))
				job := model.NewJob(fmt.Sprintf("job-%v", j), fmt.Sprintf("for i in {1..20};do for ii in {1..50};do   echo \"$(date) %v\";done;sleep 3;done;exit 0", j))

				job.SetContext(ctx)
				job.TTR = 10000000
				if JobsRegistry.Add(job) {
					jobs <- job
					log.Trace(fmt.Sprintf("sent job id %v ", job.Id))

				} else {
					log.Trace(fmt.Sprintf("Duplicated job id %v ", job.Id))
				}
				j += 1
			}
		}
	}()

	// Single gorutine for canceling jobs
	// We are getting such jobs from API
	// exists on kill

	log.Info(fmt.Sprintf("Starting canceling jobs with delay %v", interval))

	go func() {
		j := 0
		for {
			select {
			case <-ctx.Done():
				doneNumCancelJobs <- j
				log.Debug("Jobs cancelation finished [ SUCESSFULLY ]")

				return
			case <-tickerCancelJobs.C:

				n := JobsRegistry.Cleanup()
				if n > 0 {
					j += n
					log.Trace(fmt.Sprintf("Cleared %v/%v jobs", n, j))

				}
			}
		}
	}()

	numSentJobs := <-doneNumJobs
	numCancelJobs := <-doneNumCancelJobs

	log.Info(fmt.Sprintf("Sent %v jobs", numSentJobs))
	if numCancelJobs > 0 {
		log.Info(fmt.Sprintf("Canceled %v jobs", numCancelJobs))
	}

}

// GracefullShutdown cancel all running jobs
// returns error in case any job failed to cancel
func GracefullShutdown(jobs <-chan *model.Job) bool {
	// empty jobs channel
	if len(jobs) > 0 {
		log.Trace(fmt.Sprintf("jobs chan still has size %v, empty it", len(jobs)))
		for len(jobs) > 0 {
			<-jobs
		}
	}
	JobsRegistry.GracefullShutdown()
	if JobsRegistry.Len() > 0 {
		log.Trace(fmt.Sprintf("GracefullShutdown failed, '%v' jobs left ", JobsRegistry.Len()))
		return false
	}
	return true

}
