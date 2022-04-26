package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/zhlls/go-common/version"
)

const DefaultURI = "/metrics"

var (
	inited bool

	currentVersion *prometheus.GaugeVec

	// http metrics
	httpResponseTime *prometheus.SummaryVec
	httpResponseSize *prometheus.SummaryVec
	httpRequestTotal *prometheus.CounterVec
	httpRequestBytes *prometheus.CounterVec

	// grpc metrics
	grpcSentBytes     prometheus.Counter
	grpcReceivedBytes prometheus.Counter
)

func Init(namespace string) {
	currentVersion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "server",
			Name:      "version",
			Help:      "Which version is running. 1 for 'server_version' label with current version.",
		},
		[]string{"version", "revision", "go_version"})

	httpResponseTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "response_time",
			Help:      "Response Time of Each Request in Microsecond.",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpResponseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "response_size",
			Help:      "Response Bytes Size of Each Request.",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "request_total",
			Help:      "Total Number of Each Request.",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "request_bytes",
			Help:      "Total Data Bytes of Each Request.",
		},
		[]string{"method", "endpoint", "status"},
	)

	grpcSentBytes = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "network",
			Name:      "client_grpc_sent_bytes_total",
			Help:      "The total number of bytes sent to grpc clients.",
		})

	grpcReceivedBytes = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "network",
			Name:      "client_grpc_received_bytes_total",
			Help:      "The total number of bytes received from grpc clients.",
		})

	prometheus.MustRegister(
		currentVersion,
		httpResponseTime,
		httpResponseSize,
		httpRequestTotal,
		httpRequestBytes,
		grpcSentBytes,
		grpcReceivedBytes,
	)

	currentVersion.With(prometheus.Labels{
		"version":    version.Info.Version,
		"revision":   version.Info.GitRevision,
		"go_version": version.Info.GolangVersion,
	}).Set(1)

	inited = true
}

// CollectAPIResponseTime collect api resp time
func CollectAPIResponseTime(method, endpoint, status string, value float64) {
	if inited {
		httpResponseTime.WithLabelValues(method, endpoint, status).Observe(value)
	}
}

// CollectAPIResponseSize collect api resp size
func CollectAPIResponseSize(method, endpoint, status string, value float64) {
	if inited {
		httpResponseSize.WithLabelValues(method, endpoint, status).Observe(value)
	}
}

// CollectAPIRequestTotal collect api req total
func CollectAPIRequestTotal(method, endpoint, status string) {
	if inited {
		httpRequestTotal.WithLabelValues(method, endpoint, status).Inc()
	}
}

// CollectAPIRequestBytes collect api req bytes
func CollectAPIRequestBytes(method, endpoint, status string, value float64) {
	if inited {
		httpRequestBytes.WithLabelValues(method, endpoint, status).Add(value)
	}
}

func CollectGRPCSentBytes(value float64) {
	if inited {
		grpcSentBytes.Add(value)
	}
}

func CollectGRPCRecvBytes(value float64) {
	if inited {
		grpcReceivedBytes.Add(value)
	}
}
