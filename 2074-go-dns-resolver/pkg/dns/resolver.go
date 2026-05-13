package dns

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const (
	DefaultDNSServer = "8.8.8.8:53"
	DefaultTimeout   = 5 * time.Second
)

type RecordResult struct {
	RecordType string        `json:"type"`
	RecordValue string       `json:"value"`
	TTL          uint32       `json:"ttl"`
}

type DNSQueryResult struct {
	Domain      string         `json:"domain"`
	RecordType  string         `json:"record_type"`
	Records     []RecordResult `json:"records"`
	QueryTime   time.Duration  `json:"query_time_ms"`
	FromCache   bool           `json:"from_cache"`
	DNSServer   string         `json:"dns_server"`
	Timestamp   time.Time      `json:"timestamp"`
}

type Resolver struct {
	cache   *Cache
	mu      sync.Mutex
}

func NewResolver() *Resolver {
	return &Resolver{
		cache: NewCache(),
	}
}

func (r *Resolver) Query(ctx context.Context, domain, recordType, dnsServer string) (*DNSQueryResult, error) {
	if !IsValidDomain(domain) {
		return nil, &ValidationError{Message: "invalid domain format"}
	}

	qType, err := getQueryType(recordType)
	if err != nil {
		return nil, err
	}

	if dnsServer == "" {
		dnsServer = DefaultDNSServer
	}

	cacheKey := buildCacheKey(domain, recordType, dnsServer)
	if cached, found := r.cache.Get(cacheKey); found {
		cached.FromCache = true
		return cached, nil
	}

	start := time.Now()
	client := &dns.Client{
		Net:     "udp",
		Timeout: DefaultTimeout,
	}

	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(domain), qType)

	resp, _, err := client.ExchangeContext(ctx, msg, dnsServer)
	queryTime := time.Since(start)

	if err != nil {
		return nil, &ServerError{Message: err.Error()}
	}

	if resp == nil {
		return nil, &ServerError{Message: "no response from DNS server"}
	}

	records := parseDNSResponse(resp, recordType)

	result := &DNSQueryResult{
		Domain:     domain,
		RecordType: recordType,
		Records:    records,
		QueryTime:  queryTime,
		FromCache:  false,
		DNSServer:  dnsServer,
		Timestamp:  time.Now(),
	}

	if len(records) > 0 {
		r.cache.Set(cacheKey, result)
	}

	return result, nil
}

func (r *Resolver) ReverseQuery(ctx context.Context, ipStr, dnsServer string) (*DNSQueryResult, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, &ValidationError{Message: "invalid IP address format"}
	}

	if dnsServer == "" {
		dnsServer = DefaultDNSServer
	}

	reverseAddr, err := dns.ReverseAddr(ipStr)
	if err != nil {
		return nil, &ValidationError{Message: "invalid IP address for reverse lookup"}
	}

	start := time.Now()
	client := &dns.Client{
		Net:     "udp",
		Timeout: DefaultTimeout,
	}

	msg := &dns.Msg{}
	msg.SetQuestion(reverseAddr, dns.TypePTR)

	resp, _, err := client.ExchangeContext(ctx, msg, dnsServer)
	queryTime := time.Since(start)

	if err != nil {
		return nil, &ServerError{Message: err.Error()}
	}

	if resp == nil {
		return nil, &ServerError{Message: "no response from DNS server"}
	}

	records := parseDNSResponse(resp, "PTR")

	return &DNSQueryResult{
		Domain:     ipStr,
		RecordType: "PTR",
		Records:    records,
		QueryTime:  queryTime,
		FromCache:  false,
		DNSServer:  dnsServer,
		Timestamp:  time.Now(),
	}, nil
}

func getQueryType(recordType string) (uint16, error) {
	recordType = strings.ToUpper(recordType)
	switch recordType {
	case "A":
		return dns.TypeA, nil
	case "AAAA":
		return dns.TypeAAAA, nil
	case "CNAME":
		return dns.TypeCNAME, nil
	case "MX":
		return dns.TypeMX, nil
	case "TXT":
		return dns.TypeTXT, nil
	case "NS":
		return dns.TypeNS, nil
	case "SRV":
		return dns.TypeSRV, nil
	default:
		return 0, &RecordTypeError{
			Supported: []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SRV"},
		}
	}
}

func parseDNSResponse(resp *dns.Msg, recordType string) []RecordResult {
	var records []RecordResult
	recordType = strings.ToUpper(recordType)

	for _, rr := range resp.Answer {
		switch v := rr.(type) {
		case *dns.A:
			if recordType == "A" {
				records = append(records, RecordResult{
					RecordType:  "A",
					RecordValue: v.A.String(),
					TTL:         rr.Header().Ttl,
				})
			}
		case *dns.AAAA:
			if recordType == "AAAA" {
				records = append(records, RecordResult{
					RecordType:  "AAAA",
					RecordValue: v.AAAA.String(),
					TTL:         rr.Header().Ttl,
				})
			}
		case *dns.CNAME:
			if recordType == "CNAME" {
				records = append(records, RecordResult{
					RecordType:  "CNAME",
					RecordValue: v.Target,
					TTL:         rr.Header().Ttl,
				})
			}
		case *dns.MX:
			if recordType == "MX" {
				records = append(records, RecordResult{
					RecordType:  "MX",
					RecordValue: v.Mx,
					TTL:         rr.Header().Ttl,
				})
			}
		case *dns.TXT:
			if recordType == "TXT" {
				for _, txt := range v.Txt {
					records = append(records, RecordResult{
						RecordType:  "TXT",
						RecordValue: txt,
						TTL:         rr.Header().Ttl,
					})
				}
			}
		case *dns.NS:
			if recordType == "NS" {
				records = append(records, RecordResult{
					RecordType:  "NS",
					RecordValue: v.Ns,
					TTL:         rr.Header().Ttl,
				})
			}
		case *dns.SRV:
			if recordType == "SRV" {
				records = append(records, RecordResult{
					RecordType:  "SRV",
					RecordValue: v.Target,
					TTL:         rr.Header().Ttl,
				})
			}
		case *dns.PTR:
			if recordType == "PTR" {
				records = append(records, RecordResult{
					RecordType:  "PTR",
					RecordValue: v.Ptr,
					TTL:         rr.Header().Ttl,
				})
			}
		}
	}

	return records
}

func buildCacheKey(domain, recordType, dnsServer string) string {
	return strings.ToLower(domain) + "|" + strings.ToUpper(recordType) + "|" + dnsServer
}

func IsValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	domain = strings.TrimSuffix(domain, ".")
	
	for _, label := range strings.Split(domain, ".") {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, r := range label {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
				return false
			}
		}
	}

	return true
}
