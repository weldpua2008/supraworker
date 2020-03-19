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
func StartGenerateJobs(jobs chan model.Job, ctx context.Context, interval time.Duration) {
	localdone := make(chan int, 1)
	localCancelDone := make(chan int, 1)
	log.Info(fmt.Sprintf("Starting generate jobs with delay %v seconds", interval))
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		j := 0
		for {
			select {
			case <-ctx.Done():
                close(jobs)
                // empty jobs channel
                if len(jobs) > 0 {
                    log.Trace(fmt.Sprintf("jobs chan still has size %v", len(jobs)))
                    for len(jobs) > 0 {
                          <-jobs
                    }
                }
    			localdone <- j
                if GracefullShutdown(jobs) {
        			log.Info("Finished")
        		} else {
        			log.Info("Failed to shutdown jobs")
        		}

				return
			default:
				// example Job
				job := model.NewJob(fmt.Sprintf("job-%v", j), fmt.Sprintf("echo %v && date&&sleep 5 && echo $(date);exit1", j))
                // ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
                // defer cancel()
                // job.SetContext(&ctx)
                job.SetContext(ctx)
                job.TTR=10000000
                JobsRegistry.Add(job)
				jobs <- *job
				log.Trace(fmt.Sprintf("sent job id %v ", job.Id))
                // time.Sleep(500 *time.Millisecond)
				// time.Sleep(interval)
				j += 1
				// localdone <- j
				// return
				// log.Info(JobsRegistry.Len())

			}
		}
	}()

	// Single gorutine for canceling jobs
	// We are getting such jobs from API
	// exists on kill

	log.Info(fmt.Sprintf("Starting canceling jobs with delay %v seconds", interval))

	go func() {
		j := 0
		for {
			select {
			case <-ctx.Done():
				localCancelDone <- j
				return
			case <-ticker.C:

				n := JobsRegistry.Cleanup()
				if n > 0 {
					j += n
					log.Trace(fmt.Sprintf("Cleared %v/%v jobs", n, j))

				}
			}
		}
	}()

	n := <-localdone
	n1 := <-localCancelDone
	defer ticker.Stop()

    // log.Debug(fmt.Sprintf("Cannel jobs has size %s", len(jobs)))
	time.Sleep(50 * time.Millisecond)
	log.Info(fmt.Sprintf("Sent %v jobs", n))
	if n1 > 0 {
		log.Info(fmt.Sprintf("Canceled %v jobs", n1))

	}

}

// GracefullShutdown cancel all running jobs
// returns error in case any job failed to cancel
func GracefullShutdown(jobs <-chan model.Job) bool {
	JobsRegistry.GracefullShutdown()
	if JobsRegistry.Len() > 0 {
		return false
	}
	return true

}
