package ippool

import (
	"net"
	"strings"
	"sync"
	"time"
)

// IPInfo IP信息
type IPInfo struct {
	IP      net.IP
	Latency time.Duration
	LastCheck time.Time
	Available bool
}

// DomainPool 域名对应的IP池
type DomainPool struct {
	mu   sync.RWMutex
	ips  map[string]*IPInfo
	best *IPInfo
}

// Pool IP池管理器
type Pool struct {
	mu      sync.RWMutex
	domains map[string]*DomainPool
}

// New 创建IP池
func New() *Pool {
	return &Pool{
		domains: make(map[string]*DomainPool),
	}
}

// GetDomainPool 获取或创建域名池
func (p *Pool) GetDomainPool(domain string) *DomainPool {
	p.mu.RLock()
	pool, ok := p.domains[domain]
	p.mu.RUnlock()
	if ok {
		return pool
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	pool, ok = p.domains[domain]
	if ok {
		return pool
	}
	pool = &DomainPool{
		ips: make(map[string]*IPInfo),
	}
	p.domains[domain] = pool
	return pool
}

// UpdateIPs 更新域名的IP列表
func (p *Pool) UpdateIPs(domain string, ips []net.IP) {
	pool := p.GetDomainPool(domain)
	pool.mu.Lock()
	defer pool.mu.Unlock()

	newIPs := make(map[string]*IPInfo)
	for _, ip := range ips {
		key := ip.String()
		if existing, ok := pool.ips[key]; ok {
			newIPs[key] = existing
		} else {
			newIPs[key] = &IPInfo{
				IP:        ip,
				Available: false,
			}
		}
	}
	pool.ips = newIPs
	pool.updateBest()
}

// UpdateLatency 更新IP延迟
func (p *Pool) UpdateLatency(domain string, ip net.IP, latency time.Duration, available bool) {
	pool := p.GetDomainPool(domain)
	pool.mu.Lock()
	defer pool.mu.Unlock()

	key := ip.String()
	if info, ok := pool.ips[key]; ok {
		info.Latency = latency
		info.LastCheck = time.Now()
		info.Available = available
		pool.updateBest()
	}
}

// getDomainPoolReadOnly 只读获取域名池（不创建）
func (p *Pool) getDomainPoolReadOnly(domain string) *DomainPool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.domains[domain]
}

// GetBestIP 获取域名对应的最快IP（支持子域名回退匹配）
func (p *Pool) GetBestIP(domain string) (net.IP, bool) {
	// 1. 先精确匹配
	if pool := p.getDomainPoolReadOnly(domain); pool != nil {
		pool.mu.RLock()
		if pool.best != nil && pool.best.Available {
			ip := pool.best.IP
			pool.mu.RUnlock()
			return ip, true
		}
		pool.mu.RUnlock()
	}

	// 2. 尝试父域名回退匹配（如 avatars.githubusercontent.com -> githubusercontent.com）
	p.mu.RLock()
	for d, candidatePool := range p.domains {
		if strings.HasSuffix(domain, "."+d) {
			candidatePool.mu.RLock()
			if candidatePool.best != nil && candidatePool.best.Available {
				ip := candidatePool.best.IP
				candidatePool.mu.RUnlock()
				p.mu.RUnlock()
				return ip, true
			}
			candidatePool.mu.RUnlock()
		}
	}
	p.mu.RUnlock()

	return nil, false
}

// GetAllDomains 获取所有域名
func (p *Pool) GetAllDomains() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	domains := make([]string, 0, len(p.domains))
	for d := range p.domains {
		domains = append(domains, d)
	}
	return domains
}

// GetDomainIPs 获取域名的所有IP信息
func (p *Pool) GetDomainIPs(domain string) []*IPInfo {
	pool := p.GetDomainPool(domain)
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	infos := make([]*IPInfo, 0, len(pool.ips))
	for _, info := range pool.ips {
		infos = append(infos, info)
	}
	return infos
}

// updateBest 更新最优IP（调用方需持有锁）
func (dp *DomainPool) updateBest() {
	var best *IPInfo
	for _, info := range dp.ips {
		if !info.Available {
			continue
		}
		if best == nil || info.Latency < best.Latency {
			best = info
		}
	}
	dp.best = best
}
