package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	jobsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "supraworker_fetch_jobs_total",
		Help: "The total number of fetched jobs",
	})
	jobsCancelled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "supraworker_jobs_cancelled_total",
		Help: "The total number of CANCELLED jobs",
	})
)
