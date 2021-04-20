package worker

import (
	"fmt"
	"github.com/weldpua2008/supraworker/job"
	"github.com/weldpua2008/supraworker/model"
	"sync"
	"time"
)

// StartWorker run goroutine for executing commands and reporting to your API
// Note that a WaitGroup must be passed to functions by
// pointer.
func StartWorker(id int, jobs <-chan *model.Job, wg *sync.WaitGroup) {
	// On return, notify the WaitGroup that we're done.
	defer func() {
		wg.Done()
		log.Debugf("Worker %v finished ", id)
	}()

	log.Info(fmt.Sprintf("Starting worker %v", id))
	for j := range jobs {
		log.Tracef("Worker %v received Job %v TTR %v", id, j.Id, time.Duration(j.TTR)*time.Millisecond)
		mu.Lock()
		NumActiveJobs += 1
		mu.Unlock()
		if err := j.Run(); err != nil {
			log.Info(fmt.Sprintf("Job %v failed with %s", j.Id, err))
			if errFlushBuf := j.FlushSteamsBuffer(); errFlushBuf != nil {
				log.Tracef("Job %v failed to flush buffer due %v", j.Id, errFlushBuf)
			}
			_ = j.Failed()
			jobsFailed.Inc()
		} else {
			dur := time.Since(j.StartAt)
			log.Debugf("Job %v finished in %v", j.Id, dur)
			if errFlushBuf := j.FlushSteamsBuffer(); errFlushBuf != nil {
				log.Tracef("Job %v failed to flush buffer due %v", j.Id, errFlushBuf)
			}
			_ = j.Finish()
			jobsSucceeded.Inc()
			jobsDuration.Observe(dur.Seconds())
		}
		mu.Lock()
		NumActiveJobs -= 1
		mu.Unlock()

		jobsProcessed.Inc()
		job.JobsRegistry.Delete(j.StoreKey())

	}
}
