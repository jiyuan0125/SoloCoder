package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"dataconverter/config"
)

type RuleLoader struct {
	engine        *RuleEngine
	rulesDir      string
	rulesFile     string
	watchInterval time.Duration
	watcherMu     sync.Mutex
	watching      bool
	stopChan      chan struct{}
	lastModTime   time.Time
}

func NewRuleLoader(engine *RuleEngine) *RuleLoader {
	return &RuleLoader{
		engine:        engine,
		watchInterval: 5 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

func (l *RuleLoader) SetRulesDir(dir string) {
	l.rulesDir = dir
}

func (l *RuleLoader) SetRulesFile(file string) {
	l.rulesFile = file
}

func (l *RuleLoader) LoadFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}
	
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		var rules []config.Rule
		if err2 := json.Unmarshal(data, &rules); err2 != nil {
			return fmt.Errorf("failed to parse rules: %w (tried both Config and []Rule)", err)
		}
		l.engine.LoadRules(rules)
		return nil
	}
	
	l.engine.LoadRules(cfg.Rules)
	return nil
}

func (l *RuleLoader) LoadFromDir(dirPath string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list rules files: %w", err)
	}
	
	var allRules []config.Rule
	
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}
		
		var cfg config.Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			var rules []config.Rule
			if err2 := json.Unmarshal(data, &rules); err2 != nil {
				return fmt.Errorf("failed to parse %s: %w", file, err)
			}
			allRules = append(allRules, rules...)
			continue
		}
		allRules = append(allRules, cfg.Rules...)
	}
	
	l.engine.LoadRules(allRules)
	return nil
}

func (l *RuleLoader) LoadFromJSON(data []byte) error {
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		var rules []config.Rule
		if err2 := json.Unmarshal(data, &rules); err2 != nil {
			return fmt.Errorf("failed to parse rules: %w (tried both Config and []Rule)", err)
		}
		l.engine.LoadRules(rules)
		return nil
	}
	l.engine.LoadRules(cfg.Rules)
	return nil
}

func (l *RuleLoader) LoadRules(rules []config.Rule) {
	l.engine.LoadRules(rules)
}

func (l *RuleLoader) StartWatching() error {
	l.watcherMu.Lock()
	defer l.watcherMu.Unlock()
	
	if l.watching {
		return fmt.Errorf("already watching for rule changes")
	}
	
	if l.rulesFile == "" && l.rulesDir == "" {
		return fmt.Errorf("no rules file or directory configured")
	}
	
	l.watching = true
	l.stopChan = make(chan struct{})
	
	go l.watchLoop()
	
	return nil
}

func (l *RuleLoader) StopWatching() {
	l.watcherMu.Lock()
	defer l.watcherMu.Unlock()
	
	if !l.watching {
		return
	}
	
	close(l.stopChan)
	l.watching = false
}

func (l *RuleLoader) watchLoop() {
	ticker := time.NewTicker(l.watchInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-l.stopChan:
			return
		case <-ticker.C:
			l.checkAndReload()
		}
	}
}

func (l *RuleLoader) checkAndReload() {
	var modTime time.Time
	var target string
	
	if l.rulesFile != "" {
		info, err := os.Stat(l.rulesFile)
		if err != nil {
			return
		}
		modTime = info.ModTime()
		target = l.rulesFile
	} else if l.rulesDir != "" {
		info, err := os.Stat(l.rulesDir)
		if err != nil {
			return
		}
		modTime = info.ModTime()
		target = l.rulesDir
	}
	
	if modTime.After(l.lastModTime) {
		l.lastModTime = modTime
		
		if target == l.rulesFile {
			l.LoadFromFile(target)
		} else {
			l.LoadFromDir(target)
		}
	}
}

func (l *RuleLoader) GetEngine() *RuleEngine {
	return l.engine
}
