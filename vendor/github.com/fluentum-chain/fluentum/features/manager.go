package features

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Logger is a simple logger interface to replace the missing logging functionality
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	With(keyvals ...interface{}) Logger
}

// simpleLogger is a basic implementation of the Logger interface
type simpleLogger struct {
	prefix string
}

func (l *simpleLogger) log(level, msg string, keyvals ...interface{}) {
	var builder strings.Builder
	builder.WriteString(level)
	builder.WriteString("\t")
	builder.WriteString(l.prefix)
	builder.WriteString(msg)
	
	// Add key-value pairs if any
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			builder.WriteString(fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1]))
		} else {
			builder.WriteString(fmt.Sprintf(" %v", keyvals[i]))
		}
	}
	
	log.Println(builder.String())
}

func (l *simpleLogger) Debug(msg string, keyvals ...interface{}) {
	l.log("DEBUG", msg, keyvals...)
}

func (l *simpleLogger) Info(msg string, keyvals ...interface{}) {
	l.log("INFO", msg, keyvals...)
}

func (l *simpleLogger) Error(msg string, keyvals ...interface{}) {
	l.log("ERROR", msg, keyvals...)
}

func (l *simpleLogger) With(keyvals ...interface{}) Logger {
	// For simplicity, just append to the prefix
	prefix := l.prefix
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			prefix += fmt.Sprintf("%v=%v ", keyvals[i], keyvals[i+1])
		} else {
			prefix += fmt.Sprintf("%v ", keyvals[i])
		}
	}
	return &simpleLogger{prefix: prefix}
}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger() Logger {
	return &simpleLogger{}
}

// FeatureInfo represents a loadable feature
type FeatureInfo struct {
	Name        string
	Version     string
	Description string
	Enabled     bool
	Plugin      *plugin.Plugin
	Config      interface{}
}

// Config defines the interface for feature configuration
type Config interface {
	GetRegistry() string
	IsEnabled() bool
	GetAutoUpdate() bool
}

// defaultConfig provides default values for the feature manager
type defaultConfig struct{}

func (c *defaultConfig) GetRegistry() string {
	return "features"
}

func (c *defaultConfig) IsEnabled() bool {
	return true
}

func (c *defaultConfig) GetAutoUpdate() bool {
	return false
}

// Manager handles feature loading and management
type Manager struct {
	logger  Logger
	config  Config
	mu      sync.RWMutex
	features map[string]*FeatureInfo
}

// NewManager creates a new feature manager
func NewManager(cfg interface{}, logger Logger) *Manager {
	var config Config
	if cfg == nil {
		config = &defaultConfig{}
	} else if c, ok := cfg.(Config); ok {
		config = c
	} else {
		// Fallback to default config if the provided config doesn't implement the interface
		config = &defaultConfig{}
	}
	
	if logger == nil {
		logger = NewSimpleLogger()
	}
	
	return &Manager{
		logger:  logger.With("module", "features"),
		config:  config,
		features: make(map[string]*FeatureInfo),
	}
}

// LoadFeatures loads all enabled features
func (m *Manager) LoadFeatures(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the registry path from config
	registryPath := m.config.GetRegistry()
	if registryPath == "" {
		registryPath = "features"
	}

	// Ensure feature directory exists
	if err := os.MkdirAll(registryPath, 0755); err != nil {
		return fmt.Errorf("failed to create features directory: %w", err)
	}

	// If the feature is enabled in config, load it
	if m.config.IsEnabled() {
		// List all feature directories
		dirs, err := os.ReadDir(registryPath)
		if err != nil {
			return fmt.Errorf("failed to read features directory: %w", err)
		}

		// Load each feature
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				name := dir.Name()
				if err := m.loadFeature(name); err != nil {
					m.logger.Error("Failed to load feature", "name", name, "error", err)
					continue
				}
			}
		}
	}

	return nil
}

// GetFeature returns a loaded feature by name
func (m *Manager) GetFeature(name string) (*FeatureInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[name]
	return feature, exists
}

// ReloadFeature reloads a specific feature
func (m *Manager) ReloadFeature(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Unload existing feature if loaded
	if _, exists := m.features[name]; exists {
		// Note: Go's plugin system doesn't support unloading, so we just remove the reference
		delete(m.features, name)
	}

	return m.loadFeature(name)
}

// loadFeature loads a single feature
func (m *Manager) loadFeature(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if feature is already loaded
	if _, exists := m.features[name]; exists {
		return fmt.Errorf("feature %s is already loaded", name)
	}

	// Create a new feature
	feature := &FeatureInfo{
		Name: name,
	}

	// Load the plugin
	pluginPath := m.getPluginPath(name)
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %v", pluginPath, err)
	}

	// Verify the feature implements the required interface
	if _, err := plug.Lookup("Feature"); err != nil {
		return fmt.Errorf("feature %s does not export \"Feature\" symbol: %v", name, err)
	}

	// Store the plugin
	feature.Plugin = plug

	// Load the feature configuration
	if err := m.loadFeatureConfig(name, feature); err != nil {
		return fmt.Errorf("failed to load config for feature %s: %v", name, err)
	}

	// Add to the features map
	m.features[name] = feature

	m.logger.Info("Loaded feature", "name", name, "version", feature.Version)
	return nil
}

// loadFeatureConfig loads configuration for a feature
func (m *Manager) loadFeatureConfig(name string, feature *FeatureInfo) error {
	// Default config path: $HOME/.fluentum/features/{name}/config.json
	configPath := filepath.Join(m.config.GetRegistry(), name, "config.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file, use defaults
		m.logger.Debug("No config file found, using defaults", "feature", name, "path", configPath)
		return nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", configPath, err)
	}

	// Unmarshal the config
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %v", configPath, err)
	}

	// Set the config
	feature.Config = config

	// Set version if available
	if version, ok := config["version"].(string); ok {
		feature.Version = version
	}

	// Set description if available
	if description, ok := config["description"].(string); ok {
		feature.Description = description
	}

	// Set enabled if available
	if enabled, ok := config["enabled"].(bool); ok {
		feature.Enabled = enabled
	}

	m.logger.Debug("Loaded config for feature", "feature", name, "path", configPath)
	return nil
}

// getPluginPath returns the path to a feature's plugin file
func (m *Manager) getPluginPath(name string) string {
	// On Windows, we need to add .dll extension
	ext := ".so"
	if runtime.GOOS == "windows" {
		ext = ".dll"
	}
	// Ensure the directory exists
	if err := os.MkdirAll(m.config.GetRegistry(), 0755); err != nil {
		m.logger.Error("Failed to create features directory", "error", err)
	}
	return filepath.Join(m.config.GetRegistry(), "lib"+name+ext)
}

// CheckForUpdates checks for updates to installed features
func (m *Manager) CheckForUpdates() (map[string]string, error) {
	updates := make(map[string]string)

	// In a real implementation, this would check a remote registry
	// For now, we'll just return an empty map

	return updates, nil
}

// InstallFeature installs a new feature
func (m *Manager) InstallFeature(name, version string) error {
	// In a real implementation, this would download the feature
	// from a registry and install it
	return fmt.Errorf("not implemented")
}

// UninstallFeature removes an installed feature
func (m *Manager) UninstallFeature(name string) error {
	featurePath := filepath.Join(m.config.GetRegistry(), name)
	return os.RemoveAll(featurePath)
}

// ListFeatures returns a list of all installed features
func (m *Manager) ListFeatures() ([]*FeatureInfo, error) {
	var features []*FeatureInfo

	err := filepath.WalkDir(m.config.GetRegistry(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != m.config.GetRegistry() {
			name := filepath.Base(path)
			if feature, exists := m.GetFeature(name); exists {
				features = append(features, feature)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return features, nil
}

// StartUpdateChecker starts a background goroutine to check for updates
func (m *Manager) StartUpdateChecker(interval time.Duration) {
	if !m.config.GetAutoUpdate() {
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			updates, err := m.CheckForUpdates()
			if err != nil {
				m.logger.Error("Failed to check for updates", "error", err)
				continue
			}

			if len(updates) > 0 {
				m.logger.Info("Updates available", "updates", updates)
			}
		}
	}()
}

// UpdateFeature updates a feature to the latest version
func (m *Manager) UpdateFeature(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	feature, exists := m.features[name]
	if !exists {
		return fmt.Errorf("feature not found: %s", name)
	}

	// In a real implementation, this would download and install the latest version
	// For now, we'll just log the update attempt
	m.logger.Info("Updating feature", "name", name, "current_version", feature.Version)

	// Reload the feature to get the latest version
	if err := m.loadFeature(name); err != nil {
		return fmt.Errorf("failed to reload feature: %w", err)
	}

	m.logger.Info("Feature updated successfully", "name", name, "new_version", m.features[name].Version)
	return nil
}

// autoUpdateFeatures checks for and applies updates to features
func (m *Manager) autoUpdateFeatures() {
	if !m.config.GetAutoUpdate() {
		return
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Use range over the ticker channel
	for range ticker.C {
		m.logger.Debug("Checking for feature updates")
		// Check and update all features
		for name := range m.features {
			if err := m.UpdateFeature(name); err != nil {
				m.logger.Error("Failed to update feature", "name", name, "error", err)
			}
		}
	}
}
