package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type SeverityLevel string

const (
	SeverityHigh   SeverityLevel = "high"
	SeverityMedium SeverityLevel = "medium"
	SeverityLow    SeverityLevel = "low"
)

type Rule struct {
	Name        string        `json:"name"`
	Pattern     string        `json:"pattern"`
	Severity    SeverityLevel `json:"severity"`
	Description string        `json:"description,omitempty"`
	Compiled    *regexp.Regexp
}

type Config struct {
	Rules        []Rule   `json:"rules"`
	ExcludePaths []string `json:"exclude_paths"`
	Extensions   struct {
		Binary    []string `json:"binary"`
		Text      []string `json:"text"`
	} `json:"extensions"`
}

var DefaultConfig = &Config{
	Rules: []Rule{
		{
			Name:        "password_assignment",
			Pattern:     `(?i)(password|passwd|pwd)\s*[=:]\s*['"]?([^\s'";,]+)['"]?`,
			Severity:    SeverityHigh,
			Description: "Password assignment pattern",
		},
		{
			Name:        "api_key_assignment",
			Pattern:     `(?i)(api[_-]?key|apikey)\s*[=:]\s*['"]?([^\s'";,]+)['"]?`,
			Severity:    SeverityHigh,
			Description: "API key assignment pattern",
		},
		{
			Name:        "secret_assignment",
			Pattern:     `(?i)(secret|token)\s*[=:]\s*['"]?([^\s'";,]+)['"]?`,
			Severity:    SeverityHigh,
			Description: "Secret or token assignment pattern",
		},
		{
			Name:        "database_connection_string",
			Pattern:     `(?i)(mysql|postgres|postgresql|mongodb|mssql|oracle)://([^:@]+):([^@]+)@`,
			Severity:    SeverityHigh,
			Description: "Database connection string with credentials",
		},
		{
			Name:        "mysql_connection",
			Pattern:     `(?i)([^:\s]+):([^@\s]+)@tcp\([^)]+\)`,
			Severity:    SeverityHigh,
			Description: "MySQL connection string format",
		},
		{
			Name:        "private_key_header",
			Pattern:     `-----BEGIN (RSA |EC |DSA |ED25519 )?PRIVATE KEY-----`,
			Severity:    SeverityCritical,
			Description: "Private key file header",
		},
		{
			Name:        "aws_access_key",
			Pattern:     `(?i)(aws_access_key_id|aws_secret_access_key)\s*[=:]\s*['"]?([^\s'";,]+)['"]?`,
			Severity:    SeverityHigh,
			Description: "AWS access key pattern",
		},
		{
			Name:        "github_token",
			Pattern:     `(?i)gh[ps]_[a-zA-Z0-9]{36,}`,
			Severity:    SeverityHigh,
			Description: "GitHub personal access token",
		},
		{
			Name:        "authorization_header",
			Pattern:     `(?i)(authorization|auth)\s*:\s*(bearer|basic|token)\s+([^\s'";,]+)`,
			Severity:    SeverityHigh,
			Description: "Authorization header with credentials",
		},
		{
			Name:        "credential_comment",
			Pattern:     `(?i)//\s*(password|api[_-]?key|secret|token)\s*[:=]\s*([^\s'";,]+)`,
			Severity:    SeverityMedium,
			Description: "Credentials in comments",
		},
	},
	ExcludePaths: []string{
		"vendor/",
		".git/",
		"node_modules/",
		".env",
		"*.log",
	},
}

const SeverityCritical SeverityLevel = "critical"

func (c *Config) CompileRules() error {
	for i := range c.Rules {
		re, err := regexp.Compile(c.Rules[i].Pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern for rule %s: %w", c.Rules[i].Name, err)
		}
		c.Rules[i].Compiled = re
	}
	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		if err := DefaultConfig.CompileRules(); err != nil {
			return nil, err
		}
		return DefaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.CompileRules(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) IsExcludedPath(path string, additionalExcludes []string) bool {
	allExcludes := append(c.ExcludePaths, additionalExcludes...)
	
	for _, pattern := range allExcludes {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
		
		if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
			dirPattern := pattern[:len(pattern)-1]
			parts := splitPath(path)
			for _, part := range parts {
				if part == dirPattern {
					return true
				}
			}
		}
	}
	return false
}

func splitPath(path string) []string {
	var parts []string
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" {
		parts = append(parts, filepath.Base(dir))
		dir = filepath.Dir(dir)
	}
	return parts
}

var binaryExtensions = map[string]bool{
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".bin":   true,
	".o":     true,
	".a":     true,
	".lib":   true,
	".pyc":   true,
	".pyo":   true,
	".pyd":   true,
	".class": true,
	".jar":   true,
	".war":   true,
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".bmp":   true,
	".tiff":  true,
	".ico":   true,
	".pdf":   true,
	".doc":   true,
	".docx":  true,
	".xls":   true,
	".xlsx":  true,
	".ppt":   true,
	".pptx":  true,
	".zip":   true,
	".rar":   true,
	".7z":    true,
	".tar":   true,
	".gz":    true,
	".bz2":   true,
	".xz":    true,
}

func IsBinaryFile(path string) bool {
	ext := filepath.Ext(path)
	return binaryExtensions[ext]
}

func (c *Config) GetRulesBySeverity(severity SeverityLevel) []Rule {
	var rules []Rule
	for _, rule := range c.Rules {
		if matchSeverity(rule.Severity, severity) {
			rules = append(rules, rule)
		}
	}
	return rules
}

func matchSeverity(ruleSeverity, filterSeverity SeverityLevel) bool {
	if filterSeverity == "" {
		return true
	}
	
	severityOrder := map[SeverityLevel]int{
		SeverityLow:      1,
		SeverityMedium:   2,
		SeverityHigh:     3,
		SeverityCritical: 4,
	}
	
	ruleLevel := severityOrder[ruleSeverity]
	filterLevel := severityOrder[filterSeverity]
	
	return ruleLevel >= filterLevel
}
