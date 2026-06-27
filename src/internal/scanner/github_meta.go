package scanner

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// GitHubMeta GitHub官方公开的IP地址元数据
type GitHubMeta struct {
	VerifiablePasswordAuthentication bool     `json:"verifiable_password_authentication"`
	SSHKeyFingerprints               struct{} `json:"ssh_key_fingerprints"`
	SSHKeys                          []string `json:"ssh_keys"`
	Hooks                            []string `json:"hooks"`
	Web                              []string `json:"web"`
	API                              []string `json:"api"`
	Git                              []string `json:"git"`
	Packages                         []string `json:"packages"`
	Pages                            []string `json:"pages"`
	Importer                         []string `json:"importer"`
	Actions                          []string `json:"actions"`
	Dependabot                       []string `json:"dependabot"`
}

// FetchGitHubIPRanges 从GitHub官方API获取IP地址范围
func FetchGitHubIPRanges(timeout time.Duration) (*GitHubMeta, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get("https://api.github.com/meta")
	if err != nil {
		return nil, fmt.Errorf("fetch github meta: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github meta API returned status: %d", resp.StatusCode)
	}

	var meta GitHubMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decode github meta: %w", err)
	}

	return &meta, nil
}

// ExpandCIDR 将CIDR展开为单个IP列表（限制数量避免过大）
func ExpandCIDR(cidr string, maxIPs int) ([]net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ips = append(ips, dupIP(ip))
		if len(ips) >= maxIPs {
			break
		}
	}

	return ips, nil
}

// GetAllIPsFromMeta 从GitHubMeta提取所有IP（包括CIDR展开）
func GetAllIPsFromMeta(meta *GitHubMeta, maxIPsPerCIDR int) []net.IP {
	ipSet := make(map[string]net.IP)

	allCIDRs := append([]string{}, meta.Hooks...)
	allCIDRs = append(allCIDRs, meta.Web...)
	allCIDRs = append(allCIDRs, meta.API...)
	allCIDRs = append(allCIDRs, meta.Git...)
	allCIDRs = append(allCIDRs, meta.Packages...)
	allCIDRs = append(allCIDRs, meta.Pages...)
	allCIDRs = append(allCIDRs, meta.Actions...)
	allCIDRs = append(allCIDRs, meta.Dependabot...)

	for _, cidr := range allCIDRs {
		ips, err := ExpandCIDR(cidr, maxIPsPerCIDR)
		if err != nil {
			continue
		}
		for _, ip := range ips {
			ipSet[ip.String()] = ip
		}
	}

	result := make([]net.IP, 0, len(ipSet))
	for _, ip := range ipSet {
		result = append(result, ip)
	}
	return result
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}
