package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
	"time"

	"github.com/fluentum-chain/fluentum/types"
)

// PluginManager manages the loading and lifecycle of Fluentum plugins
type PluginManager struct {
	plugins map[string]interface{}
	mutex   sync.RWMutex
	config  *PluginManagerConfig
}

// PluginManagerConfig contains configuration for the plugin manager
type PluginManagerConfig struct {
	PluginDirectory string            `json:"plugin_directory"`
	AutoLoad        bool              `json:"auto_load"`
	PluginConfigs   map[string]interface{} `json:"plugin_configs"`
	MaxPlugins      int               `json:"max_plugins"`
}

// DefaultPluginManagerConfig returns default configuration
func DefaultPluginManagerConfig() *PluginManagerConfig {
	return &PluginManagerConfig{
		PluginDirectory: "./plugins",
		AutoLoad:        true,
		MaxPlugins:      10,
		PluginConfigs:   make(map[string]interface{}),
	}
}

// Global plugin manager instance
var (
	pluginManager *PluginManager
	once          sync.Once
)

// Instance returns the global plugin manager instance
func Instance() *PluginManager {
	once.Do(func() {
		pluginManager = &PluginManager{
			plugins: make(map[string]interface{}),
			config:  DefaultPluginManagerConfig(),
		}
	})
	return pluginManager
}

// Initialize initializes the plugin manager with configuration
func (pm *PluginManager) Initialize(config *PluginManagerConfig) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.config = config

	// Create plugin directory if it doesn't exist
	if err := os.MkdirAll(pm.config.PluginDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Auto-load plugins if enabled
	if pm.config.AutoLoad {
		return pm.autoLoadPlugins()
	}

	return nil
}

// LoadPlugin loads a plugin from the specified path
func (pm *PluginManager) LoadPlugin(pluginPath, symbolName string) (interface{}, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Check if plugin is already loaded
	if plugin, exists := pm.plugins[pluginPath]; exists {
		return plugin, nil
	}

	// Check plugin count limit
	if len(pm.plugins) >= pm.config.MaxPlugins {
		return nil, fmt.Errorf("maximum number of plugins (%d) reached", pm.config.MaxPlugins)
	}

	// Load the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", pluginPath, err)
	}

	// Look up the symbol
	sym, err := p.Lookup(symbolName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup symbol %s in plugin %s: %w", symbolName, pluginPath, err)
	}

	// Store the plugin
	pm.plugins[pluginPath] = sym

	return sym, nil
}

// LoadAIPlugin loads the AI validation plugin
func (pm *PluginManager) LoadAIPlugin() (AIValidatorPlugin, error) {
	pluginPath := filepath.Join(pm.config.PluginDirectory, "qmoe_validator.so")
	
	plugin, err := pm.LoadPlugin(pluginPath, "AIValidatorPlugin")
	if err != nil {
		return nil, err
	}

	aiPlugin, ok := plugin.(AIValidatorPlugin)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement AIValidatorPlugin interface")
	}

	return aiPlugin, nil
}

// LoadSigner loads the quantum signing plugin
func (pm *PluginManager) LoadSigner() (SignerPlugin, error) {
	pluginPath := filepath.Join(pm.config.PluginDirectory, "quantum_signer.so")
	
	plugin, err := pm.LoadPlugin(pluginPath, "SignerPlugin")
	if err != nil {
		return nil, err
	}

	signer, ok := plugin.(SignerPlugin)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement SignerPlugin interface")
	}

	return signer, nil
}

// GetPlugin returns a loaded plugin by path
func (pm *PluginManager) GetPlugin(pluginPath string) (interface{}, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[pluginPath]
	return plugin, exists
}

// UnloadPlugin unloads a plugin
func (pm *PluginManager) UnloadPlugin(pluginPath string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.plugins[pluginPath]; !exists {
		return fmt.Errorf("plugin %s not loaded", pluginPath)
	}

	delete(pm.plugins, pluginPath)
	return nil
}

// ListPlugins returns a list of loaded plugins
func (pm *PluginManager) ListPlugins() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]string, 0, len(pm.plugins))
	for path := range pm.plugins {
		plugins = append(plugins, path)
	}

	return plugins
}

// GetPluginCount returns the number of loaded plugins
func (pm *PluginManager) GetPluginCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return len(pm.plugins)
}

// GetConfig returns the plugin manager configuration
func (pm *PluginManager) GetConfig() *PluginManagerConfig {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.config
}

// AutoLoadPlugins automatically loads plugins from the plugin directory
func (pm *PluginManager) autoLoadPlugins() error {
	entries, err := os.ReadDir(pm.config.PluginDirectory)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a plugin file
		if filepath.Ext(entry.Name()) == ".so" || filepath.Ext(entry.Name()) == ".dll" {
			pluginPath := filepath.Join(pm.config.PluginDirectory, entry.Name())
			
			// Try to load as AI plugin
			if _, err := pm.LoadAIPlugin(); err == nil {
				continue
			}

			// Try to load as signer plugin
			if _, err := pm.LoadSigner(); err == nil {
				continue
			}

			// Try generic plugin loading
			if _, err := pm.LoadPlugin(pluginPath, "Plugin"); err != nil {
				// Log warning but don't fail
				fmt.Printf("Warning: Failed to auto-load plugin %s: %v\n", pluginPath, err)
			}
		}
	}

	return nil
}

// ValidatePlugin validates a plugin before loading
func (pm *PluginManager) ValidatePlugin(pluginPath string) error {
	// Check if file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file does not exist: %s", pluginPath)
	}

	// Check file permissions
	info, err := os.Stat(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to stat plugin file: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("plugin file is not executable: %s", pluginPath)
	}

	return nil
}

// ReloadPlugin reloads a plugin
func (pm *PluginManager) ReloadPlugin(pluginPath string) error {
	// Unload first
	if err := pm.UnloadPlugin(pluginPath); err != nil {
		return err
	}

	// Load again
	_, err := pm.LoadPlugin(pluginPath, "Plugin")
	return err
}

// GetPluginInfo returns information about a loaded plugin
func (pm *PluginManager) GetPluginInfo(pluginPath string) (*PluginInfo, error) {
	plugin, exists := pm.GetPlugin(pluginPath)
	if !exists {
		return nil, fmt.Errorf("plugin not loaded: %s", pluginPath)
	}

	info := &PluginInfo{
		Path:      pluginPath,
		Type:      "unknown",
		Loaded:    true,
		LoadTime:  time.Now(), // We don't track actual load time, using current time
	}

	// Determine plugin type
	switch plugin.(type) {
	case AIValidatorPlugin:
		info.Type = "ai_validator"
	case SignerPlugin:
		info.Type = "signer"
	default:
		info.Type = "generic"
	}

	return info, nil
}

// PluginInfo contains information about a loaded plugin
type PluginInfo struct {
	Path     string    `json:"path"`
	Type     string    `json:"type"`
	Loaded   bool      `json:"loaded"`
	LoadTime time.Time `json:"load_time"`
}

// Shutdown gracefully shuts down the plugin manager
func (pm *PluginManager) Shutdown() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Unload all plugins
	for path := range pm.plugins {
		delete(pm.plugins, path)
	}

	return nil
}

// RegisterPlugin registers a plugin manually
func (pm *PluginManager) RegisterPlugin(name string, plugin interface{}) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if len(pm.plugins) >= pm.config.MaxPlugins {
		return fmt.Errorf("maximum number of plugins (%d) reached", pm.config.MaxPlugins)
	}

	pm.plugins[name] = plugin
	return nil
}

// GetAIPlugin returns the loaded AI plugin
func (pm *PluginManager) GetAIPlugin() (AIValidatorPlugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, plugin := range pm.plugins {
		if aiPlugin, ok := plugin.(AIValidatorPlugin); ok {
			return aiPlugin, nil
		}
	}

	return nil, fmt.Errorf("no AI validator plugin loaded")
}

// GetSigner returns the loaded signer plugin
func (pm *PluginManager) GetSigner() (SignerPlugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, plugin := range pm.plugins {
		if signer, ok := plugin.(SignerPlugin); ok {
			return signer, nil
		}
	}

	return nil, fmt.Errorf("no signer plugin loaded")
}

// LoadPluginWithConfig loads a plugin with configuration
func (pm *PluginManager) LoadPluginWithConfig(pluginPath, symbolName string, config interface{}) (interface{}, error) {
	plugin, err := pm.LoadPlugin(pluginPath, symbolName)
	if err != nil {
		return nil, err
	}

	// Initialize plugin with config if it supports it
	if initializer, ok := plugin.(interface{ Initialize(interface{}) error }); ok {
		if err := initializer.Initialize(config); err != nil {
			return nil, fmt.Errorf("failed to initialize plugin: %w", err)
		}
	}

	return plugin, nil
}

// LoadAIPluginWithConfig loads AI plugin with configuration
func (pm *PluginManager) LoadAIPluginWithConfig(config map[string]interface{}) (AIValidatorPlugin, error) {
	plugin, err := pm.LoadPluginWithConfig(
		filepath.Join(pm.config.PluginDirectory, "qmoe_validator.so"),
		"AIValidatorPlugin",
		config,
	)
	if err != nil {
		return nil, err
	}

	aiPlugin, ok := plugin.(AIValidatorPlugin)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement AIValidatorPlugin interface")
	}

	return aiPlugin, nil
}

// LoadSignerWithConfig loads signer plugin with configuration
func (pm *PluginManager) LoadSignerWithConfig(config interface{}) (SignerPlugin, error) {
	plugin, err := pm.LoadPluginWithConfig(
		filepath.Join(pm.config.PluginDirectory, "quantum_signer.so"),
		"SignerPlugin",
		config,
	)
	if err != nil {
		return nil, err
	}

	signer, ok := plugin.(SignerPlugin)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement SignerPlugin interface")
	}

	return signer, nil
}

// GetPluginStats returns statistics about loaded plugins
func (pm *PluginManager) GetPluginStats() *PluginStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := &PluginStats{
		TotalPlugins: len(pm.plugins),
		MaxPlugins:   pm.config.MaxPlugins,
		PluginTypes:  make(map[string]int),
	}

	for _, plugin := range pm.plugins {
		switch plugin.(type) {
		case AIValidatorPlugin:
			stats.PluginTypes["ai_validator"]++
		case SignerPlugin:
			stats.PluginTypes["signer"]++
		default:
			stats.PluginTypes["generic"]++
		}
	}

	return stats
}

// PluginStats contains statistics about loaded plugins
type PluginStats struct {
	TotalPlugins int            `json:"total_plugins"`
	MaxPlugins   int            `json:"max_plugins"`
	PluginTypes  map[string]int `json:"plugin_types"`
}

// Import time package for PluginInfo
import "time" 
