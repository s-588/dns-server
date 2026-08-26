package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec

	DNSQueriesTotal  *prometheus.CounterVec
	DNSQueryDuration *prometheus.HistogramVec
	DNSRecordsFound  *prometheus.CounterVec

	LoginAttemptsTotal *prometheus.CounterVec
	RROperationsTotal  *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		DNSQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dns_queries_total",
				Help: "Total number of DNS queries",
			},
			[]string{"qtype", "rcode"},
		),
		DNSQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dns_query_duration_seconds",
				Help:    "DNS query handling latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"qtype"},
		),
		DNSRecordsFound: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dns_records_found_total",
				Help: "Number of resource records returned for a query",
			},
			[]string{"qtype"},
		),

		LoginAttemptsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "login_attempts_total",
				Help: "Login attempts by result",
			},
			[]string{"result"}, // success | invalid_user | wrong_password | error
		),
		RROperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rr_operations_total",
				Help: "Resource-record CRUD operations",
			},
			[]string{"operation", "result"}, // get/post/patch/delete + success/error
		),
	}
}

// RegisterAll registers the default collectors and every custom metric on the given registry.
func (m *Metrics) RegisterAll(reg *prometheus.Registry) {
	reg.MustRegister(
		// default collectors
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),

		// custom
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.DNSQueriesTotal,
		m.DNSQueryDuration,
		m.DNSRecordsFound,
		m.LoginAttemptsTotal,
		m.RROperationsTotal,
	)
}
