package parser

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type Config map[string]string

type Parser interface {
    ParseFile(filePath string) (Config, error)
    ParseString(content string) (Config, error)
}

func NewParser(filePath string) Parser {
    ext := strings.ToLower(filepath.Ext(filePath))
    if ext == ".yaml" || ext == ".yml" {
        return &YAMLParser{}
    }
    return &KeyValueParser{}
}

func ParseFile(filePath string) (Config, error) {
    parser := NewParser(filePath)
    return parser.ParseFile(filePath)
}

func normalizeValue(value string) string {
    value = strings.TrimSpace(value)
    if len(value) >= 2 {
        first := value[0]
        last := value[len(value)-1]
        if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
            value = value[1 : len(value)-1]
        }
    }
    lower := strings.ToLower(value)
    if lower == "true" || lower == "false" || lower == "null" {
        return lower
    }
    return value
}

func isCommentOrEmpty(line string) bool {
    trimmed := strings.TrimSpace(line)
    return trimmed == "" || strings.HasPrefix(trimmed, "#")
}

type KeyValueParser struct{}

func (p *KeyValueParser) ParseFile(filePath string) (Config, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    config := make(Config)
    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()
        
        if isCommentOrEmpty(line) {
            continue
        }
        
        key, value, err := p.parseLine(line)
        if err != nil {
            return nil, fmt.Errorf("line %d: %w", lineNum, err)
        }
        
        config[key] = value
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return config, nil
}

func (p *KeyValueParser) ParseString(content string) (Config, error) {
    config := make(Config)
    scanner := bufio.NewScanner(strings.NewReader(content))
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()
        
        if isCommentOrEmpty(line) {
            continue
        }
        
        key, value, err := p.parseLine(line)
        if err != nil {
            return nil, fmt.Errorf("line %d: %w", lineNum, err)
        }
        
        config[key] = value
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return config, nil
}

func (p *KeyValueParser) parseLine(line string) (string, string, error) {
    line = strings.TrimSpace(line)
    equalIndex := strings.Index(line, "=")
    if equalIndex == -1 {
        return "", "", fmt.Errorf("invalid line format: %s", line)
    }
    
    key := strings.TrimSpace(line[:equalIndex])
    if key == "" {
        return "", "", fmt.Errorf("empty key in line: %s", line)
    }
    
    value := strings.TrimSpace(line[equalIndex+1:])
    value = normalizeValue(value)
    
    return key, value, nil
}

type YAMLParser struct{}

func (p *YAMLParser) ParseFile(filePath string) (Config, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    config := make(Config)
    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()
        
        if isCommentOrEmpty(line) {
            continue
        }
        
        key, value, err := p.parseLine(line)
        if err != nil {
            return nil, fmt.Errorf("line %d: %w", lineNum, err)
        }
        
        if key != "" {
            config[key] = value
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return config, nil
}

func (p *YAMLParser) ParseString(content string) (Config, error) {
    config := make(Config)
    scanner := bufio.NewScanner(strings.NewReader(content))
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()
        
        if isCommentOrEmpty(line) {
            continue
        }
        
        key, value, err := p.parseLine(line)
        if err != nil {
            return nil, fmt.Errorf("line %d: %w", lineNum, err)
        }
        
        if key != "" {
            config[key] = value
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return config, nil
}

func (p *YAMLParser) parseLine(line string) (string, string, error) {
    line = strings.TrimSpace(line)
    
    if strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ") {
        return "", "", nil
    }
    
    colonIndex := strings.Index(line, ":")
    if colonIndex == -1 {
        return "", "", fmt.Errorf("invalid YAML line format: %s", line)
    }
    
    key := strings.TrimSpace(line[:colonIndex])
    if key == "" {
        return "", "", fmt.Errorf("empty key in line: %s", line)
    }
    
    value := strings.TrimSpace(line[colonIndex+1:])
    value = normalizeValue(value)
    
    return key, value, nil
}
