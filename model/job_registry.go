package model

import (
	"fmt"
	"sync"
	"time"
)

// NewRegistry returns a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		all: make(map[string]*Job),
	}
}

// Registry holds all Job Records.
type Registry struct {
	all map[string]*Job
	mu  sync.RWMutex
}

// Add a job.
// Returns false on duplicate or invalid job id.
func (r *Registry) Add(rec *Job) bool {
	if rec == nil || rec.StoreKey() == "" {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.all[rec.StoreKey()]; ok {
		return false
	}

	r.all[rec.StoreKey()] = rec
	//log.Tracef("Adding Job %s => %p", rec.StoreKey(), rec)
	//for k, v := range r.all {
	//	log.Infof("Existig Job [%s] %s => %p", k, v.StoreKey(), v)
	//}

	return true
}

// Map function
func (r *Registry) Map(f func(string, *Job)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range r.all {
		f(k, v)
	}
}

// Len returns length of registry.
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
	//log.Tracef("Delete Job %s ", id)

	return true
}

// Cleanup by job TTR.
// Return number of cleaned jobs.
// TODO: Consider new timeout status & flow
//  - Add batch
func (r *Registry) Cleanup() (num int) {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range r.all {
		end := v.StartAt.Add(time.Duration(v.TTR) * time.Millisecond)
		if (v.TTR > 0) && (now.After(end)) {
			if err := v.Cancel(); err != nil {
				log.Debugf("[TIMEOUT] failed cancel job %s %v StartAt %v, %v", v.Id, err, v.StartAt, err)
			} else {
				log.Tracef("[TIMEOUT] successfully canceled job %s StartAt %v, TTR %v", v.Id, v.StartAt, time.Duration(v.TTR)*time.Millisecond)
			}
			//log.Tracef("Cleanup Job %s => %v", v.StoreKey(), &v)

			delete(r.all, k)
			num += 1
		} else if len(v.Id) < 1 {
			log.Tracef("[EMPTY Job] %v", v)
		}

	}
	return num
}

// GracefullyShutdown is used when we stop the Registry.
// cancel all running & pending job
// return false if we can't cancel any job
func (r *Registry) GracefullyShutdown() bool {
	r.Cleanup()
	r.mu.Lock()
	defer r.mu.Unlock()
	failed := false
	log.Debug("start GracefullyShutdown")
	for k, v := range r.all {
		if !IsTerminalStatus(v.Status) {
			if err := v.Cancel(); err != nil {
				log.Debug(fmt.Sprintf("failed cancel job %s %v", v.Id, err))
				failed = true
			} else {
				log.Debug(fmt.Sprintf("successfully canceled job %s", v.Id))
			}
		}
		delete(r.all, k)
	}
	return failed
}

// Record fetch job by Job ID.
// Follows comma ok idiom
func (r *Registry) Record(jid string) (*Job, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rec, ok := r.all[jid]; ok {
		return rec, true
	}

	return nil, false
}

// Record fetch all jobs.
// Follows comma ok idiom
func (r *Registry) All() []Job {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.all == nil {
		return []Job{}
	}
	res := make([]Job, len(r.all))
	for _, v := range r.all {
		res = append(res, *v)
	}
	return res
}
