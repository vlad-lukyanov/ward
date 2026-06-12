package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/vlad-lukyanov/ward/internal/config"
	"github.com/vlad-lukyanov/ward/internal/monitor"
)

var version = "dev"

const defaultConfig = `check_interval: 10s
log_file: /var/log/ward.log

notification:
  enabled: false
  telegram:
    bot_token: ""
    chat_id: ""

metrics:
  enabled: false
  listen: ":9090"

services:
  - name: nginx
    restart_on_fail: true
    max_restarts: 3
    restart_cooldown: 30s

  - name: redis
    restart_on_fail: true
    max_restarts: 5
    restart_cooldown: 10s

  - name: sshd
    restart_on_fail: false
`

func usage() {
	fmt.Fprintf(os.Stderr, `ward - OpenRC service watchdog

Usage:
  ward [flags]
  ward <command> [flags]

Commands:
  status    Show status of all monitored services
  list      List all monitored services
  init      Generate default config file

Flags:
  -config    Config file path (default: /etc/ward/config.yaml)
  -version   Print version and exit

Examples:
  ward status
  ward list
  ward init > /etc/ward/config.yaml
  sudo ward -config /etc/ward/config.yaml
`)
}

func getConfigPath() string {
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-config" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return "/etc/ward/config.yaml"
}

func cmdStatus(cfg *config.Config) {
	services := cfg.Services
	if len(services) == 0 {
		fmt.Println("No services configured")
		return
	}

	fmt.Printf("%-20s %-10s %-8s %-10s\n", "SERVICE", "STATUS", "RESTART", "AUTO")
	fmt.Println(strings.Repeat("-", 52))

	for _, svc := range services {
		status := getStatus(svc.Name)
		restarts := "-"
		if svc.RestartOnFail {
			restarts = fmt.Sprintf("0/%d", svc.MaxRestarts)
		}
		auto := "no"
		if svc.RestartOnFail {
			auto = "yes"
		}
		fmt.Printf("%-20s %-10s %-8s %-10s\n", svc.Name, status, restarts, auto)
	}
}

func cmdList(cfg *config.Config) {
	services := cfg.Services
	if len(services) == 0 {
		fmt.Println("No services configured")
		return
	}

	for _, svc := range services {
		status := "monitoring"
		if !svc.RestartOnFail {
			status = "watch only"
		}
		fmt.Printf("%-20s %s\n", svc.Name, status)
	}
}

func cmdInit() {
	fmt.Print(defaultConfig)
}

func getStatus(name string) string {
	out, err := exec.Command("rc-service", "-s", name, "status").CombinedOutput()
	if err != nil {
		return "stopped"
	}
	s := strings.ToLower(strings.TrimSpace(string(out)))
	if strings.Contains(s, "running") || strings.Contains(s, "started") {
		return "running"
	}
	return "stopped"
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "-version", "--version":
		fmt.Println("ward", version)
		return
	case "-help", "--help", "-h":
		usage()
		return
	case "status":
		configPath := getConfigPath()
		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			os.Exit(1)
		}
		cmdStatus(cfg)
		return
	case "list":
		configPath := getConfigPath()
		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			os.Exit(1)
		}
		cmdList(cfg)
		return
	case "init":
		cmdInit()
		return
	}

	configPath := "/etc/ward/config.yaml"
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-config" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
			break
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	var logger *log.Logger
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		logger = log.New(f, "", log.LstdFlags)
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	stop := make(chan os.Signal, 1)
	reload := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reload, syscall.SIGHUP)

	logger.Printf("ward started, checking %d services every %s", len(cfg.Services), cfg.CheckInterval)

	mon := monitor.New(cfg, configPath, logger)
	mon.Run(stop, reload)

	logger.Println("ward stopped")
}
