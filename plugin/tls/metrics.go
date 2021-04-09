package tls

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	tlsSessionTicketsRotateStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "tls",
		Name:      "session_tickets_rotate_status",
		Help:      "Status of the last tickets rotation.",
	})
	tlsSessionTicketsRotateTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "tls",
		Name:      "session_tickets_rotate_time",
		Help:      "Time when the TLS session tickets were rotated.",
	})
)
