package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	MediaDir         string `yaml:"media_dir"`
	MetadataDir      string `yaml:"metadata_dir"`
	TranscriptionDir string `yaml:"transcription_dir"`
}

// AppConfig holds the global configuration
var AppConfig *Config

// loadConfig loads configuration from file, environment variables, and command line flags
func loadConfig() (*Config, error) {
	config := &Config{}

	// Set default values to ~/.local/share/reelarc/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	defaultBaseDir := filepath.Join(homeDir, ".local", "share", "reelarc")
	config.MediaDir = filepath.Join(defaultBaseDir, "media")
	config.MetadataDir = filepath.Join(defaultBaseDir, "metadata")
	config.TranscriptionDir = filepath.Join(defaultBaseDir, "transcription")

	// Parse command line flags
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.StringVar(&config.MediaDir, "media-dir", config.MediaDir, "Directory for media files")
	flag.StringVar(&config.MetadataDir, "metadata-dir", config.MetadataDir, "Directory for metadata files")
	flag.StringVar(&config.TranscriptionDir, "transcription-dir", config.TranscriptionDir, "Directory for transcription files")
	flag.Parse()

	// Check for config file path from environment variable if not set via flag
	if configFile == "" {
		if envConfigFile := os.Getenv("REELARC_CONFIG"); envConfigFile != "" {
			configFile = envConfigFile
		} else {
			// Default config file location
			configFile = filepath.Join(homeDir, ".config", "reelarc", "config.yaml")
		}
	}

	// Load config file if it exists
	if _, err := os.Stat(configFile); err == nil {
		if err := loadConfigFromFile(configFile, config); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %v", configFile, err)
		}
	}

	// Override with environment variables
	if envMediaDir := os.Getenv("REELARC_MEDIA_DIR"); envMediaDir != "" {
		config.MediaDir = envMediaDir
	}
	if envMetadataDir := os.Getenv("REELARC_METADATA_DIR"); envMetadataDir != "" {
		config.MetadataDir = envMetadataDir
	}
	if envTranscriptionDir := os.Getenv("REELARC_TRANSCRIPTION_DIR"); envTranscriptionDir != "" {
		config.TranscriptionDir = envTranscriptionDir
	}

	// Ensure all directories are absolute paths
	config.MediaDir, err = filepath.Abs(config.MediaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve media directory path: %v", err)
	}
	config.MetadataDir, err = filepath.Abs(config.MetadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve metadata directory path: %v", err)
	}
	config.TranscriptionDir, err = filepath.Abs(config.TranscriptionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve transcription directory path: %v", err)
	}

	return config, nil
}

// loadConfigFromFile loads configuration from a YAML file
func loadConfigFromFile(filename string, config *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}
