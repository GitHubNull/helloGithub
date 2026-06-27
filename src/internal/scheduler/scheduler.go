package scheduler

import (
	"time"

	"helloGithub/src/internal/config"
	"helloGithub/src/internal/logger"
	"helloGithub/src/internal/scanner"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	scanner   *scanner.Scanner
	cfg       config.ScannerConfig
	domains   []string
	ticker    *time.Ticker
	stopCh    chan struct{}
	log       *logger.Logger
}

// New 创建调度器
func New(cfg config.ScannerConfig, domains []string, sc *scanner.Scanner, log *logger.Logger) *Scheduler {
	return &Scheduler{
		scanner: sc,
		cfg:     cfg,
		domains: domains,
		stopCh:  make(chan struct{}),
		log:     log.WithComponent("scheduler"),
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() {
	s.log.Info("scheduler started", "interval", s.cfg.Interval)

	// 立即执行一次扫描
	go s.scanner.ScanAll(s.domains)

	// 创建定时器
	s.ticker = time.NewTicker(s.cfg.Interval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.log.Info("triggering scheduled scan")
				go s.scanner.ScanAll(s.domains)
			case <-s.stopCh:
				s.ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	close(s.stopCh)
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.log.Info("scheduler stopped")
}
