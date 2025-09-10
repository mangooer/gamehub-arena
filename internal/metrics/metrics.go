package metrics

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mangooer/gamehub-arena/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	config *config.MetricsConfig

	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec
	GRPCRequestsTotal    *prometheus.CounterVec
	GRPCRequestDuration  *prometheus.HistogramVec
	GRPCRequestsInFlight *prometheus.GaugeVec

	// 业务指标
	ActiveUsers    prometheus.Gauge
	GameRoomsTotal prometheus.Gauge
	MatchesTotal   prometheus.Counter
	MatchDuration  prometheus.Histogram

	// 系统指标
	DatabaseConnections prometheus.Gauge
	RedisConnections    prometheus.Gauge
	MemoryUsage         prometheus.Gauge
	CPUUsage            prometheus.Gauge
}

func NewMetrics(config *config.MetricsConfig) *Metrics {
	namespace := config.Namespace
	subsystem := config.Subsystem

	m := &Metrics{
		config: config,

		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			}, []string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
			}, []string{"method", "path", "status"},
		),
		HTTPRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_requests_in_flight",
				Help:      "Number of HTTP requests in flight",
			}, []string{"method", "path"},
		),
		GRPCRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_requests_total",
				Help:      "Total number of GRPC requests",
			}, []string{"method", "status"},
		),
		GRPCRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_request_duration_seconds",
				Help:      "GRPC request duration in seconds",
			}, []string{"method", "status"},
		),
		GRPCRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_requests_in_flight",
				Help:      "Number of GRPC requests in flight",
			}, []string{"method"},
		),
		ActiveUsers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "active_users",
				Help:      "Number of active users",
			},
		),
		GameRoomsTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "game_rooms_total",
				Help:      "Number of game rooms",
			},
		),
		MatchesTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "matches_total",
				Help:      "Number of matches",
			},
		),
		MatchDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "match_duration_seconds",
				Help:      "Match duration in seconds",
				Buckets:   []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
			},
		),
		DatabaseConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "database_connections",
				Help:      "Number of database connections",
			},
		),
		RedisConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "redis_connections",
				Help:      "Number of redis connections",
			},
		),
		MemoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
		),
		CPUUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cpu_usage_percentage",
				Help:      "CPU usage in percentage",
			},
		),
	}

	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.GRPCRequestsTotal,
		m.GRPCRequestDuration,
		m.GRPCRequestsInFlight,
		m.ActiveUsers,
		m.GameRoomsTotal,
		m.MatchesTotal,
		m.MatchDuration,
		m.DatabaseConnections,
		m.RedisConnections,
		m.MemoryUsage,
		m.CPUUsage,
	)

	return m
}

func (m *Metrics) Start() error {

	if !m.config.Enabled {
		return nil
	}

	http.Handle(m.config.Path, promhttp.Handler())

	go func() {
		log.Printf("Metrics server listening on port %d", m.config.Port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", m.config.Port), nil); err != nil {
			panic(err)
		}
	}()

	return nil
}

// 记录HTTP请求
func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration time.Duration, inFlight int) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path, strconv.Itoa(status)).Observe(duration.Seconds())
	m.HTTPRequestsInFlight.WithLabelValues(method, path).Set(float64(inFlight))
}

// 记录GRPC请求
func (m *Metrics) RecordGRPCRequest(method, status string, duration time.Duration, inFlight int) {
	m.GRPCRequestsTotal.WithLabelValues(method, status).Inc()
	m.GRPCRequestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
	m.GRPCRequestsInFlight.WithLabelValues(method).Set(float64(inFlight))
}

// 记录数据库连接
func (m *Metrics) RecordDatabaseConnection(connections int) {
	m.DatabaseConnections.Set(float64(connections))
}

// 增加活跃用户数
func (m *Metrics) IncActiveUsers() {
	m.ActiveUsers.Inc()
}

// 减少活跃用户数
func (m *Metrics) DecActiveUsers() {
	m.ActiveUsers.Dec()
}

// 设置游戏房间总数
func (m *Metrics) SetGameRoomsTotal(count float64) {
	m.GameRoomsTotal.Set(count)
}

// 增加匹配总数
func (m *Metrics) IncMatchesTotal() {
	m.MatchesTotal.Inc()
}

// 记录匹配持续时间
func (m *Metrics) RecordMatchDuration(duration time.Duration) {
	m.MatchDuration.Observe(duration.Seconds())
}
