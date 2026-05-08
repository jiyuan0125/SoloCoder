package dnscache

import (
	"dns-cache-service/common"
	"strings"
)

type mockRecord struct {
	A     []string
	AAAA  []string
	CNAME string
	MX    []string
	TXT   []string
	TTL   int
}

var mockData = map[string]*mockRecord{
	"google.com": {
		A:     []string{"142.250.74.46", "142.250.74.14"},
		AAAA:  []string{"2404:6800:4005:815::200e"},
		CNAME: "",
		MX:    []string{"smtp.google.com"},
		TXT:   []string{"v=spf1 include:_spf.google.com ~all"},
		TTL:   300,
	},
	"www.google.com": {
		A:     []string{"142.250.74.36"},
		AAAA:  []string{"2404:6800:4005:813::2004"},
		CNAME: "google.com",
		TTL:   300,
	},
	"baidu.com": {
		A:     []string{"110.242.68.66", "110.242.68.3"},
		AAAA:  []string{"2400:da00:2::29"},
		MX:    []string{"mx.baidu.com"},
		TTL:   600,
	},
	"github.com": {
		A:     []string{"20.205.243.166"},
		AAAA:  []string{},
		MX:    []string{"alt1.aspmx.l.google.com"},
		TTL:   30,
	},
	"example.com": {
		A:     []string{"93.184.216.34"},
		AAAA:  []string{"2606:2800:220:1:248:1893:25c8:1946"},
		TTL:   86400,
	},
	"test.example.com": {
		A:     []string{"192.168.1.100"},
		TTL:   0,
	},
}

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve(domain string, recordType common.RecordType) ([]common.DNSRecord, bool, error) {
	domain = strings.ToLower(domain)
	record, exists := mockData[domain]

	if !exists {
		return nil, true, nil
	}

	var records []common.DNSRecord
	switch recordType {
	case common.TypeA:
		for _, ip := range record.A {
			records = append(records, common.DNSRecord{
				Type:  common.TypeA,
				Value: ip,
				TTL:   record.TTL,
			})
		}
	case common.TypeAAAA:
		for _, ip := range record.AAAA {
			records = append(records, common.DNSRecord{
				Type:  common.TypeAAAA,
				Value: ip,
				TTL:   record.TTL,
			})
		}
	case common.TypeCNAME:
		if record.CNAME != "" {
			records = append(records, common.DNSRecord{
				Type:  common.TypeCNAME,
				Value: record.CNAME,
				TTL:   record.TTL,
			})
		}
	case common.TypeMX:
		for _, mx := range record.MX {
			records = append(records, common.DNSRecord{
				Type:  common.TypeMX,
				Value: mx,
				TTL:   record.TTL,
			})
		}
	case common.TypeTXT:
		for _, txt := range record.TXT {
			records = append(records, common.DNSRecord{
				Type:  common.TypeTXT,
				Value: txt,
				TTL:   record.TTL,
			})
		}
	}

	if len(records) == 0 {
		return nil, false, nil
	}

	return records, false, nil
}
