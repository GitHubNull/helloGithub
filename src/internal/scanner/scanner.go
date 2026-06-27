package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"helloGithub/src/internal/config"
	"helloGithub/src/internal/ippool"
	"helloGithub/src/internal/logger"
)

// Scanner IP扫描器
type Scanner struct {
	cfg      config.ScannerConfig
	pool     *ippool.Pool
	log      *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// New 创建扫描器
func New(cfg config.ScannerConfig, pool *ippool.Pool, log *logger.Logger) *Scanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scanner{
		cfg:    cfg,
		pool:   pool,
		log:    log.WithComponent("scanner"),
		ctx:    ctx,
		cancel: cancel,
	}
}

// ScanAll 扫描所有域名的IP，同时获取GitHub官方IP范围进行测速
func (s *Scanner) ScanAll(domains []string) {
	s.log.Info("starting IP scan", "domains", len(domains))

	// 1. 先扫描各域名对应的IP
	for _, domain := range domains {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		ips, err := s.lookupIPs(domain)
		if err != nil {
			s.log.Error("lookup failed", "domain", domain, "error", err)
			continue
		}

		s.log.Info("domain IPs found", "domain", domain, "count", len(ips))
		s.pool.UpdateIPs(domain, ips)
		s.testDomainIPs(domain, ips)
	}

	// 2. 获取GitHub官方IP范围并测速
	go s.scanGitHubIPRanges(domains)

	s.log.Info("IP scan completed")
}

// scanGitHubIPRanges 从GitHub官方API获取IP范围并测速
func (s *Scanner) scanGitHubIPRanges(domains []string) {
	s.log.Info("fetching GitHub official IP ranges")

	meta, err := FetchGitHubIPRanges(s.cfg.Timeout * 3)
	if err != nil {
		s.log.Error("fetch github meta failed", "error", err)
		return
	}

	// 每个CIDR最多展开256个IP，避免过多
	ips := GetAllIPsFromMeta(meta, 256)
	s.log.Info("GitHub IP ranges fetched", "total_ips", len(ips))

	if len(ips) == 0 {
		return
	}

	// 将官方IP范围关联到主要域名（github.com）进行测速
	primaryDomain := "github.com"
	if len(domains) > 0 {
		primaryDomain = domains[0]
	}

	s.pool.UpdateIPs(primaryDomain, ips)
	s.testDomainIPs(primaryDomain, ips)

	s.log.Info("GitHub IP ranges scan completed", "domain", primaryDomain, "ip_count", len(ips))
}

// lookupIPs 查询域名的所有IP
func (s *Scanner) lookupIPs(domain string) ([]net.IP, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, fmt.Errorf("lookup %s: %w", domain, err)
	}

	// 过滤IPv4地址
	var ipv4s []net.IP
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			ipv4s = append(ipv4s, ipv4)
		}
	}

	return ipv4s, nil
}

// testDomainIPs 测试域名对应的所有IP延迟
func (s *Scanner) testDomainIPs(domain string, ips []net.IP) {
	semaphore := make(chan struct{}, s.cfg.Concurrency)
	var wg sync.WaitGroup

	for _, ip := range ips {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(ip net.IP) {
			defer wg.Done()
			defer func() { <-semaphore }()

			latency, available := s.testIPLatency(ip)
			s.pool.UpdateLatency(domain, ip, latency, available)

			if available {
				s.log.Debug("IP latency test", "domain", domain, "ip", ip, "latency_ms", latency.Milliseconds())
			}
		}(ip)
	}

	wg.Wait()
}

// testIPLatency 测试单个IP的延迟
func (s *Scanner) testIPLatency(ip net.IP) (time.Duration, bool) {
	addr := fmt.Sprintf("%s:%d", ip.String(), s.cfg.TestPort)
	
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, s.cfg.Timeout)
	if err != nil {
		return s.cfg.MaxLatency, false
	}
	defer conn.Close()

	latency := time.Since(start)
	if latency > s.cfg.MaxLatency {
		return s.cfg.MaxLatency, false
	}

	return latency, true
}

// Stop 停止扫描器
func (s *Scanner) Stop() {
	s.cancel()
	s.wg.Wait()
	s.log.Info("scanner stopped")
}
