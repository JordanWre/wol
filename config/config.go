package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

const (
	koanfDelimiter = "."
	koanfTag       = "koanf"
	configFilename = "config.yaml"
)

var k = koanf.New(koanfDelimiter)

// Machine represents a machine to wake up
type Machine struct {
	// Name of the machine
	Name string `koanf:"name"`
	// MAC address of the machine
	Mac string `koanf:"mac"`
}

// Server represents the server configuration
type Server struct {
	// Listen address for the server
	Listen string `koanf:"listen"`
}

// Config represents the configuration for the application
type Config struct {
	// Machines represents the list of machines to wake up
	Machines []Machine `koanf:"machines"`
	// Server represents the server configuration
	Server Server `koanf:"server"`
}

// NewConfig creates a new Config instance
func NewConfig() *Config {
	return &Config{}
}

// Load loads the configuration from the config file
//
// The configuration file is searched in user's home directory and the current. Files are loaded in order and merged.
func (c *Config) Load() error {
	// Load defaults first
	defaults := &Config{
		Server: Server{
			Listen: ":7777",
		},
	}
	err := k.Load(structs.Provider(defaults, koanfTag), nil)
	if err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Order here matters as later values will override earlier ones
	paths := []string{
		filepath.Join(home, ".wol", configFilename),
		filepath.Join(".", configFilename),
	}

	for _, path := range paths {
		err = k.Load(file.Provider(path), yaml.Parser())

		// Ignore error if file does not exist
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	err = k.Unmarshal("", c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
