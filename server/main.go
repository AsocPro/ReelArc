package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

// TimelineItem represents a single item in the timeline
type TimelineItem struct {
	ID        string `yaml:"id" json:"id"`
	Content   string `json:"content"` // This will be stored in the Markdown body
	Start     string `yaml:"start" json:"start"`
	End       string `yaml:"end,omitempty" json:"end,omitempty"`
	Type      string `yaml:"type,omitempty" json:"type,omitempty"`
	MediaPath string `yaml:"mediapath,omitempty" json:"mediaPath,omitempty"`
}

// MediaMetadata represents metadata for a media file
type MediaMetadata struct {
	ID            string            `yaml:"id" json:"id"`
	Filename      string            `yaml:"filename" json:"filename"`
	Path          string            `yaml:"path" json:"path"`
	Type          string            `yaml:"type" json:"type"`
	Timestamp     string            `yaml:"timestamp" json:"timestamp"`
	Duration      float64           `yaml:"duration,omitempty" json:"duration,omitempty"`
	Transcription string            `json:"transcription"` // This will be stored in the Markdown body
	Notes         string            `json:"notes"`         // Same as transcription but more explicit for notes field
	Labels        []string          `yaml:"labels" json:"labels"`
	Transcripts   []TranscriptEntry `yaml:"transcripts,omitempty" json:"transcripts,omitempty"`
}

// MediaItem represents a media item in the mock data
type MediaItem struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	Timestamp     string   `json:"timestamp"`
	Duration      float64  `json:"duration,omitempty"`
	Filename      string   `json:"filename"`
	Transcription string   `json:"transcription"`
	Labels        []string `json:"labels"`
}

// TranscriptEntry represents a single transcript entry
type TranscriptEntry struct {
	Start    float64 `yaml:"start" json:"start"`
	End      float64 `yaml:"end" json:"end"`
	Text     string  `yaml:"text" json:"text"`
	Segment  int     `yaml:"segment" json:"segment"`
	Speaker  string  `yaml:"speaker,omitempty" json:"speaker,omitempty"`
	Metadata string  `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// RelativeDate represents relative date parameters
type RelativeDate struct {
	Period    string `json:"period"`
	Offset    int    `json:"offset,omitempty"`
	Anchor    string `json:"anchor,omitempty"`
	Count     int    `json:"count,omitempty"`
	Direction string `json:"direction,omitempty"`
}

// FilterDateRange represents date range criteria for filters
type FilterDateRange struct {
	Type      string `yaml:"type" json:"type"`
	// For fixed dates
	StartDate string `yaml:"startDate,omitempty" json:"startDate,omitempty"`
	EndDate   string `yaml:"endDate,omitempty" json:"endDate,omitempty"`
	// For simple relative dates (backward compatibility)
	Period    string `yaml:"period,omitempty" json:"period,omitempty"`
	Offset    int    `yaml:"offset,omitempty" json:"offset,omitempty"`
	Anchor    string `yaml:"anchor,omitempty" json:"anchor,omitempty"`
	Count     int    `yaml:"count,omitempty" json:"count,omitempty"`
	Direction string `yaml:"direction,omitempty" json:"direction,omitempty"`
	// For complex relative date ranges
	StartRelative *RelativeDate `json:"startRelative,omitempty"`
	EndRelative   *RelativeDate `json:"endRelative,omitempty"`
}

// FilterCriteria represents filter criteria
type FilterCriteria struct {
	DateRange  *FilterDateRange `yaml:"dateRange" json:"dateRange"`
	Labels     []string         `yaml:"labels" json:"labels"`
	MediaTypes []string         `yaml:"mediaTypes" json:"mediaTypes"`
}

// Filter represents a single quick filter
type Filter struct {
	ID          string         `yaml:"id" json:"id"`
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Type        string         `yaml:"type" json:"type"`
	Icon        string         `yaml:"icon" json:"icon"`
	Enabled     bool           `yaml:"enabled" json:"enabled"`
	Criteria    FilterCriteria `yaml:"criteria" json:"criteria"`
}

// FiltersMetadata represents metadata about the filters configuration
type FiltersMetadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Created     string `yaml:"created" json:"created"`
	Updated     string `yaml:"updated" json:"updated"`
}

// FiltersConfig represents the complete filters configuration
type FiltersConfig struct {
	Version       string          `yaml:"version" json:"version"`
	Metadata      FiltersMetadata `yaml:"metadata" json:"metadata"`
	DefaultFilter string          `yaml:"defaultFilter" json:"defaultFilter"`
	DisplayOrder  []string        `yaml:"displayOrder" json:"displayOrder"`
	Filters       []Filter        `yaml:"filters" json:"filters"`
}

const (
	clientDir = "./client/dist"

	// File extensions
	mdExt = ".md"
)

// Global directory variables (set from config)
var (
	mediaDir     string
	metadataDir  string
	timelineDir  string
	timelineFile string
	configDir    string
)

// ApplyFilterCriteria applies FilterCriteria to a MediaMetadata item
func ApplyFilterCriteria(metadata MediaMetadata, criteria FilterCriteria) bool {
	// Apply date range filtering
	if criteria.DateRange != nil {
		if !applyDateRangeFilter(metadata, *criteria.DateRange) {
			return false
		}
	}

	// Apply label filtering
	if len(criteria.Labels) > 0 {
		if !applyLabelsFilter(metadata, criteria.Labels) {
			return false
		}
	}

	// Apply media type filtering
	if len(criteria.MediaTypes) > 0 {
		if !applyMediaTypesFilter(metadata, criteria.MediaTypes) {
			return false
		}
	}

	return true
}

// applyDateRangeFilter applies date range filtering to a metadata item
func applyDateRangeFilter(metadata MediaMetadata, dateRange FilterDateRange) bool {
	itemTime, parseErr := time.Parse(time.RFC3339, metadata.Timestamp)
	if parseErr != nil {
		log.Printf("Failed to parse timestamp: %v", parseErr)
		return false
	}

	switch dateRange.Type {
	case "fixed":
		return applyFixedDateRange(itemTime, dateRange)
	case "relative":
		return applyRelativeDateRange(itemTime, dateRange)
	default:
		log.Printf("Unknown date range type: %s", dateRange.Type)
		return false
	}
}

// applyFixedDateRange applies fixed date range filtering
func applyFixedDateRange(itemTime time.Time, dateRange FilterDateRange) bool {
	if dateRange.StartDate != "" {
		startTime, err := time.Parse(time.RFC3339, dateRange.StartDate)
		if err != nil {
			log.Printf("Failed to parse start date: %v", err)
			return false
		}
		if itemTime.Before(startTime) {
			return false
		}
	}

	if dateRange.EndDate != "" {
		endTime, err := time.Parse(time.RFC3339, dateRange.EndDate)
		if err != nil {
			log.Printf("Failed to parse end date: %v", err)
			return false
		}
		if itemTime.After(endTime) {
			return false
		}
	}

	return true
}

// applyRelativeDateRange applies relative date range filtering
func applyRelativeDateRange(itemTime time.Time, dateRange FilterDateRange) bool {
	now := time.Now()
	
	var startTime, endTime time.Time
	
	// Check if this is a range with separate start and end relative dates
	if dateRange.StartRelative != nil && dateRange.EndRelative != nil {
		// Calculate start time from startRelative
		startTime = calculateRelativeTime(now, *dateRange.StartRelative)
		// Calculate end time from endRelative  
		endTime = calculateRelativeTime(now, *dateRange.EndRelative)
		
		// Ensure startTime is before endTime
		if startTime.After(endTime) {
			startTime, endTime = endTime, startTime
		}
	} else {
		// Backward compatibility: use the original single-period logic
		switch dateRange.Period {
		case "day":
			if dateRange.Offset == 0 {
				// Today
				startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				endTime = startTime.Add(24 * time.Hour)
			} else {
				// Specific day offset
				targetDate := now.AddDate(0, 0, dateRange.Offset)
				startTime = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
				endTime = startTime.Add(24 * time.Hour)
			}
		case "days":
			if dateRange.Count > 0 {
				// Last N days (using count)
				if dateRange.Direction == "backward" || dateRange.Direction == "" {
					endTime = now
					startTime = now.AddDate(0, 0, -dateRange.Count)
				}
			} else if dateRange.Offset < 0 {
				// Last N days (using offset)
				endTime = now
				startTime = now.AddDate(0, 0, dateRange.Offset)
			}
		case "week":
			weekStart := getWeekStart(now)
			if dateRange.Offset == 0 {
				// This week
				startTime = weekStart
				endTime = weekStart.AddDate(0, 0, 7)
			} else {
				// Week offset
				targetWeek := weekStart.AddDate(0, 0, dateRange.Offset*7)
				startTime = targetWeek
				endTime = targetWeek.AddDate(0, 0, 7)
			}
		case "month":
			if dateRange.Offset == 0 {
				// This month
				startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				endTime = startTime.AddDate(0, 1, 0)
			} else {
				// Month offset
				targetMonth := now.AddDate(0, dateRange.Offset, 0)
				startTime = time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, targetMonth.Location())
				endTime = startTime.AddDate(0, 1, 0)
			}
		default:
			log.Printf("Unknown period type: %s", dateRange.Period)
			return false
		}
	}

	return (itemTime.After(startTime) || itemTime.Equal(startTime)) && itemTime.Before(endTime)
}

// calculateRelativeTime calculates a specific time based on relative parameters
func calculateRelativeTime(now time.Time, relative RelativeDate) time.Time {
	switch relative.Period {
	case "day":
		if relative.Offset == 0 {
			// Today
			return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		} else {
			// Specific day offset
			targetDate := now.AddDate(0, 0, relative.Offset)
			return time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		}
	case "days":
		if relative.Count > 0 {
			// N days in specified direction
			if relative.Direction == "backward" || relative.Direction == "" {
				return now.AddDate(0, 0, -relative.Count)
			} else {
				return now.AddDate(0, 0, relative.Count)
			}
		} else if relative.Offset != 0 {
			// Days offset
			return now.AddDate(0, 0, relative.Offset)
		}
		return now
	case "week":
		weekStart := getWeekStart(now)
		if relative.Offset == 0 {
			// This week
			return weekStart
		} else {
			// Week offset
			return weekStart.AddDate(0, 0, relative.Offset*7)
		}
	case "month":
		if relative.Offset == 0 {
			// This month
			return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		} else {
			// Month offset
			targetMonth := now.AddDate(0, relative.Offset, 0)
			return time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, targetMonth.Location())
		}
	default:
		log.Printf("Unknown period type: %s", relative.Period)
		return now
	}
}

// getWeekStart returns the start of the week (Monday) for the given time
func getWeekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	daysToMonday := weekday - 1
	monday := t.AddDate(0, 0, -daysToMonday)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// applyLabelsFilter applies label filtering to a metadata item
func applyLabelsFilter(metadata MediaMetadata, filterLabels []string) bool {
	if len(filterLabels) == 0 {
		return true
	}

	for _, filterLabel := range filterLabels {
		for _, itemLabel := range metadata.Labels {
			if strings.EqualFold(strings.TrimSpace(itemLabel), strings.TrimSpace(filterLabel)) {
				return true
			}
		}
	}
	return false
}

// applyMediaTypesFilter applies media type filtering to a metadata item
func applyMediaTypesFilter(metadata MediaMetadata, filterTypes []string) bool {
	if len(filterTypes) == 0 {
		return true
	}

	for _, filterType := range filterTypes {
		if strings.EqualFold(metadata.Type, filterType) {
			return true
		}
	}
	return false
}

// LoadFiltersConfig loads the filters configuration from filters.yaml
func LoadFiltersConfig() (*FiltersConfig, error) {
	filtersPath := filepath.Join(configDir, "filters.yaml")
	
	data, err := os.ReadFile(filtersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read filters file: %v", err)
	}

	var config FiltersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse filters YAML: %v", err)
	}

	return &config, nil
}

// GetFilterByID retrieves a specific filter by its ID from the filters configuration
func GetFilterByID(filterID string) (*Filter, error) {
	config, err := LoadFiltersConfig()
	if err != nil {
		return nil, err
	}

	for _, filter := range config.Filters {
		if filter.ID == filterID {
			return &filter, nil
		}
	}

	return nil, fmt.Errorf("filter with ID %s not found", filterID)
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	AppConfig = config

	// Set global directory variables from config
	mediaDir = config.MediaDir
	metadataDir = config.MetadataDir
	timelineDir = filepath.Join(config.MetadataDir, "timeline")
	timelineFile = filepath.Join(config.MetadataDir, "timeline.md")
	
	// Set config directory (same logic as in config.go)
	if envConfigDir := os.Getenv("REELARC_CONFIG"); envConfigDir != "" {
		configDir = envConfigDir
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home directory: %v", err)
		}
		configDir = filepath.Join(homeDir, ".config", "reelarc")
	}

	// Ensure data directories exist
	ensureDirectories()

	// Initialize transcription system
	InitTranscriptionSystem()

	// API routes
	http.HandleFunc("/api/timeline", handleTimeline)
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/metadata/", handleMetadata)
	http.HandleFunc("/api/media", handleMedia)
	http.HandleFunc("/api/transcription/status", handleTranscriptionStatus)
	http.HandleFunc("/api/labels/update", handleUpdateLabels)
	http.HandleFunc("/api/filters", handleFilters)
	http.HandleFunc("/api/filters/reorder", handleFiltersReorder)
	http.HandleFunc("/api/filters/", handleFiltersWithID)

	// Serve media files
	http.HandleFunc("/media/", handleMediaFiles)

	// Serve static files in production
	http.HandleFunc("/", handleStaticFiles)

	// Start server
	port := 8080
	fmt.Printf("Server starting on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func ensureDirectories() {
	dirs := []string{mediaDir, metadataDir, timelineDir, AppConfig.TranscriptionDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

// Helper function to read a Markdown file with frontmatter
func readMarkdownFile(filePath string, data interface{}) (string, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Parse frontmatter
	body, err := frontmatter.Parse(bytes.NewReader(content), data)
	if err != nil {
		return "", fmt.Errorf("failed to parse frontmatter: %v", err)
	}

	return string(body), nil
}

// scanMarkdownFiles scans both metadata directory and external markdown directories for markdown files
func scanMarkdownFiles(filterFunc func(MediaMetadata) bool) ([]MediaMetadata, error) {
	allMetadata := make([]MediaMetadata, 0)
	
	// Scan the main metadata directory
	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %v", err)
	}
	
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), mdExt) {
			filePath := filepath.Join(metadataDir, file.Name())
			metadata, err := readMetadataFromFile(filePath)
			if err != nil {
				log.Printf("Failed to read metadata file %s: %v", file.Name(), err)
				continue
			}
			
			if filterFunc == nil || filterFunc(metadata) {
				allMetadata = append(allMetadata, metadata)
			}
		}
	}
	
	// Scan external markdown directories
	for _, externalDir := range AppConfig.ExternalMarkdownDirs {
		// Check if directory exists
		if _, err := os.Stat(externalDir); os.IsNotExist(err) {
			log.Printf("External markdown directory does not exist: %s", externalDir)
			continue
		}
		
		files, err := os.ReadDir(externalDir)
		if err != nil {
			log.Printf("Failed to read external markdown directory %s: %v", externalDir, err)
			continue
		}
		
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), mdExt) {
				filePath := filepath.Join(externalDir, file.Name())
				metadata, err := readMetadataFromFile(filePath)
				if err != nil {
					log.Printf("Failed to read external markdown file %s: %v", file.Name(), err)
					continue
				}
				
				if filterFunc == nil || filterFunc(metadata) {
					allMetadata = append(allMetadata, metadata)
				}
			}
		}
	}
	
	return allMetadata, nil
}

// readMetadataFromFile reads metadata from a markdown file
func readMetadataFromFile(filePath string) (MediaMetadata, error) {
	var metadata MediaMetadata
	
	// Read Markdown file with frontmatter
	content, readErr := readMarkdownFile(filePath, &metadata)
	if readErr != nil {
		return metadata, readErr
	}
	metadata.Transcription = content
	metadata.Notes = content // Set notes field to same content
	
	// Ensure Labels is never nil
	if metadata.Labels == nil {
		metadata.Labels = []string{}
	}
	
	return metadata, nil
}

// findMetadataFileByID searches for a metadata file by ID in both metadata directory and external directories
func findMetadataFileByID(id string) (string, MediaMetadata, error) {
	var metadata MediaMetadata
	
	// Search in main metadata directory
	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return "", metadata, fmt.Errorf("failed to read metadata directory: %v", err)
	}
	
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), mdExt) {
			filePath := filepath.Join(metadataDir, file.Name())
			
			tempMetadata, readErr := readMetadataFromFile(filePath)
			if readErr != nil {
				log.Printf("Failed to read metadata file %s: %v", file.Name(), readErr)
				continue
			}
			
			if tempMetadata.ID == id {
				return filePath, tempMetadata, nil
			}
		}
	}
	
	// Search in external markdown directories
	for _, externalDir := range AppConfig.ExternalMarkdownDirs {
		// Check if directory exists
		if _, err := os.Stat(externalDir); os.IsNotExist(err) {
			continue
		}
		
		files, err := os.ReadDir(externalDir)
		if err != nil {
			log.Printf("Failed to read external markdown directory %s: %v", externalDir, err)
			continue
		}
		
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), mdExt) {
				filePath := filepath.Join(externalDir, file.Name())
				
				tempMetadata, readErr := readMetadataFromFile(filePath)
				if readErr != nil {
					log.Printf("Failed to read external markdown file %s: %v", file.Name(), readErr)
					continue
				}
				
				if tempMetadata.ID == id {
					return filePath, tempMetadata, nil
				}
			}
		}
	}
	
	return "", metadata, fmt.Errorf("metadata with ID %s not found", id)
}

// Helper function to write a Markdown file with frontmatter
func writeMarkdownFile(filePath string, data interface{}, body string) error {
	// Create a buffer to store the file content
	var buf bytes.Buffer

	// Write frontmatter with delimiters
	buf.WriteString("---\n")

	// If data is not nil, write the YAML frontmatter
	if data != nil {
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal frontmatter data to YAML: %v", err)
		}
		buf.Write(yamlData)
	}

	buf.WriteString("---\n\n")

	// Write body
	if body != "" {
		buf.WriteString(body)
	}

	// Write to file
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

func handleTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read from individual Markdown files in the timeline directory
	files, err := os.ReadDir(timelineDir)
	if err != nil {
		http.Error(w, "Failed to read timeline data", http.StatusInternalServerError)
		return
	}

	var items []TimelineItem
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), mdExt) {
			filePath := filepath.Join(timelineDir, file.Name())
			var item TimelineItem
			content, err := readMarkdownFile(filePath, &item)
			if err != nil {
				log.Printf("Failed to read timeline item %s: %v", file.Name(), err)
				continue
			}
			item.Content = content
			items = append(items, item)
		}
	}

	// Marshal the combined items
	data, err := json.Marshal(items)
	if err != nil {
		http.Error(w, "Failed to marshal timeline data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form, 50 MB max (increased for multiple files)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get all files from the form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	// Response data for all files
	type FileResponse struct {
		Status   string `json:"status"`
		Filename string `json:"filename"`
		Path     string `json:"path"`
		Metadata string `json:"metadata"`
	}
	responses := make([]FileResponse, 0, len(files))

	// Process each file
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file %s: %v", fileHeader.Filename, err)
			continue
		}
		defer file.Close()

		// Create file path
		filename := fileHeader.Filename
		filePath := filepath.Join(mediaDir, filename)

		// Create a temporary buffer to store the file content
		// We need this to read EXIF data and then save the file
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", filename, err)
			continue
		}

		// Create file
		dst, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating file %s: %v", filename, err)
			continue
		}
		defer dst.Close()

		// Copy file content
		if _, err := dst.Write(fileBytes); err != nil {
			log.Printf("Error saving file %s: %v", filename, err)
			continue
		}

		// Create metadata
		mediaType := "unknown"
		if strings.HasSuffix(strings.ToLower(filename), ".mp3") || strings.HasSuffix(strings.ToLower(filename), ".wav") {
			mediaType = "audio"
		} else if strings.HasSuffix(strings.ToLower(filename), ".mp4") || strings.HasSuffix(strings.ToLower(filename), ".mov") {
			mediaType = "video"
		} else if strings.HasSuffix(strings.ToLower(filename), ".jpg") || strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
			mediaType = "photo"
		} else if strings.HasSuffix(strings.ToLower(filename), ".md") || strings.HasSuffix(strings.ToLower(filename), ".markdown") {
			mediaType = "note"
		}

		// Try to extract timestamp from EXIF data for photos and videos
		timestamp := time.Now().Format(time.RFC3339)
		log.Printf("Processing EXIF data for file: %s (type: %s)", filename, mediaType)

		if mediaType == "photo" || mediaType == "video" {
			// Use exiftool to extract metadata in JSON format
			log.Printf("Running exiftool on file: %s", filePath)
			cmd := exec.Command("exiftool", "-json", filePath)
			output, err := cmd.Output()
			if err != nil {
				log.Printf("Error running exiftool: %v", err)
			} else {
				log.Printf("Exiftool output length: %d bytes", len(output))

				// Parse the JSON output
				var exifData []map[string]interface{}
				if err := json.Unmarshal(output, &exifData); err != nil {
					log.Printf("Error parsing exiftool JSON output: %v", err)
				} else if len(exifData) == 0 {
					log.Printf("No EXIF data found in exiftool output")
				} else {
					// Log available tags for debugging
					log.Printf("Available EXIF tags:")
					for key := range exifData[0] {
						log.Printf("  - %s: %v", key, exifData[0][key])
					}

					// Try to get DateTimeOriginal first
					if dateTimeStr, ok := exifData[0]["DateTimeOriginal"].(string); ok && dateTimeStr != "" {
						log.Printf("Found DateTimeOriginal: %s", dateTimeStr)
						// Parse the date string (format typically: "YYYY:MM:DD HH:MM:SS")
						if dateTime, err := time.Parse("2006:01:02 15:04:05", dateTimeStr); err != nil {
							log.Printf("Error parsing DateTimeOriginal: %v", err)
						} else {
							timestamp = dateTime.Format(time.RFC3339)
							log.Printf("Using DateTimeOriginal as timestamp: %s", timestamp)
						}
					} else if createDateStr, ok := exifData[0]["CreateDate"].(string); ok && createDateStr != "" {
						// Fallback to CreateDate if DateTimeOriginal doesn't exist
						log.Printf("DateTimeOriginal not found, using CreateDate: %s", createDateStr)
						if dateTime, err := time.Parse("2006:01:02 15:04:05", createDateStr); err != nil {
							log.Printf("Error parsing CreateDate: %v", err)
						} else {
							timestamp = dateTime.Format(time.RFC3339)
							log.Printf("Using CreateDate as timestamp: %s", timestamp)
						}
					} else {
						log.Printf("Neither DateTimeOriginal nor CreateDate found in EXIF data")
					}
				}
			}
		} else if mediaType == "note" {
			// For markdown files, try to extract existing metadata from frontmatter
			log.Printf("Processing markdown file for existing metadata: %s", filename)
			var existingMetadata MediaMetadata
			body, err := readMarkdownFile(filePath, &existingMetadata)
			if err == nil && existingMetadata.Timestamp != "" {
				// Use existing timestamp if present
				timestamp = existingMetadata.Timestamp
				log.Printf("Found existing timestamp in markdown frontmatter: %s", timestamp)
			} else {
				log.Printf("No existing timestamp found in markdown, using current time")
			}
			_ = body // We have the body content if needed later
		} else {
			log.Printf("Skipping EXIF extraction for non-photo/video file type: %s", mediaType)
		}

		log.Printf("Final timestamp for file %s: %s", filename, timestamp)

		metadata := MediaMetadata{
			ID:            fmt.Sprintf("%d", time.Now().UnixNano()),
			Filename:      filename,
			Path:          "/media/" + filename,
			Type:          mediaType,
			Timestamp:     timestamp,
			Transcription: "",
			Labels:        []string{},
		}

		// Save metadata as Markdown with frontmatter
		metadataPath := filepath.Join(metadataDir, filename+mdExt)

		// Create frontmatter data
		frontmatterData := struct {
			ID        string   `yaml:"id"`
			Filename  string   `yaml:"filename"`
			Path      string   `yaml:"path"`
			Type      string   `yaml:"type"`
			Timestamp string   `yaml:"timestamp"`
			Duration  float64  `yaml:"duration,omitempty"`
			Labels    []string `yaml:"labels"`
		}{
			ID:        metadata.ID,
			Filename:  metadata.Filename,
			Path:      metadata.Path,
			Type:      metadata.Type,
			Timestamp: metadata.Timestamp,
			Duration:  metadata.Duration,
			Labels:    metadata.Labels,
		}

		// Write the Markdown file with frontmatter
		if err := writeMarkdownFile(metadataPath, frontmatterData, metadata.Transcription); err != nil {
			log.Printf("Error saving metadata for %s: %v", filename, err)
			continue
		}

		// Add to transcription queue if it's an audio or video file
		if mediaType == "audio" || mediaType == "video" {
			log.Printf("Adding %s to transcription queue", filename)
			TQueue.AddToQueue(filename)
		}

		// Add to responses
		responses = append(responses, FileResponse{
			Status:   "success",
			Filename: filename,
			Path:     "/media/" + filename,
			Metadata: "/api/metadata/" + filename,
		})
	}

	// Return success response with all files
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"files":  responses,
		"count":  len(responses),
	})
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/api/metadata/")

	// If no filename is provided, return all metadata files (with optional filtering)
	if filename == "" {
		// Parse query parameters for filtering
		queryParams := r.URL.Query()
		filterID := queryParams.Get("filter")
		startDate := queryParams.Get("startDate")
		endDate := queryParams.Get("endDate")
		labelsParam := queryParams.Get("labels")
		mediaTypesParam := queryParams.Get("mediaTypes")
		dateRangeFilterParam := queryParams.Get("dateRangeFilter")

		var filterFunc func(MediaMetadata) bool

		// If a filter ID is provided, use the new filter spec
		if filterID != "" {
			filter, err := GetFilterByID(filterID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Filter not found: %v", err), http.StatusNotFound)
				return
			}
			
			if !filter.Enabled {
				http.Error(w, "Filter is disabled", http.StatusBadRequest)
				return
			}

			filterFunc = func(metadata MediaMetadata) bool {
				return ApplyFilterCriteria(metadata, filter.Criteria)
			}
		} else {
			// Fallback to legacy query parameter filtering for backward compatibility
			var filterLabels []string
			if labelsParam != "" {
				filterLabels = strings.Split(labelsParam, ",")
				// Trim whitespace from each label
				for i, label := range filterLabels {
					filterLabels[i] = strings.TrimSpace(label)
				}
			}

			var filterMediaTypes []string
			if mediaTypesParam != "" {
				filterMediaTypes = strings.Split(mediaTypesParam, ",")
				// Trim whitespace from each type
				for i, mediaType := range filterMediaTypes {
					filterMediaTypes[i] = strings.TrimSpace(mediaType)
				}
			}

			// Validate date format if provided (but we'll use them directly in FilterCriteria)
			if startDate != "" {
				if _, err := time.Parse(time.RFC3339, startDate); err != nil {
					http.Error(w, "Invalid startDate format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)", http.StatusBadRequest)
					return
				}
			}
			if endDate != "" {
				if _, err := time.Parse(time.RFC3339, endDate); err != nil {
					http.Error(w, "Invalid endDate format. Use RFC3339 format (e.g., 2023-12-31T23:59:59Z)", http.StatusBadRequest)
					return
				}
			}

			// Create legacy filter criteria
			var legacyDateRange *FilterDateRange
			
			// Check for relative date range filter parameter
			if dateRangeFilterParam != "" {
				var dateRangeFilter FilterDateRange
				if err := json.Unmarshal([]byte(dateRangeFilterParam), &dateRangeFilter); err != nil {
					http.Error(w, "Invalid dateRangeFilter format", http.StatusBadRequest)
					return
				}
				legacyDateRange = &dateRangeFilter
			} else if startDate != "" || endDate != "" {
				legacyDateRange = &FilterDateRange{
					Type:      "fixed",
					StartDate: startDate,
					EndDate:   endDate,
				}
			}

			legacyCriteria := FilterCriteria{
				DateRange:  legacyDateRange,
				Labels:     filterLabels,
				MediaTypes: filterMediaTypes,
			}

			filterFunc = func(metadata MediaMetadata) bool {
				return ApplyFilterCriteria(metadata, legacyCriteria)
			}
		}

		// Use the new scanning function to get all metadata files
		allMetadata, err := scanMarkdownFiles(filterFunc)
		if err != nil {
			http.Error(w, "Failed to read metadata files", http.StatusInternalServerError)
			return
		}

		// Marshal the combined metadata
		responseData, err := json.Marshal(allMetadata)
		if err != nil {
			http.Error(w, "Failed to marshal metadata", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
		return
	}

	// If a filename is provided, return that specific metadata file
	metadataPath := filepath.Join(metadataDir, filename+mdExt)

	var metadata MediaMetadata
	// Read Markdown file with frontmatter
	content, readErr := readMarkdownFile(metadataPath, &metadata)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			http.Error(w, "Metadata not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to read metadata", http.StatusInternalServerError)
		}
		return
	}
	metadata.Transcription = content
	metadata.Notes = content // Set notes field to same content

	// Ensure Labels is never nil
	if metadata.Labels == nil {
		metadata.Labels = []string{}
	}

	// Marshal the metadata
	responseData, marshalErr := json.Marshal(metadata)
	if marshalErr != nil {
		http.Error(w, "Failed to marshal metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)
}
func handleMediaFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/media/")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(mediaDir, filename)
	http.ServeFile(w, r, filePath)
}

func handleMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters for filtering
	queryParams := r.URL.Query()
	filterID := queryParams.Get("filter")
	startDate := queryParams.Get("startDate")
	endDate := queryParams.Get("endDate")
	labelsParam := queryParams.Get("labels")
	mediaTypesParam := queryParams.Get("mediaTypes")
	dateRangeFilterParam := queryParams.Get("dateRangeFilter")

	var filterFunc func(MediaMetadata) bool

	// If a filter ID is provided, use the new filter spec
	if filterID != "" {
		filter, err := GetFilterByID(filterID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Filter not found: %v", err), http.StatusNotFound)
			return
		}
		
		if !filter.Enabled {
			http.Error(w, "Filter is disabled", http.StatusBadRequest)
			return
		}

		filterFunc = func(metadata MediaMetadata) bool {
			return ApplyFilterCriteria(metadata, filter.Criteria)
		}
	} else {
		// Fallback to legacy query parameter filtering for backward compatibility
		var filterLabels []string
		if labelsParam != "" {
			filterLabels = strings.Split(labelsParam, ",")
			// Trim whitespace from each label
			for i, label := range filterLabels {
				filterLabels[i] = strings.TrimSpace(label)
			}
		}

		var filterMediaTypes []string
		if mediaTypesParam != "" {
			filterMediaTypes = strings.Split(mediaTypesParam, ",")
			// Trim whitespace from each type
			for i, mediaType := range filterMediaTypes {
				filterMediaTypes[i] = strings.TrimSpace(mediaType)
			}
		}

		// Validate date format if provided (but we'll use them directly in FilterCriteria)
		if startDate != "" {
			if _, err := time.Parse(time.RFC3339, startDate); err != nil {
				http.Error(w, "Invalid startDate format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)", http.StatusBadRequest)
				return
			}
		}
		if endDate != "" {
			if _, err := time.Parse(time.RFC3339, endDate); err != nil {
				http.Error(w, "Invalid endDate format. Use RFC3339 format (e.g., 2023-12-31T23:59:59Z)", http.StatusBadRequest)
				return
			}
		}

		// Create legacy filter criteria
		var legacyDateRange *FilterDateRange
		
		// Check for relative date range filter parameter
		if dateRangeFilterParam != "" {
			var dateRangeFilter FilterDateRange
			if err := json.Unmarshal([]byte(dateRangeFilterParam), &dateRangeFilter); err != nil {
				http.Error(w, "Invalid dateRangeFilter format", http.StatusBadRequest)
				return
			}
			legacyDateRange = &dateRangeFilter
		} else if startDate != "" || endDate != "" {
			legacyDateRange = &FilterDateRange{
				Type:      "fixed",
				StartDate: startDate,
				EndDate:   endDate,
			}
		}

		legacyCriteria := FilterCriteria{
			DateRange:  legacyDateRange,
			Labels:     filterLabels,
			MediaTypes: filterMediaTypes,
		}

		filterFunc = func(metadata MediaMetadata) bool {
			return ApplyFilterCriteria(metadata, legacyCriteria)
		}
	}

	// Use the new scanning function to get all metadata files
	allMetadata, err := scanMarkdownFiles(filterFunc)
	if err != nil {
		http.Error(w, "Failed to read metadata files", http.StatusInternalServerError)
		return
	}

	// Marshal the combined metadata
	responseData, err := json.Marshal(allMetadata)
	if err != nil {
		http.Error(w, "Failed to marshal metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)
}

func handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// In development, this would proxy to the Bun dev server
	// In production, serve from the client/dist directory
	path := filepath.Join(clientDir, r.URL.Path)

	// If the file doesn't exist or is a directory, serve index.html
	info, err := os.Stat(path)
	if os.IsNotExist(err) || info.IsDir() {
		http.ServeFile(w, r, filepath.Join(clientDir, "index.html"))
		return
	}

	http.ServeFile(w, r, path)
}

// Handler for transcription status API
func handleTranscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all transcription statuses
	statuses := TQueue.GetAllStatuses()

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

// UpdateLabelsRequest represents the request body for updating labels
type UpdateLabelsRequest struct {
	ID     string   `json:"id"`
	Labels []string `json:"labels"`
}

// Handler for updating labels API
func handleUpdateLabels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req UpdateLabelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Find the metadata file by ID
	targetFile, metadata, err := findMetadataFileByID(req.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Media item not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to search for metadata", http.StatusInternalServerError)
		}
		return
	}

	// Update the labels
	metadata.Labels = req.Labels

	// Create frontmatter data for writing
	frontmatterData := struct {
		ID        string   `yaml:"id"`
		Filename  string   `yaml:"filename"`
		Path      string   `yaml:"path"`
		Type      string   `yaml:"type"`
		Timestamp string   `yaml:"timestamp"`
		Duration  float64  `yaml:"duration,omitempty"`
		Labels    []string `yaml:"labels"`
	}{
		ID:        metadata.ID,
		Filename:  metadata.Filename,
		Path:      metadata.Path,
		Type:      metadata.Type,
		Timestamp: metadata.Timestamp,
		Duration:  metadata.Duration,
		Labels:    metadata.Labels,
	}

	// Write the updated metadata back to the file
	if err := writeMarkdownFile(targetFile, frontmatterData, metadata.Transcription); err != nil {
		log.Printf("Error updating metadata file %s: %v", targetFile, err)
		http.Error(w, "Failed to update labels", http.StatusInternalServerError)
		return
	}

	// Return the updated metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// Handler for quick filters API
func handleFilters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetFilters(w, r)
	case http.MethodPost:
		handleSaveFilter(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler for GET /api/filters
func handleGetFilters(w http.ResponseWriter, r *http.Request) {
	// Construct the filters.yaml path from config directory
	filtersPath := filepath.Join(configDir, "filters.yaml")
	
	// Check if the filters.yaml file exists
	if _, err := os.Stat(filtersPath); os.IsNotExist(err) {
		http.Error(w, "Filters configuration not found", http.StatusNotFound)
		return
	}

	// Read the filters.yaml file
	data, err := os.ReadFile(filtersPath)
	if err != nil {
		log.Printf("Error reading filters file %s: %v", filtersPath, err)
		http.Error(w, "Failed to read filters configuration", http.StatusInternalServerError)
		return
	}

	// Parse the YAML data
	var filtersConfig FiltersConfig
	if err := yaml.Unmarshal(data, &filtersConfig); err != nil {
		log.Printf("Error parsing filters YAML: %v", err)
		http.Error(w, "Failed to parse filters configuration", http.StatusInternalServerError)
		return
	}

	// Return the filters configuration as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtersConfig)
}

// SaveFilterRequest represents the request body for saving a filter
type SaveFilterRequest struct {
	Filter Filter `json:"filter"`
}

// Handler for POST /api/filters - save a new filter
func handleSaveFilter(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req SaveFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Filter.ID == "" || req.Filter.Name == "" {
		http.Error(w, "Filter ID and name are required", http.StatusBadRequest)
		return
	}

	// Load current filters config
	config, err := LoadFiltersConfig()
	if err != nil {
		http.Error(w, "Failed to load filters configuration", http.StatusInternalServerError)
		return
	}

	// Check if filter already exists (update) or is new (add)
	filterExists := false
	for i, existingFilter := range config.Filters {
		if existingFilter.ID == req.Filter.ID {
			// Update existing filter
			config.Filters[i] = req.Filter
			filterExists = true
			break
		}
	}

	// If filter doesn't exist, add it
	if !filterExists {
		config.Filters = append(config.Filters, req.Filter)
		// Add to display order if not already there
		found := false
		for _, id := range config.DisplayOrder {
			if id == req.Filter.ID {
				found = true
				break
			}
		}
		if !found {
			config.DisplayOrder = append(config.DisplayOrder, req.Filter.ID)
		}
	}

	// Update metadata timestamp
	config.Metadata.Updated = time.Now().Format(time.RFC3339)

	// Save the updated configuration
	if err := SaveFiltersConfig(config); err != nil {
		log.Printf("Error saving filters configuration: %v", err)
		http.Error(w, "Failed to save filter", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"filter": req.Filter,
	})
}

// ReorderFiltersRequest represents the request body for reordering filters
type ReorderFiltersRequest struct {
	DisplayOrder []string `json:"displayOrder"`
}

// Handler for PUT /api/filters/reorder
func handleFiltersReorder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req ReorderFiltersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that displayOrder is provided
	if req.DisplayOrder == nil {
		http.Error(w, "displayOrder is required", http.StatusBadRequest)
		return
	}

	// Load current filters config
	config, err := LoadFiltersConfig()
	if err != nil {
		http.Error(w, "Failed to load filters configuration", http.StatusInternalServerError)
		return
	}

	// Validate that all filter IDs in displayOrder exist
	filterMap := make(map[string]bool)
	for _, filter := range config.Filters {
		filterMap[filter.ID] = true
	}

	for _, filterID := range req.DisplayOrder {
		if !filterMap[filterID] {
			http.Error(w, fmt.Sprintf("Filter ID %s does not exist", filterID), http.StatusBadRequest)
			return
		}
	}

	// Update display order
	config.DisplayOrder = req.DisplayOrder
	config.Metadata.Updated = time.Now().Format(time.RFC3339)

	// Save the updated configuration
	if err := SaveFiltersConfig(config); err != nil {
		log.Printf("Error saving filters configuration: %v", err)
		http.Error(w, "Failed to reorder filters", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"displayOrder": req.DisplayOrder,
	})
}

// Handler for /api/filters/{id} - handles DELETE requests
func handleFiltersWithID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filter ID from URL
	filterID := strings.TrimPrefix(r.URL.Path, "/api/filters/")
	if filterID == "" {
		http.Error(w, "Filter ID is required", http.StatusBadRequest)
		return
	}

	// Load current filters config
	config, err := LoadFiltersConfig()
	if err != nil {
		http.Error(w, "Failed to load filters configuration", http.StatusInternalServerError)
		return
	}

	// Find and remove the filter
	filterFound := false
	for i, filter := range config.Filters {
		if filter.ID == filterID {
			// Remove filter from slice
			config.Filters = append(config.Filters[:i], config.Filters[i+1:]...)
			filterFound = true
			break
		}
	}

	if !filterFound {
		http.Error(w, "Filter not found", http.StatusNotFound)
		return
	}

	// Remove from display order
	newDisplayOrder := make([]string, 0, len(config.DisplayOrder))
	for _, id := range config.DisplayOrder {
		if id != filterID {
			newDisplayOrder = append(newDisplayOrder, id)
		}
	}
	config.DisplayOrder = newDisplayOrder

	// Update metadata timestamp
	config.Metadata.Updated = time.Now().Format(time.RFC3339)

	// Save the updated configuration
	if err := SaveFiltersConfig(config); err != nil {
		log.Printf("Error saving filters configuration: %v", err)
		http.Error(w, "Failed to delete filter", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"deletedID": filterID,
	})
}

// SaveFiltersConfig saves the filters configuration to filters.yaml
func SaveFiltersConfig(config *FiltersConfig) error {
	filtersPath := filepath.Join(configDir, "filters.yaml")
	
	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal filters config to YAML: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filtersPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write filters file: %v", err)
	}

	return nil
}
