package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"

	model "github.com/weldpua2008/supraworker/model"
)

var (
	log          = logrus.WithFields(logrus.Fields{"package": "job"})
	JobsRegistry = model.NewRegistry()
)

type ApiJobRequest struct {
	Job_status string `json:"job_status"`
	Limit      int64  `json:"limit"`
}

// Example response
// {
//   "job_id": "dbd618f0-a878-e477-7234-2ef24cb85ef6",
//   "jobStatus": "RUNNING",
//   "has_error": false,
//   "error_msg": "",
//   "run_uid": "0f37a129-eb52-96a7-198b-44515220547e",
//   "job_name": "Untitled",
//   "cmd": "su  - hadoop -c 'hdfs ls ''",
//   "parameters": [],
//   "createDate": "1583414512",
//   "lastUpdated": "1583415483",
//   "stopDate": "1586092912",
//   "extra_run_id": "scheduled__2020-03-05T09:21:40.961391+00:00"
// }
type ApiJobResponse struct {
	JobId       string   `json:"job_id"`
	JobStatus   string   `json:"jobStatus"`
	JobName     string   `json:"job_name"`
	RunUID      string   `json:"run_uid"`
	ExtraRunUID string   `json:"extra_run_id"`
	CMD         string   `json:"cmd"`
	Parameters  []string `json:"parameters"`
	CreateDate  string   `json:"createDate"`
	LastUpdated string   `json:"lastUpdated"`
	StopDate    string   `json:"stopDate"`
}

// NewApiJobRequest prepare struct for Jobs for execution request
func NewApiJobRequest() *ApiJobRequest {
	return &ApiJobRequest{
		Job_status: "PENDING",
		Limit:      5,
	}
}

// StartGenerateJobs gorutine for getting jobs from API with internal
// exists on kill
func StartGenerateJobs(jobs chan *model.Job, ctx context.Context, interval time.Duration) error {

	if len(model.FetchNewJobAPIURL) < 1 {
		close(jobs)
		return fmt.Errorf("FetchNewJobAPIURL is undefined")
	}
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
				// job := model.NewJob(fmt.Sprintf("job-%v", j), fmt.Sprintf("for i in {1..20};do for ii in {1..50};do   echo \"$(date) %v\";done;sleep 3;done;exit 0", j))
				c := NewApiJobRequest()
				jsonStr, err := json.Marshal(&c)
				if err != nil {
					log.Tracef("Failed to marshal request due %s", err)
					// sendSucessfully = false
					continue
				} else {
					log.Tracef("Marshal request %v", c)

				}
				log.Trace(fmt.Sprintf("New Job request %s  to %s", model.FetchNewJobAPIMethod, model.FetchNewJobAPIURL))

				req, err := http.NewRequest(model.FetchNewJobAPIMethod,
					model.FetchNewJobAPIURL,
					bytes.NewBuffer(jsonStr))
				req.Header.Set("Content-Type", "application/json")

				// req.Header.Set("X-Custom-Header", "myvalue")
				// req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Tracef("Failed to send request due %s", err)
					// sendSucessfully = false
					continue
				}
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				var jobResponse ApiJobResponse
				err = json.Unmarshal(body, &jobResponse)
				if err != nil {
					log.Tracef("error Unmarshal response: %v due %s", body, err)
					continue
				}

				job := model.NewJob(fmt.Sprintf("%v", jobResponse.JobId), fmt.Sprintf("%s", jobResponse.CMD))
				job.RunUID = jobResponse.RunUID
				job.ExtraRunUID = jobResponse.ExtraRunUID

				job.SetContext(ctx)
				job.TTR = 0
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
	return nil
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
