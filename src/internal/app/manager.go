package app

import (
	"fmt"
	"sync"

	"helloGithub/src/internal/config"
	"helloGithub/src/internal/dns"
	"helloGithub/src/internal/ippool"
	"helloGithub/src/internal/logger"
	"helloGithub/src/internal/scanner"
	"helloGithub/src/internal/scheduler"
)

// ServiceManager 服务管理器，统一管理所有组件生命周期
// 为后续GUI集成提供清晰的Start/Stop/Status接口
type ServiceManager struct {
	cfg       *config.Config
	log       *logger.Logger
	pool      *ippool.Pool
	scanner   *scanner.Scanner
	scheduler *scheduler.Scheduler
	dnsServer *dns.Server

	mu     sync.RWMutex
	status Status
}

// Status 服务状态
type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusRunning
	StatusStopping
)

func (s Status) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	default:
		return "unknown"
	}
}

// NewServiceManager 创建服务管理器
func NewServiceManager(cfg *config.Config) (*ServiceManager, error) {
	log, err := logger.New(logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		OutputPath: cfg.Logger.OutputPath,
	})
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	pool := ippool.New()
	sc := scanner.New(cfg.Scanner, pool, log)
	sch := scheduler.New(cfg.Scanner, cfg.Domains, sc, log)
	dnsServer := dns.New(cfg.DNS, pool, cfg.Upstream, log)

	return &ServiceManager{
		cfg:       cfg,
		log:       log,
		pool:      pool,
		scanner:   sc,
		scheduler: sch,
		dnsServer: dnsServer,
		status:    StatusStopped,
	}, nil
}

// Start 启动所有服务
func (sm *ServiceManager) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.status != StatusStopped {
		return fmt.Errorf("service already %s", sm.status)
	}

	sm.status = StatusStarting
	sm.log.Info("starting all services...")

	// 启动DNS服务器
	go func() {
		if err := sm.dnsServer.Start(); err != nil {
			sm.log.Error("DNS server failed", "error", err)
		}
	}()

	// 启动定时扫描
	sm.scheduler.Start()

	sm.status = StatusRunning
	sm.log.Info("all services started",
		"dns_addr", fmt.Sprintf("%s:%d", sm.cfg.DNS.ListenAddr, sm.cfg.DNS.ListenPort),
	)

	return nil
}

// Stop 停止所有服务
func (sm *ServiceManager) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.status != StatusRunning {
		return fmt.Errorf("service not running, current status: %s", sm.status)
	}

	sm.status = StatusStopping
	sm.log.Info("stopping all services...")

	sm.scheduler.Stop()
	sm.scanner.Stop()
	if err := sm.dnsServer.Stop(); err != nil {
		sm.log.Error("stop DNS server", "error", err)
	}

	sm.status = StatusStopped
	sm.log.Info("all services stopped")

	return nil
}

// Status 获取当前服务状态
func (sm *ServiceManager) Status() Status {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.status
}

// GetPool 获取IP池（供GUI展示数据）
func (sm *ServiceManager) GetPool() *ippool.Pool {
	return sm.pool
}

// GetLogger 获取日志实例
func (sm *ServiceManager) GetLogger() *logger.Logger {
	return sm.log
}

// GetConfig 获取配置
func (sm *ServiceManager) GetConfig() *config.Config {
	return sm.cfg
}

// TriggerScan 手动触发一次扫描（供GUI调用）
func (sm *ServiceManager) TriggerScan() {
	go sm.scanner.ScanAll(sm.cfg.Domains)
}

// GetBestIP 获取指定域名的最优IP（供GUI展示）
func (sm *ServiceManager) GetBestIP(domain string) (string, bool) {
	ip, ok := sm.pool.GetBestIP(domain)
	if !ok {
		return "", false
	}
	return ip.String(), true
}

// GetAllDomains 获取所有监控域名
func (sm *ServiceManager) GetAllDomains() []string {
	return sm.cfg.Domains
}
