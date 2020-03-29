package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	// "runtime"

	config "github.com/weldpua2008/supraworker/config"
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

// GetNewJobs fetch from your API the jobs for execution
func GetNewJobs(ctx context.Context) (error, []map[string]interface{}) {

	// localctx, cancel := context.WithCancel(ctx)
	// defer cancel()
	var rawResponseArray []map[string]interface{}
	var rawResponse map[string]interface{}

	// c := NewApiJobRequest()
	t := viper.GetStringMapString("jobs.get.params")
	c := make(map[string]string)
	for k, v := range t {
		var tpl_bytes bytes.Buffer
		tpl := template.Must(template.New("params").Parse(v))
		err := tpl.Execute(&tpl_bytes, config.C)
		if err != nil {
			log.Warn("executing template:", err)
		}
		c[k] = tpl_bytes.String()
		// log.Info(fmt.Sprintf("%s -> %s\n", k, tpl_bytes.String()))
	}
	var req *http.Request
	var err error
	if len(c) > 0 {
		jsonStr, err := json.Marshal(&c)

		if err != nil {
			return fmt.Errorf("Failed to marshal request due %s", err), nil
		}
		log.Trace(fmt.Sprintf("New Job request %s  to %s \nwith %s", model.FetchNewJobAPIMethod, model.FetchNewJobAPIURL, jsonStr))
		// req, err = http.NewRequestWithContext(localctx,
		req, err = http.NewRequest(model.FetchNewJobAPIMethod,
			model.FetchNewJobAPIURL,
			bytes.NewBuffer(jsonStr))

	} else {
		// req, err = http.NewRequestWithContext(localctx,
		req, err = http.NewRequest(model.FetchNewJobAPIMethod,
			model.FetchNewJobAPIURL, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: time.Duration(15 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request due %s", err), nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error read response body got %s", err), nil
	}
	err = json.Unmarshal(body, &rawResponseArray)
	if err != nil {
		err = json.Unmarshal(body, &rawResponse)
		if err != nil {
			return fmt.Errorf("error Unmarshal response: %s due %s", body, err), nil
		}
		rawResponseArray = append(rawResponseArray, rawResponse)
	}
	return nil, rawResponseArray

}

// StartGenerateJobs gorutine for getting jobs from API with internal
// exists on kill
func StartGenerateJobs(jobs chan *model.Job, ctx context.Context, interval time.Duration) error {
	if len(model.FetchNewJobAPIURL) < 1 {
		close(jobs)
		log.Warn("Please provide URL to fetch new Jobs")
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
				// log.Tracef("The number of goroutines that currently exist.: %v", runtime.NumGoroutine())
				if err, jobsData := GetNewJobs(ctx); err == nil {
					var JobId string
					var CMD string
					var RunUID string
					var ExtraRunUID string

					for _, jobResponse := range jobsData {

						for key, value := range jobResponse {
							// log.Infof("k %v, v %v", key, value)
							switch strings.ToLower(key) {
							case "id", "jobid", "job_id", "job_uid":
								JobId = fmt.Sprintf("%v", value)
							case "cmd", "command", "execute":
								CMD = fmt.Sprintf("%v", value)
							case "runid", "runuid", "run_id", "run_uid":
								RunUID = fmt.Sprintf("%v", value)
							case "extrarunid", "extrarunuid", "extrarun_id", "extrarun_uid", "extra_run_id", "extra_run_uid":
								ExtraRunUID = fmt.Sprintf("%v", value)
							}
						}
						if len(JobId) < 1 {
							continue
						}

						job := model.NewJob(fmt.Sprintf("%v", JobId), fmt.Sprintf("%s", CMD))
						job.RunUID = RunUID
						job.ExtraRunUID = ExtraRunUID
						job.RawParams = append(job.RawParams, jobResponse)

						job.SetContext(ctx)
						job.TTR = 0
						if JobsRegistry.Add(job) {
							jobs <- job
							j += 1
							log.Trace(fmt.Sprintf("sent job id %v ", job.Id))
						} else {
							log.Trace(fmt.Sprintf("Duplicated job id %v ", job.Id))
						}
					}
				} else {
					log.Trace(fmt.Sprintf("Failed fetch a new Jobs portion due %v ", err))
				}

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
