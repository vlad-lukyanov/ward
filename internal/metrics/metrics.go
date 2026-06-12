package metrics

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	enabled  bool
	listen   string
	logger   *log.Logger
	registry *prometheus.Registry

	ServiceStatus       *prometheus.GaugeVec
	RestartTotal        *prometheus.CounterVec
	RestartFailed       *prometheus.CounterVec
	RestartCount        *prometheus.GaugeVec
	MaxRestarts         *prometheus.GaugeVec
	LastRestart         *prometheus.GaugeVec
	CooldownSeconds     *prometheus.GaugeVec
	RestartOnFail       *prometheus.GaugeVec
	ConfigServices      prometheus.Gauge
	CheckInterval       prometheus.Gauge
}

func New(enabled bool, listen string, logger *log.Logger) *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		enabled:  enabled,
		listen:   listen,
		logger:   logger,
		registry: reg,

		ServiceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ward_service_status",
				Help: "Current status of service (1=running, 0=stopped)",
			},
			[]string{"service"},
		),

		RestartTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ward_service_restarts_total",
				Help: "Total number of service restarts",
			},
			[]string{"service"},
		),

		RestartFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ward_service_restart_failures_total",
				Help: "Total number of failed service restarts",
			},
			[]string{"service"},
		),

		RestartCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ward_service_restart_count",
				Help: "Current restart count since last successful check",
			},
			[]string{"service"},
		),

		MaxRestarts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ward_service_max_restarts",
				Help: "Configured maximum restart attempts",
			},
			[]string{"service"},
		),

		LastRestart: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ward_service_last_restart_timestamp",
				Help: "Unix timestamp of last restart",
			},
			[]string{"service"},
		),

		CooldownSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ward_service_cooldown_seconds",
				Help: "Configured cooldown between restarts in seconds",
			},
			[]string{"service"},
		),

		RestartOnFail: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ward_service_restart_on_fail",
				Help: "Whether auto-restart is enabled (1=yes, 0=no)",
			},
			[]string{"service"},
		),

		ConfigServices: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ward_config_services",
				Help: "Number of configured services",
			},
		),

		CheckInterval: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ward_check_interval_seconds",
				Help: "Check interval in seconds",
			},
		),
	}

	reg.MustRegister(m.ServiceStatus)
	reg.MustRegister(m.RestartTotal)
	reg.MustRegister(m.RestartFailed)
	reg.MustRegister(m.RestartCount)
	reg.MustRegister(m.MaxRestarts)
	reg.MustRegister(m.LastRestart)
	reg.MustRegister(m.CooldownSeconds)
	reg.MustRegister(m.RestartOnFail)
	reg.MustRegister(m.ConfigServices)
	reg.MustRegister(m.CheckInterval)

	return m
}

func (m *Metrics) Start() error {
	if !m.enabled || m.listen == "" {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	m.logger.Printf("prometheus metrics listening on %s", m.listen)

	go func() {
		if err := http.ListenAndServe(m.listen, mux); err != nil {
			m.logger.Printf("prometheus server error: %v", err)
		}
	}()

	return nil
}

func (m *Metrics) SetServiceStatus(service string, running bool) {
	val := 0.0
	if running {
		val = 1.0
	}
	m.ServiceStatus.WithLabelValues(service).Set(val)
}

func (m *Metrics) IncRestart(service string) {
	m.RestartTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncRestartFailed(service string) {
	m.RestartFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) SetRestartCount(service string, count int) {
	m.RestartCount.WithLabelValues(service).Set(float64(count))
}

func (m *Metrics) SetMaxRestarts(service string, max int) {
	m.MaxRestarts.WithLabelValues(service).Set(float64(max))
}

func (m *Metrics) SetLastRestart(service string, t time.Time) {
	if !t.IsZero() {
		m.LastRestart.WithLabelValues(service).Set(float64(t.Unix()))
	}
}

func (m *Metrics) SetCooldownSeconds(service string, seconds float64) {
	m.CooldownSeconds.WithLabelValues(service).Set(seconds)
}

func (m *Metrics) SetRestartOnFail(service string, enabled bool) {
	val := 0.0
	if enabled {
		val = 1.0
	}
	m.RestartOnFail.WithLabelValues(service).Set(val)
}

func (m *Metrics) SetConfigServices(n float64) {
	m.ConfigServices.Set(n)
}

func (m *Metrics) SetCheckInterval(seconds float64) {
	m.CheckInterval.Set(seconds)
}
