package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"plugin"
	"sync"
	"time"
	"unsafe"
)

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"

var (
	ErrPluginNotLoaded   = errors.New("plugin not loaded")
	ErrInvalidPluginType = errors.New("invalid plugin type")
	ErrPluginNotFound    = errors.New("plugin not found")
	ErrPluginInitFailed  = errors.New("plugin initialization failed")
)

// PluginManager manages loaded plugins
type PluginManager struct {
	signerPlugin SignerPlugin
	pluginInfo   *PluginInfo
	mu           sync.RWMutex
	config       PluginConfig
}

var (
	instance *PluginManager
	once     sync.Once
)

// Instance returns the singleton PluginManager
func Instance() *PluginManager {
	once.Do(func() {
		instance = &PluginManager{
			config: DefaultPluginConfig(),
		}
	})
	return instance
}

// LoadSignerPlugin loads a signing plugin from a shared library
func (pm *PluginManager) LoadSignerPlugin(path string, config PluginConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", path, err)
	}

	// Look up the exported SignerPlugin symbol
	sym, err := p.Lookup("SignerPlugin")
	if err != nil {
		return fmt.Errorf("failed to find SignerPlugin symbol: %w", err)
	}

	// Type assert to SignerPlugin
	signer, ok := sym.(SignerPlugin)
	if !ok {
		return ErrInvalidPluginType
	}

	// Look up the Initialize function
	initSym, err := p.Lookup("Initialize")
	if err != nil {
		return fmt.Errorf("failed to find Initialize symbol: %w", err)
	}

	initFunc, ok := initSym.(func(*C.char) error)
	if !ok {
		return ErrInvalidPluginType
	}

	// Convert config to C string
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	cConfig := C.CString(string(configJSON))
	defer C.free(unsafe.Pointer(cConfig))

	// Call Initialize
	if err := initFunc(cConfig); err != nil {
		return fmt.Errorf("plugin initialization failed: %w", err)
	}

	// Store the plugin and info
	pm.signerPlugin = signer
	pm.config = config
	pm.pluginInfo = &PluginInfo{
		Name:          signer.AlgorithmName(),
		Version:       "1.0.0", // Could be extracted from plugin metadata
		Algorithm:     signer.AlgorithmName(),
		SecurityLevel: signer.SecurityLevel(),
		LoadedAt:      time.Now(),
		Path:          path,
	}

	return nil
}

// GetSigner returns the loaded signer plugin
func (pm *PluginManager) GetSigner() (SignerPlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.signerPlugin == nil {
		return nil, ErrPluginNotLoaded
	}
	return pm.signerPlugin, nil
}

// GetPluginInfo returns information about the loaded plugin
func (pm *PluginManager) GetPluginInfo() (*PluginInfo, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.pluginInfo == nil {
		return nil, ErrPluginNotLoaded
	}
	return pm.pluginInfo, nil
}

// GetConfig returns the current plugin configuration
func (pm *PluginManager) GetConfig() PluginConfig {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.config
}

// IsPluginLoaded returns true if a plugin is currently loaded
func (pm *PluginManager) IsPluginLoaded() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.signerPlugin != nil
}

// UnloadPlugin unloads the current plugin
func (pm *PluginManager) UnloadPlugin() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.signerPlugin = nil
	pm.pluginInfo = nil
}

// ReloadPlugin reloads the plugin with new configuration
func (pm *PluginManager) ReloadPlugin(path string, config PluginConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Unload current plugin
	pm.signerPlugin = nil
	pm.pluginInfo = nil

	// Load new plugin
	return pm.loadSignerPluginInternal(path, config)
}

// loadSignerPluginInternal is the internal implementation without locking
func (pm *PluginManager) loadSignerPluginInternal(path string, config PluginConfig) error {
	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", path, err)
	}

	// Look up the exported SignerPlugin symbol
	sym, err := p.Lookup("SignerPlugin")
	if err != nil {
		return fmt.Errorf("failed to find SignerPlugin symbol: %w", err)
	}

	// Type assert to SignerPlugin
	signer, ok := sym.(SignerPlugin)
	if !ok {
		return ErrInvalidPluginType
	}

	// Look up the Initialize function
	initSym, err := p.Lookup("Initialize")
	if err != nil {
		return fmt.Errorf("failed to find Initialize symbol: %w", err)
	}

	initFunc, ok := initSym.(func(*C.char) error)
	if !ok {
		return ErrInvalidPluginType
	}

	// Convert config to C string
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	cConfig := C.CString(string(configJSON))
	defer C.free(unsafe.Pointer(cConfig))

	// Call Initialize
	if err := initFunc(cConfig); err != nil {
		return fmt.Errorf("plugin initialization failed: %w", err)
	}

	// Store the plugin and info
	pm.signerPlugin = signer
	pm.config = config
	pm.pluginInfo = &PluginInfo{
		Name:          signer.AlgorithmName(),
		Version:       "1.0.0",
		Algorithm:     signer.AlgorithmName(),
		SecurityLevel: signer.SecurityLevel(),
		LoadedAt:      time.Now(),
		Path:          path,
	}

	return nil
}

// ValidatePlugin validates that a plugin implements the required interface
func (pm *PluginManager) ValidatePlugin(path string) error {
	// Load the plugin temporarily
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", path, err)
	}

	// Check for required symbols
	requiredSymbols := []string{"SignerPlugin", "Initialize"}
	for _, symbol := range requiredSymbols {
		if _, err := p.Lookup(symbol); err != nil {
			return fmt.Errorf("missing required symbol %s: %w", symbol, err)
		}
	}

	// Try to get the SignerPlugin
	sym, err := p.Lookup("SignerPlugin")
	if err != nil {
		return fmt.Errorf("failed to find SignerPlugin symbol: %w", err)
	}

	// Type assert to SignerPlugin
	signer, ok := sym.(SignerPlugin)
	if !ok {
		return ErrInvalidPluginType
	}

	// Basic validation
	if signer.AlgorithmName() == "" {
		return errors.New("plugin must return a valid algorithm name")
	}

	if signer.SignatureSize() <= 0 {
		return errors.New("plugin must return a valid signature size")
	}

	if signer.PublicKeySize() <= 0 {
		return errors.New("plugin must return a valid public key size")
	}

	return nil
} 