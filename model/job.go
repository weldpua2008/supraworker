package model

import (
	"os/exec"
	"sync"
	"time"
    "io"
    "fmt"
    "bufio"
)

// Registry holds all Job Records.
type Registry struct {
	all map[string]*Job
	mu  sync.RWMutex
}

// Jobber defines a job interface.
type Jobber interface {
	Run() error
	Cancel() error
	Finish() error
}

const (
	JOB_STATUS_PENDING          = "pending"
	JOB_STATUS_IN_PROGRESS      = "in_progress"
	JOB_STATUS_SUCCESS          = "success"
	JOB_STATUS_ERROR            = "error"
	JOB_STATUS_CANCEL_REQUESTED = "cancel_requested"
	JOB_STATUS_CANCELED         = "canceled"
)

type Job struct {
	Id             string
	Priority       int64
	CreateAt       time.Time
	StartAt        time.Time
	LastActivityAt time.Time
	Status         string
	MaxAttempts    int    // Absoulute max num of attempts.
	MaxFails       int    // Absolute max number of failures.
	TTL            uint64 // max time to live in seconds
	TTR            uint64 // Time-to-run
	CMD            string // Comamand
    StreamInterval time.Duration
	mu             sync.RWMutex
}

func (j *Job) updatelastActivity() {
	j.LastActivityAt = time.Now()
}

func (j *Job) updateStatus(status string) error {
	j.Status = status
	return nil
}

// Cancel job
// update your API
func (j *Job) Cancel() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if (j.Status == JOB_STATUS_PENDING) || (j.Status == JOB_STATUS_IN_PROGRESS) {
		j.updateStatus(JOB_STATUS_CANCELED)
		j.updatelastActivity()
	}
	return nil
}

// Cancel job
// update your API
func (j *Job) SendLogStream(logStream []string) error {
	for _, oneStream := range logStream {
        fmt.Printf("%s",oneStream)
    }
	return nil

}

func (j Job) runcmd() error {

	cmd := exec.Command("bash", "-c", j.CMD)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cmd.StdoutPipe, %s",err)
	}

    stderr  , err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("cmd.StderrPipe, %s",err)
	}
	err =  cmd.Start()
    if err != nil {
    		return fmt.Errorf("cmd.Start, %s",err)
    }
    jobs := make(chan string)
    done := make(chan bool)
    logSent := make(chan bool)
    // parse stdout & stderr
    go func() {
		merged := io.MultiReader(stderr, stdout)
		scanner := bufio.NewScanner(merged)
        defer func() {
    		 done <- true
             logSent <- true
    	}()

		for scanner.Scan() {
			msg := scanner.Text()
            jobs <- fmt.Sprintf("%s\n", msg)
		}
	}()

    // send logs to streaming API
    go func() {
        var logsCache []string

        ticker := time.NewTicker(j.StreamInterval)
        defer ticker.Stop()
        for {
            select {
            case msg := <-jobs :
                    logsCache = append(logsCache, msg)
                    if len(logsCache) > 10 {
                        j.SendLogStream(logsCache )
                        logsCache = nil
                    }
            case <-done:
                // TODO: catch error
                j.SendLogStream(logsCache )
                return
            case  <-ticker.C:
                if len(logsCache) > 0 {
                    // TODO: catch error
                    j.SendLogStream(logsCache )
                    logsCache = nil
                }
            }
        }

    }()
    //
    //
	// buf := bufio.NewReader(stdout) // Notice that this is not in a loop
	// num := 1
	// for {
	// 	// line, _, _ := buf.ReadLine()
    //     line, err := buf.ReadString('\n')
    //     if err == io.EOF {
    //         break
    //     }
    //     if err != nil && err != io.EOF {
    //           return err
    //     }
    //
	// 	num += 1
	// 	fmt.Println(string(line))
	// }
    // if err = cmd.Wait(); err != nil {
	// 	log.Info(fmt.Errorf("cmd.Wait, %s",err))
	// }
    err = cmd.Wait()
    <- logSent
    return err
}

// Run job
// return error in case we have exit code greater then 0
func (j Job) Run() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.StartAt = time.Now()
    err:= j.runcmd()
	j.updatelastActivity()
	j.updateStatus(JOB_STATUS_IN_PROGRESS)
	return err
}

// Finish sucessfull job
// update your API
func (j Job) Finish() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.updatelastActivity()
	j.updateStatus(JOB_STATUS_SUCCESS)
	return nil
}

func NewJob(id string, cmd string) *Job {
	return &Job{
		Id:             id,
		CreateAt:       time.Now(),
		StartAt:        time.Now(),
		LastActivityAt: time.Now(),
		Status:         JOB_STATUS_PENDING,
		MaxFails:       1,
		MaxAttempts:    1,
		CMD:            cmd,
		TTL:            1,
        StreamInterval: time.Duration(5) * time.Second,
	}
}

func NewRegistry() *Registry {
	return &Registry{
		all: make(map[string]*Job),
	}
}

// Add a job.
// Returns false on duplicate or invalid job id.
func (r *Registry) Add(rec *Job) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rec == nil || rec.Id == "" {
		return false
	}

	if _, ok := r.all[rec.Id]; ok {
		return false
	}

	r.all[rec.Id] = rec
	return true
}

// Return length of registry
func (r *Registry) Len() int {
	r.mu.RLock()
	c := len(r.all)
	r.mu.RUnlock()
	return c
}

// Delete a job by job ID.
// Return false if record does not exist.
func (r *Registry) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.all[id]
	if !ok {
		return false
	}
	delete(r.all, id)
	return true
}

// Cleanup by job TTL.
// Return number of cleaned jobs.
func (r *Registry) Cleanup() (num int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for k, v := range r.all {
		end := v.StartAt.Add(time.Duration(v.TTL) * time.Second)
		if (end.Before(now.Add(-7 * 24 * time.Hour))) || (end.After(now)) {
			v.Cancel()
			delete(r.all, k)
			num += 1
		}

	}
	return num
}

// GracefullShutdown
// cancel all running & pending job
// return false if we can't cancel any job
func (r *Registry) GracefullShutdown() bool {
	r.Cleanup()
	r.mu.Lock()
	defer r.mu.Unlock()
	failed := false
	for k, v := range r.all {
		if err := v.Cancel(); err != nil {
			failed = true
		}
		delete(r.all, k)
	}
	return failed
}

// Look up job by ID
// Follows comma ok idiom
func (r *Registry) Record(jid string) (*Job, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rec, ok := r.all[jid]; ok {
		return rec, true
	}

	return nil, false
}
