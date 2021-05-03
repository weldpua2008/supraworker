package worker

import (
	"context"
	"fmt"
	"github.com/weldpua2008/supraworker/job"
	"github.com/weldpua2008/supraworker/metrics"
	"github.com/weldpua2008/supraworker/model"
	"sync"
	"time"
)

// StartWorker run goroutine for executing commands and reporting to your API
// Note that a WaitGroup must be passed to functions by
// pointer.
// There are several scenarios for the Job execution:
//	1). Job execution finished with error/success [Regular flow]
//	2). Cancelled because of TTR [Timeout]
//	3). Cancelled by Job's Registry because of Cleanup process (TTR) [Cancel]
//	4). Cancelled when we fetch external API (cancellation information) [Cancel]

func StartWorker(id int, jobs <-chan *model.Job, wg *sync.WaitGroup) {

	logWorker := log.WithField("worker", id)
	// On return, notify the WaitGroup that we're done.
	defer func() {
		logWorker.Debugf("[FINISHED]")
		metrics.WorkerStatistics.WithLabelValues(
			"finished",
			fmt.Sprintf("worker-%d", id),
		).Inc()
		wg.Done()
	}()

	logWorker.Info("Starting")
	metrics.WorkerStatistics.WithLabelValues(
		"live",
		fmt.Sprintf("worker-%d", id),
	).Inc()
	for j := range jobs {

		logJob := logWorker.WithField("job_id", j.Id)
		ctx := context.WithValue(*j.GetContext(), "worker", id)
		j.SetContext(ctx)

		logJob.Tracef("New Job with TTR %v", time.Duration(j.TTR)*time.Millisecond)
		metrics.WorkerStatistics.WithLabelValues(
			"newjob",
			fmt.Sprintf("worker-%v", id),
		).Inc()
		mu.Lock()
		NumActiveJobs += 1
		mu.Unlock()
		errJobRun := j.Run()
		if errFlushBuf := j.FlushSteamsBuffer(); errFlushBuf != nil {
			logJob.Tracef("failed to flush logstream buffer due %v", errFlushBuf)
		}
		if errJobRun != nil {
			if j.Status != model.JOB_STATUS_CANCELED {
				if errJobRun == context.DeadlineExceeded {
					if errTimeout := j.Timeout(); errTimeout != nil {
						logJob.Tracef("[Timeout()] got: %v ", errTimeout)
					}
				}
				if errFail := j.Failed(); errFail != nil {
					logJob.Tracef("[Failed()] got: %v ", errFail)
				}
				jobsFailed.Inc()
				logJob.Infof("Failed with %s", errJobRun)
			} else {
				logJob.Infof("failed with %s and state %s", errJobRun, j.Status)
			}
		} else {
			dur := time.Since(j.StartAt)
			if err := j.Finish(); err != nil {
				logJob.Debugf("finished in %v got %v", dur, err)
			} else {
				logJob.Debugf("finished in %v", dur)
			}

			jobsSucceeded.Inc()
			jobsDuration.Observe(dur.Seconds())
		}

		//if err := j.Run(); err != nil {
		//	log.Infof("Job %v failed with %s", j.Id, err)
		//	if errFlushBuf := j.FlushSteamsBuffer(); errFlushBuf != nil {
		//		log.Tracef("Job %v failed to flush buffer due %v", j.Id, errFlushBuf)
		//	}
		//	_ = j.Failed()
		//	jobsFailed.Inc()
		//} else {
		//	dur := time.Since(j.StartAt)
		//	log.Debugf("Job %v finished in %v", j.Id, dur)
		//	if errFlushBuf := j.FlushSteamsBuffer(); errFlushBuf != nil {
		//		log.Tracef("Job %v failed to flush buffer due %v", j.Id, errFlushBuf)
		//	}
		//	_ = j.Finish()
		//	jobsSucceeded.Inc()
		//	jobsDuration.Observe(dur.Seconds())
		//}
		mu.Lock()
		NumActiveJobs -= 1
		mu.Unlock()

		jobsProcessed.Inc()
		job.JobsRegistry.Delete(j.StoreKey())

	}
}
