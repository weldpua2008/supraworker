package metrics

import (
	// "sync"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FetchCancelLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "supraworker",
		Subsystem: "jobs_cancel",
		Name:      "latency_ns",
		Help:      "The latency distribution of jobs in cancellation flow processed",
	},
		[]string{"type"},
	)
	FetchNewJobLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "suprasched",
		Subsystem: "jobs_fetch",
		Name:      "latency_ns",
		Help:      "The latency distribution of new jobs processed",
	},
		[]string{"type"},
	)
	WorkerStatistics = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "supraworker",
			Subsystem: "worker",
			Name:      "stats",
			Help:      "Statistics of workers.",
		},
		[]string{
			// Type
			"type",
			// What is the Operation?
			"operation",
		},
	)
	RegistryStatistics = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "supraworker",
			Subsystem: "registry",
			Name:      "calls",
			Help:      "Number of API calls to registry by Type.",
		},
		[]string{
			// For example Worker
			"initiator",
			// Of what type is the request?
			"type",
			// What is the Operation?
			"operation",
		},
	)

//
//	ApiCallsStatistics = promauto.NewCounterVec(
//		prometheus.CounterOpts{
//			Namespace: "suprasched",
//			Subsystem: "api",
//			Name:      "calls",
//			Help:      "Number of API calls to 3rd party API partitioned by Type.",
//		},
//		[]string{
//			// For example Amazon
//			"provider",
//			// Which profile is used?
//			"profile",
//			// Of what type is the request?
//			"type",
//			// What is the Operation?
//			"operation",
//		},
//	)
//
//	ReqClustersTerminated = promauto.NewCounterVec(prometheus.CounterOpts{
//		Namespace: "suprasched",
//		Subsystem: "req_clusters",
//		Name:      "terminated_total",
//		Help:      "Number of API calls for Cluster termination to 3rd party API partitioned by Type.",
//	},
//		[]string{
//			// For example Amazon
//			"provider",
//			// Which profile is used?
//			"profile",
//			// Of what type is the request?
//			"type",
//		},
//	)
)
