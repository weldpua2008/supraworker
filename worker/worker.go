package worker

import (
	"context"
	"fmt"
	"github.com/weldpua2008/supraworker/job"
	"github.com/weldpua2008/supraworker/metrics"
	"github.com/weldpua2008/supraworker/model"
	"runtime/pprof"
	"sync"
	"time"
)

// StartWorker run goroutine for executing commands and reporting to your API
// Note that a WaitGroup must be passed to functions by
// pointer.
func StartWorker(id int, jobs <-chan *model.Job, wg *sync.WaitGroup) {
	// On return, notify the WaitGroup that we're done.
	// add pprof labels for more useful profiles
	ctx := context.Background()
	defer pprof.SetGoroutineLabels(ctx)
	ctx = pprof.WithLabels(ctx, pprof.Labels("worker", fmt.Sprintf("worker-%d", id)))
	pprof.SetGoroutineLabels(ctx)

	defer func() {
		wg.Done()
		log.Debugf("Worker %d finished ", id)
		metrics.WorkerStatistics.WithLabelValues(
			"finished",
			fmt.Sprintf("worker-%d", id),
		).Inc()
	}()

	log.Infof("Starting worker %d", id)
	metrics.WorkerStatistics.WithLabelValues(
		"live",
		fmt.Sprintf("worker-%d", id),
	).Inc()
	for j := range jobs {
		log.Tracef("Worker %v received Job %v TTR %v", id, j.Id, time.Duration(j.TTR)*time.Millisecond)
		metrics.WorkerStatistics.WithLabelValues(
			"newjob",
			fmt.Sprintf("worker-%v", id),
		).Inc()
		mu.Lock()
		NumActiveJobs += 1
		mu.Unlock()
		errJobRun := j.Run()
		if errFlushBuf := j.FlushSteamsBuffer(); errFlushBuf != nil {
			log.Tracef("Job %v failed to flush buffer due %v", j.Id, errFlushBuf)
		}
		if errJobRun != nil {
			log.Infof("Job %v failed with %s", j.Id, errJobRun)
			_ = j.Failed()
			jobsFailed.Inc()
		} else {
			dur := time.Since(j.StartAt)
			log.Debugf("Job %v finished in %v", j.Id, dur)
			_ = j.Finish()
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
