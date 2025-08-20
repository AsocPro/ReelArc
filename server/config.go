package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	MediaDir         string   `yaml:"media_dir"`
	MetadataDir      string   `yaml:"metadata_dir"`
	TranscriptionDir string   `yaml:"transcription_dir"`
	ExternalMarkdownDirs []string `yaml:"external_markdown_dirs"`
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
	var configDir string
	var externalMarkdownDirs string
	flag.StringVar(&configDir, "config", "", "Path to configuration directory")
	flag.StringVar(&config.MediaDir, "media-dir", config.MediaDir, "Directory for media files")
	flag.StringVar(&config.MetadataDir, "metadata-dir", config.MetadataDir, "Directory for metadata files")
	flag.StringVar(&config.TranscriptionDir, "transcription-dir", config.TranscriptionDir, "Directory for transcription files")
	flag.StringVar(&externalMarkdownDirs, "external-markdown-dirs", "", "Comma-separated list of directories to monitor for external markdown files")
	flag.Parse()

	// Parse external markdown directories from command line flag
	if externalMarkdownDirs != "" {
		dirs := strings.Split(externalMarkdownDirs, ",")
		for i, dir := range dirs {
			dirs[i] = strings.TrimSpace(dir)
		}
		config.ExternalMarkdownDirs = dirs
	}

	// Check for config directory from environment variable if not set via flag
	if configDir == "" {
		if envConfigDir := os.Getenv("REELARC_CONFIG"); envConfigDir != "" {
			configDir = envConfigDir
		} else {
			// Default config directory location
			configDir = filepath.Join(homeDir, ".config", "reelarc")
		}
	}

	// Construct config file path from config directory
	configFile := filepath.Join(configDir, "config.yaml")

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
	if envExternalMarkdownDirs := os.Getenv("REELARC_EXTERNAL_MARKDOWN_DIRS"); envExternalMarkdownDirs != "" {
		dirs := strings.Split(envExternalMarkdownDirs, ",")
		for i, dir := range dirs {
			dirs[i] = strings.TrimSpace(dir)
		}
		config.ExternalMarkdownDirs = dirs
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
	
	// Resolve external markdown directories to absolute paths
	for i, dir := range config.ExternalMarkdownDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve external markdown directory path %s: %v", dir, err)
		}
		config.ExternalMarkdownDirs[i] = absDir
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
