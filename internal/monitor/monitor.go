package monitor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/vlad-lukyanov/ward/internal/config"
	"github.com/vlad-lukyanov/ward/internal/metrics"
	"github.com/vlad-lukyanov/ward/internal/notification"
)

type serviceState struct {
	restartCount  int
	lastRestartAt time.Time
}

type Monitor struct {
	cfg        *config.Config
	configPath string
	states     map[string]*serviceState
	mu         sync.Mutex
	logger     *log.Logger
	notify     *notification.Notifier
	metrics    *metrics.Metrics
}

func New(cfg *config.Config, configPath string, logger *log.Logger) *Monitor {
	m := &Monitor{
		cfg:        cfg,
		configPath: configPath,
		states:     make(map[string]*serviceState),
		logger:     logger,
		notify: notification.New(
			cfg.Notification.Enabled,
			cfg.Notification.Telegram.BotToken,
			cfg.Notification.Telegram.ChatID,
			logger,
		),
		metrics: metrics.New(cfg.Metrics.Enabled, cfg.Metrics.Listen, logger),
	}

	m.metrics.SetConfigServices(float64(len(cfg.Services)))
	m.metrics.SetCheckInterval(cfg.CheckInterval.Seconds())

	for _, svc := range cfg.Services {
		m.metrics.SetMaxRestarts(svc.Name, svc.MaxRestarts)
		m.metrics.SetCooldownSeconds(svc.Name, svc.RestartCooldown.Seconds())
		m.metrics.SetRestartOnFail(svc.Name, svc.RestartOnFail)
	}

	return m
}

func (m *Monitor) Reload() error {
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cfg = cfg
	m.notify.Reload(
		cfg.Notification.Enabled,
		cfg.Notification.Telegram.BotToken,
		cfg.Notification.Telegram.ChatID,
	)
	m.metrics.SetConfigServices(float64(len(cfg.Services)))
	m.metrics.SetCheckInterval(cfg.CheckInterval.Seconds())

	for _, svc := range cfg.Services {
		m.metrics.SetMaxRestarts(svc.Name, svc.MaxRestarts)
		m.metrics.SetCooldownSeconds(svc.Name, svc.RestartCooldown.Seconds())
		m.metrics.SetRestartOnFail(svc.Name, svc.RestartOnFail)
	}

	known := make(map[string]bool, len(cfg.Services))
	for _, svc := range cfg.Services {
		known[svc.Name] = true
	}

	for name := range m.states {
		if !known[name] {
			delete(m.states, name)
		}
	}

	m.logger.Printf("config reloaded, %d services", len(cfg.Services))
	return nil
}

func (m *Monitor) Run(stop <-chan os.Signal, reload <-chan os.Signal) {
	m.metrics.Start()

	ticker := time.NewTicker(m.cfg.CheckInterval)
	defer ticker.Stop()

	m.check()

	for {
		select {
		case <-ticker.C:
			m.check()
		case <-reload:
			if err := m.Reload(); err != nil {
				m.logger.Printf("reload failed: %v", err)
			}
		case <-stop:
			return
		}
	}
}

func (m *Monitor) check() {
	for _, svc := range m.cfg.Services {
		running := m.isRunning(svc.Name)
		m.metrics.SetServiceStatus(svc.Name, running)

		m.mu.Lock()
		st := m.states[svc.Name]
		m.mu.Unlock()

		if st != nil {
			m.metrics.SetRestartCount(svc.Name, st.restartCount)
			m.metrics.SetLastRestart(svc.Name, st.lastRestartAt)
		} else {
			m.metrics.SetRestartCount(svc.Name, 0)
		}

		if running {
			m.mu.Lock()
			m.states[svc.Name] = &serviceState{}
			m.mu.Unlock()
			m.metrics.SetRestartCount(svc.Name, 0)
			continue
		}

		m.logger.Printf("service %s is not running", svc.Name)

		if !svc.RestartOnFail {
			continue
		}

		m.mu.Lock()
		if st == nil {
			st = &serviceState{}
			m.states[svc.Name] = st
		}
		m.mu.Unlock()

		if st.restartCount >= svc.MaxRestarts {
			m.logger.Printf("service %s exceeded max restarts (%d), skipping", svc.Name, svc.MaxRestarts)
			m.notify.NotifyMaxRestarts(svc.Name, svc.MaxRestarts)
			continue
		}

		if !st.lastRestartAt.IsZero() && time.Since(st.lastRestartAt) < svc.RestartCooldown {
			m.logger.Printf("service %s in cooldown, waiting", svc.Name)
			continue
		}

		if err := m.restart(svc.Name); err != nil {
			m.logger.Printf("failed to restart %s: %v", svc.Name, err)
			m.metrics.IncRestartFailed(svc.Name)
			m.notify.NotifyFailedRestart(svc.Name, err)
			continue
		}

		m.mu.Lock()
		st.restartCount++
		st.lastRestartAt = time.Now()
		m.mu.Unlock()

		m.metrics.IncRestart(svc.Name)
		m.metrics.SetRestartCount(svc.Name, st.restartCount)
		m.metrics.SetLastRestart(svc.Name, st.lastRestartAt)
		m.logger.Printf("restarted %s (attempt %d/%d)", svc.Name, st.restartCount, svc.MaxRestarts)
		m.notify.NotifyRestart(svc.Name, st.restartCount, svc.MaxRestarts)
	}
}

func (m *Monitor) isRunning(name string) bool {
	out, err := exec.Command("rc-service", "-s", name, "status").CombinedOutput()
	if err != nil {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(string(out)))
	return strings.Contains(s, "running") || strings.Contains(s, "started")
}

func (m *Monitor) restart(name string) error {
	out, err := exec.Command("rc-service", name, "restart").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w (%s)", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}
