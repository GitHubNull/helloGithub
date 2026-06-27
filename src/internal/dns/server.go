package dns

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"

	"helloGithub/src/internal/config"
	"helloGithub/src/internal/ippool"
	"helloGithub/src/internal/logger"
)

// Server DNS代理服务器
type Server struct {
	cfg      config.DNSConfig
	pool     *ippool.Pool
	log      *logger.Logger
	upstream []string
	server   *dns.Server
}

// New 创建DNS服务器
func New(cfg config.DNSConfig, pool *ippool.Pool, upstream []string, log *logger.Logger) *Server {
	return &Server{
		cfg:      cfg,
		pool:     pool,
		upstream: upstream,
		log:      log.WithComponent("dns"),
	}
}

// Start 启动DNS服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.ListenAddr, s.cfg.ListenPort)

	s.server = &dns.Server{
		Addr:      addr,
		Net:       "udp",
		Handler:   dns.HandlerFunc(s.handleRequest),
		UDPSize:   65535,
	}

	s.log.Info("starting DNS server", "addr", addr)
	return s.server.ListenAndServe()
}

// Stop 停止DNS服务器
func (s *Server) Stop() error {
	s.log.Info("stopping DNS server")
	if s.server != nil {
		return s.server.Shutdown()
	}
	return nil
}

// handleRequest 处理DNS请求
func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Rcode = dns.RcodeFormatError
		w.WriteMsg(m)
		return
	}

	question := r.Question[0]
	domain := strings.TrimSuffix(question.Name, ".")

	s.log.Debug("DNS query", "domain", domain, "type", dns.TypeToString[question.Qtype])

	// 只处理A记录查询
	if question.Qtype != dns.TypeA {
		s.forwardQuery(w, r)
		return
	}

	// 检查是否是GitHub域名（支持子域名匹配）
	if ip, ok := s.pool.GetBestIP(domain); ok {
		s.log.Debug("returning optimized IP", "domain", domain, "ip", ip.String())
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = false
		m.Answer = append(m.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			A: ip,
		})
		w.WriteMsg(m)
		return
	}

	// 非GitHub域名，转发到上游DNS
	s.forwardQuery(w, r)
}

// forwardQuery 转发DNS查询到上游服务器
func (s *Server) forwardQuery(w dns.ResponseWriter, r *dns.Msg) {
	for _, upstream := range s.upstream {
		c := &dns.Client{
			Timeout: 3 * time.Second,
		}
		m, _, err := c.Exchange(r, upstream)
		if err != nil {
			s.log.Debug("upstream query failed", "server", upstream, "error", err)
			continue
		}
		w.WriteMsg(m)
		return
	}

	// 所有上游都失败，返回SERVFAIL
	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeServerFailure
	w.WriteMsg(m)
}
