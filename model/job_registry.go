package model

import (
	"fmt"
	"sync"
	"time"
)

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
	if rec == nil || rec.Id == "" {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
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
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range r.all {
		end := v.StartAt.Add(time.Duration(v.TTL) * time.Second)
		if (end.Before(now.Add(-7 * 24 * time.Hour))) || (end.After(now)) {
			if !IsTerminalStatus(v.Status) {
				if err := v.Cancel(); err != nil {
					log.Debug(fmt.Sprintf("failed cancel job %s %v", v.Id, err))
				} else {
					log.Debug(fmt.Sprintf("sucessfully canceled job %s", v.Id))
				}
			}
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
	log.Debug("start GracefullShutdown")
	for k, v := range r.all {
		if !IsTerminalStatus(v.Status) {
			if err := v.Cancel(); err != nil {
				log.Debug(fmt.Sprintf("failed cancel job %s %v", v.Id, err))
				failed = true
			} else {
				log.Debug(fmt.Sprintf("sucessfully canceled job %s", v.Id))
			}
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
