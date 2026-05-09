package plugin

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"sync"
	"sync/atomic"
	"time"
)

type loadedPlugin struct {
	name           string
	extension      Extension
	info           PluginMetadata
	activeRequests int32
	unloading      bool
	plugin         *plugin.Plugin
}

type Manager struct {
	plugins    map[string]*loadedPlugin
	mu         sync.RWMutex
	pluginDir  string
	stopWatching chan struct{}
	watching   bool
}

func NewManager(pluginDir string) *Manager {
	return &Manager{
		plugins:    make(map[string]*loadedPlugin),
		pluginDir:  pluginDir,
	}
}

func (m *Manager) PluginDir() string {
	return m.pluginDir
}

func (m *Manager) List() []PluginMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]PluginMetadata, 0, len(m.plugins))
	for name, p := range m.plugins {
		info := p.info
		info.Name = name
		infos = append(infos, info)
	}
	return infos
}

func (m *Manager) Execute(ctx context.Context, name string, input []byte) ([]byte, error) {
	m.mu.RLock()
	p, ok := m.plugins[name]
	if !ok {
		m.mu.RUnlock()
		return nil, ErrPluginNotFound
	}

	if p.unloading {
		m.mu.RUnlock()
		return nil, ErrPluginNotLoaded
	}

	atomic.AddInt32(&p.activeRequests, 1)
	m.mu.RUnlock()

	defer atomic.AddInt32(&p.activeRequests, -1)

	done := make(chan result, 1)
	go func() {
		var res result
		defer func() {
			if r := recover(); r != nil {
				res = result{err: fmt.Errorf("%w: %v", ErrPanic, r)}
				log.Printf("Plugin %s panicked: %v", name, r)
			}
			done <- res
		}()

		output, err := p.extension.Execute(ctx, input)
		res = result{output: output, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ErrTimeout
	case res := <-done:
		return res.output, res.err
	}
}

type result struct {
	output []byte
	err    error
}

func (m *Manager) Scan() error {
	if err := os.MkdirAll(m.pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(m.pluginDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to scan plugin directory: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	discovered := make(map[string]bool)
	for _, file := range files {
		name := filepath.Base(file)
		discovered[name] = true

		if _, ok := m.plugins[name]; ok {
			continue
		}

		if err := m.loadPluginUnlocked(file); err != nil {
			log.Printf("Failed to load plugin %s: %v", name, err)
		}
	}

	for name, p := range m.plugins {
		if !discovered[name] {
			p.unloading = true
			delete(m.plugins, name)
			go m.cleanupPlugin(p)
		}
	}

	return nil
}

func (m *Manager) loadPluginUnlocked(path string) error {
	name := filepath.Base(path)

	plug, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	infoSym, err := plug.Lookup("PluginInfo")
	if err != nil {
		return fmt.Errorf("plugin missing PluginInfo: %w", err)
	}

	info, ok := infoSym.(*PluginMetadata)
	if !ok {
		return fmt.Errorf("PluginInfo has wrong type")
	}

	if info.APIVersion != CurrentAPIVersion {
		return fmt.Errorf("%w: plugin requires %s, system is %s",
			ErrVersionMismatch, info.APIVersion, CurrentAPIVersion)
	}

	extSym, err := plug.Lookup("Extension")
	if err != nil {
		return fmt.Errorf("plugin missing Extension: %w", err)
	}

	ext, ok := extSym.(Extension)
	if !ok {
		return fmt.Errorf("Extension has wrong type")
	}

	m.plugins[name] = &loadedPlugin{
		name:      name,
		extension: ext,
		info:      *info,
		plugin:    plug,
	}

	log.Printf("Loaded plugin %s (API: %s)", name, info.APIVersion)
	return nil
}

func (m *Manager) cleanupPlugin(p *loadedPlugin) {
	for atomic.LoadInt32(&p.activeRequests) > 0 {
	}

	log.Printf("Unloaded plugin %s", p.name)
}

func (m *Manager) StartWatching() error {
	m.mu.Lock()
	if m.watching {
		m.mu.Unlock()
		return nil
	}
	m.watching = true
	m.stopWatching = make(chan struct{})
	m.mu.Unlock()

	if err := m.Scan(); err != nil {
		log.Printf("Initial scan failed: %v", err)
	}

	go m.watchLoop()
	return nil
}

func (m *Manager) StopWatching() {
	m.mu.Lock()
	if !m.watching {
		m.mu.Unlock()
		return
	}
	close(m.stopWatching)
	m.watching = false
	m.mu.Unlock()
}

func (m *Manager) watchLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopWatching:
			return
		case <-ticker.C:
			if err := m.Scan(); err != nil {
				log.Printf("Periodic scan failed: %v", err)
			}
		}
	}
}
