package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/env"
	"github.com/Lorenta-Tech/kiosk-server/pkg/mail"
)

const (
	serviceName        = "kiosk-server"
	maxTrackedFailures = 1000
)

type Mailer interface {
	Send(to []string, subject string, body string) error
}

type Pinger interface {
	PingContext(ctx context.Context) error
}

type DBConfig struct {
	Enabled           bool
	Interval          time.Duration
	Timeout           time.Duration
	FailureThreshold  int
	RepeatInterval    time.Duration
	Recipients        []string
	NotifyOnRecovered bool
}

type DBMonitor struct {
	db     Pinger
	mailer Mailer
	logger *slog.Logger
	config DBConfig
	host   string

	mu        sync.RWMutex
	failures  int
	down      bool
	downSince time.Time
}

func NewDBMonitor(db Pinger, mailer Mailer, logger *slog.Logger) *DBMonitor {
	return NewDBMonitorWithConfig(db, mailer, logger, loadConfig())
}

func NewDBMonitorWithConfig(db Pinger, mailer Mailer, logger *slog.Logger, config DBConfig) *DBMonitor {
	return &DBMonitor{
		db:     db,
		mailer: mailer,
		logger: logger,
		config: config,
		host:   dbHost(),
	}
}

func loadConfig() DBConfig {
	return DBConfig{
		Enabled:          env.GetBool("DB_MONITOR_ENABLED", true),
		Interval:         time.Duration(env.GetInt("DB_MONITOR_INTERVAL_SECONDS", 30)) * time.Second,
		Timeout:          time.Duration(env.GetInt("DB_MONITOR_TIMEOUT_SECONDS", 5)) * time.Second,
		FailureThreshold: env.GetInt("DB_MONITOR_FAILURE_THRESHOLD", 2),
		RepeatInterval:   time.Duration(env.GetInt("DB_MONITOR_REPEAT_INTERVAL_MINUTES", 30)) * time.Minute,
		Recipients: env.GetList(
			"DB_MONITOR_EMAILS",
			[]string{"suhasdeveloper07@gmail.com"},
		),
		NotifyOnRecovered: env.GetBool("DB_MONITOR_NOTIFY_RECOVERY", true),
	}
}

func dbHost() string {
	raw := env.GetString("DATABASE_URL", "")
	if raw == "" {
		return "unknown"
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "unknown"
	}
	return parsed.Host
}

func (m *DBMonitor) Start(ctx context.Context) {
	if !m.config.Enabled {
		m.logger.Info("db monitor disabled")
		return
	}
	go m.run(ctx)
}

func (m *DBMonitor) run(ctx context.Context) {
	m.logger.Info(
		"db monitor started",
		"interval", m.config.Interval.String(),
		"failure_threshold", m.config.FailureThreshold,
		"recipients", m.config.Recipients,
		"host", m.host,
	)

	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	m.probe(ctx)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("db monitor stopped")
			return
		case <-ticker.C:
			m.probe(ctx)
		}
	}
}

func (m *DBMonitor) probe(ctx context.Context) {
	err := m.ping(ctx)

	m.failures = 0
	m.mu.Lock()
	if err == nil {
		wasDown := m.down
		m.down = false
		if wasDown {
			downSince := m.downSince
			m.downSince = time.Time{}
			m.mu.Unlock()

			m.logger.Info("database recovered", "downtime", time.Since(downSince).String())
			if m.config.NotifyOnRecovered {
				m.notifyRecovered(downSince)
			}
			return
		}
		m.mu.Unlock()
		return
	}

	m.failures++
	failures := m.failures
	alreadyDown := m.down

	if failures > maxTrackedFailures {
		m.failures = maxTrackedFailures
	}

	switch {
	case !alreadyDown && failures < m.config.FailureThreshold:
		m.mu.Unlock()
		m.logger.Warn(
			"database health check failed",
			"error", err,
			"consecutive_failures", failures,
			"threshold", m.config.FailureThreshold,
		)
		return
	case !alreadyDown:
		m.down = true
		m.downSince = time.Now()
		downSince := m.downSince
		m.mu.Unlock()

		m.logger.Error(
			"database is down",
			"error", err,
			"consecutive_failures", failures,
			"host", m.host,
		)
		m.notifyDown(downSince, err, failures)
		return
	}

	downSince := m.downSince
	m.mu.Unlock()

	if !m.shouldRepeat() {
		return
	}

	m.logger.Warn("database still down, repeating alert", "error", err, "downtime", time.Since(downSince).String())
	m.notifyDown(downSince, err, failures)
}

func (m *DBMonitor) shouldRepeat() bool {
	if m.config.RepeatInterval <= 0 {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.downSince) >= m.config.RepeatInterval
}

func (m *DBMonitor) ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()
	return m.db.PingContext(pingCtx)
}

func (m *DBMonitor) notifyDown(downSince time.Time, cause error, failures int) {
	template, err := mail.BuildDBDownTemplate(
		serviceName,
		m.host,
		downSince.Format(time.RFC3339),
		truncate(cause.Error(), 300),
		humanDuration(time.Since(downSince)),
		failures,
	)
	if err != nil {
		m.logger.Error("failed to build db down email", "error", err)
		return
	}

	if err := m.mailer.Send(m.config.Recipients, template.Subject, template.Body); err != nil {
		m.logger.Error("failed to send db down email", "error", err, "recipients", m.config.Recipients)
		return
	}

	m.logger.Info("db down email sent", "recipients", m.config.Recipients, "subject", template.Subject)
}

func (m *DBMonitor) notifyRecovered(downSince time.Time) {
	template, err := mail.BuildDBRecoveredTemplate(
		serviceName,
		m.host,
		downSince.Format(time.RFC3339),
		humanDuration(time.Since(downSince)),
	)
	if err != nil {
		m.logger.Error("failed to build db recovered email", "error", err)
		return
	}

	if err := m.mailer.Send(m.config.Recipients, template.Subject, template.Body); err != nil {
		m.logger.Error("failed to send db recovered email", "error", err, "recipients", m.config.Recipients)
		return
	}

	m.logger.Info("db recovered email sent", "recipients", m.config.Recipients, "subject", template.Subject)
}

func truncate(msg string, limit int) string {
	msg = strings.TrimSpace(msg)
	if len(msg) <= limit {
		return msg
	}
	return msg[:limit] + "..."
}

func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	return fmt.Sprintf("%d hours %d minutes", int(d.Hours()), int(d.Minutes())%60)
}
