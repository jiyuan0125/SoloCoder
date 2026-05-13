package whois

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type WHOISRecord struct {
	Domain          string   `json:"domain"`
	Registrar       string   `json:"registrar"`
	CreationDate    string   `json:"creation_date"`
	ExpirationDate  string   `json:"expiration_date"`
	UpdatedDate     string   `json:"updated_date"`
	NameServers     []string `json:"name_servers"`
	Status          []string `json:"status"`
	RawResult       string   `json:"raw_result,omitempty"`
}

type WHOISQueryResult struct {
	Domain         string    `json:"domain"`
	Record         *WHOISRecord `json:"record"`
	QueryTime      time.Duration `json:"query_time_ms"`
	QueriedAt      time.Time `json:"queried_at"`
}

type WHOISService struct {
	db *sql.DB
}

func NewWHOISService(dbPath string) (*WHOISService, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	service := &WHOISService{db: db}
	if err := service.init(); err != nil {
		return nil, err
	}

	return service, nil
}

func (w *WHOISService) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS whois_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL UNIQUE,
		registrar TEXT,
		creation_date TEXT,
		expiration_date TEXT,
		updated_date TEXT,
		name_servers TEXT,
		status TEXT,
		raw_result TEXT,
		queried_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := w.db.Exec(schema)
	return err
}

func (w *WHOISService) Query(ctx context.Context, domain string) (*WHOISQueryResult, error) {
	domain = strings.TrimSpace(strings.ToLower(domain))
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	start := time.Now()

	whoisServer, err := getWHOISServer(domain)
	if err != nil {
		return nil, err
	}

	rawResult, err := queryWHOISRaw(ctx, domain, whoisServer)
	if err != nil {
		return nil, err
	}

	record := parseWHOISResponse(domain, rawResult)

	queryTime := time.Since(start)

	if err := w.saveRecord(domain, record, rawResult); err != nil {
		return nil, err
	}

	return &WHOISQueryResult{
		Domain:    domain,
		Record:    record,
		QueryTime: queryTime,
		QueriedAt: time.Now(),
	}, nil
}

func getWHOISServer(domain string) (string, error) {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain")
	}

	tld := parts[len(parts)-1]
	
	tldServers := map[string]string{
		"com": "whois.verisign-grs.com",
		"net": "whois.verisign-grs.com",
		"org": "whois.pir.org",
		"edu": "whois.educause.edu",
		"gov": "whois.nic.gov",
		"mil": "whois.nic.mil",
		"info": "whois.afilias.info",
		"biz": "whois.neulevel.biz",
		"io": "whois.nic.io",
		"co": "whois.nic.co",
		"me": "whois.nic.me",
		"tv": "whois.nic.tv",
		"cc": "ccwhois.verisign-grs.com",
		"cn": "whois.cnnic.cn",
		"jp": "whois.jprs.jp",
		"de": "whois.denic.de",
		"fr": "whois.nic.fr",
		"uk": "whois.nic.uk",
		"au": "whois.auda.org.au",
		"ca": "whois.cira.ca",
		"in": "whois.registry.in",
		"br": "whois.registro.br",
	}

	if server, exists := tldServers[tld]; exists {
		return server + ":43", nil
	}

	return "whois.iana.org:43", nil
}

func queryWHOISRaw(ctx context.Context, domain, server string) (string, error) {
	var d net.Dialer
	d.Timeout = 30 * time.Second

	conn, err := d.DialContext(ctx, "tcp", server)
	if err != nil {
		return "", fmt.Errorf("failed to connect to WHOIS server: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	_, err = fmt.Fprintf(conn, "%s\r\n", domain)
	if err != nil {
		return "", fmt.Errorf("failed to send WHOIS query: %w", err)
	}

	var response strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		response.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading WHOIS response: %w", err)
	}

	return response.String(), nil
}

func parseWHOISResponse(domain, raw string) *WHOISRecord {
	record := &WHOISRecord{
		Domain:      domain,
		RawResult:   raw,
		NameServers: []string{},
		Status:      []string{},
	}

	lines := strings.Split(raw, "\n")
	
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, ">") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "registrar", "sponsoring registrar", "registrar name":
			if record.Registrar == "" {
				record.Registrar = value
			}
		case "creation date", "created", "registration date", "domain registration date":
			if record.CreationDate == "" {
				record.CreationDate = value
			}
		case "expiration date", "expires", "expiry date", "registry expiry date":
			if record.ExpirationDate == "" {
				record.ExpirationDate = value
			}
		case "updated date", "last updated", "modified", "last modified":
			if record.UpdatedDate == "" {
				record.UpdatedDate = value
			}
		case "name server", "nameserver", "nserver", "dns", "domain nameservers":
			ns := strings.ToLower(value)
			if !seen[ns] {
				seen[ns] = true
				record.NameServers = append(record.NameServers, ns)
			}
		case "status", "domain status":
			if !seen[value] {
				seen[value] = true
				record.Status = append(record.Status, value)
			}
		}
	}

	return record
}

func (w *WHOISService) saveRecord(domain string, record *WHOISRecord, rawResult string) error {
	nsJSON, _ := json.Marshal(record.NameServers)
	statusJSON, _ := json.Marshal(record.Status)

	query := `
	INSERT INTO whois_records (domain, registrar, creation_date, expiration_date, updated_date, name_servers, status, raw_result, queried_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(domain) DO UPDATE SET
		registrar = excluded.registrar,
		creation_date = excluded.creation_date,
		expiration_date = excluded.expiration_date,
		updated_date = excluded.updated_date,
		name_servers = excluded.name_servers,
		status = excluded.status,
		raw_result = excluded.raw_result,
		queried_at = excluded.queried_at
	`

	_, err := w.db.Exec(
		query,
		domain,
		record.Registrar,
		record.CreationDate,
		record.ExpirationDate,
		record.UpdatedDate,
		string(nsJSON),
		string(statusJSON),
		rawResult,
		time.Now(),
	)

	return err
}

func (w *WHOISService) Close() error {
	return w.db.Close()
}
