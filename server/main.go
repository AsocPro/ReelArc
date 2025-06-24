package main

import (
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
)

// TimelineItem represents a single item in the timeline
type TimelineItem struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Start     string `json:"start"`
	End       string `json:"end,omitempty"`
	Type      string `json:"type,omitempty"`
	MediaPath string `json:"mediaPath,omitempty"`
}

// MediaMetadata represents metadata for a media file
type MediaMetadata struct {
	ID            string            `json:"id"`
	Filename      string            `json:"filename"`
	Path          string            `json:"path"`
	Type          string            `json:"type"`
	Timestamp     string            `json:"timestamp"`
	Duration      float64           `json:"duration,omitempty"`
	Transcription string            `json:"transcription"`
	Labels        []string          `json:"labels"`
	Transcripts   []TranscriptEntry `json:"transcripts,omitempty"`
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
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Text     string  `json:"text"`
	Segment  int     `json:"segment"`
	Speaker  string  `json:"speaker,omitempty"`
	Metadata string  `json:"metadata,omitempty"`
}

const (
	dataDir      = "./data"
	mediaDir     = "./data/media"
	metadataDir  = "./data/metadata"
	timelineFile = "./data/timeline.json"
	clientDir    = "./client/dist"
)

func main() {
	// Ensure data directories exist
	ensureDirectories()

	// Create sample timeline data if it doesn't exist
	ensureSampleTimelineData()

	// API routes
	http.HandleFunc("/api/timeline", handleTimeline)
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/metadata/", handleMetadata)
	http.HandleFunc("/api/media", handleMedia)

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
	dirs := []string{dataDir, mediaDir, metadataDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func ensureSampleTimelineData() {
	if _, err := os.Stat(timelineFile); os.IsNotExist(err) {
		// Create sample timeline data
		items := []TimelineItem{
			{
				ID:      "1",
				Content: "Sample Audio",
				Start:   "2023-01-01",
				Type:    "audio",
			},
			{
				ID:      "2",
				Content: "Sample Video",
				Start:   "2023-01-02",
				End:     "2023-01-03",
				Type:    "video",
			},
			{
				ID:      "3",
				Content: "Sample Image",
				Start:   "2023-01-04",
				Type:    "image",
			},
		}

		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal sample timeline data: %v", err)
		}

		if err := os.WriteFile(timelineFile, data, 0644); err != nil {
			log.Fatalf("Failed to write sample timeline data: %v", err)
		}

		fmt.Println("Created sample timeline data")
	}
}

func handleTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := os.ReadFile(timelineFile)
	if err != nil {
		http.Error(w, "Failed to read timeline data", http.StatusInternalServerError)
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

	// Parse multipart form, 10 MB max
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create file path
	filename := handler.Filename
	filePath := filepath.Join(mediaDir, filename)

	// Create a temporary buffer to store the file content
	// We need this to read EXIF data and then save the file
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := dst.Write(fileBytes); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create metadata
	mediaType := "unknown"
	if strings.HasSuffix(strings.ToLower(filename), ".mp3") || strings.HasSuffix(strings.ToLower(filename), ".wav") {
		mediaType = "audio"
	} else if strings.HasSuffix(strings.ToLower(filename), ".mp4") || strings.HasSuffix(strings.ToLower(filename), ".mov") {
		mediaType = "video"
	} else if strings.HasSuffix(strings.ToLower(filename), ".jpg") || strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
		mediaType = "photo"
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

	// Save metadata
	metadataPath := filepath.Join(metadataDir, filename+".json")
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		http.Error(w, "Failed to create metadata", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		http.Error(w, "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"filename": filename,
		"path":     "/media/" + filename,
		"metadata": "/api/metadata/" + filename,
	})
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/api/metadata/")
	
	// If no filename is provided, return all metadata files
	if filename == "" {
		// Read all JSON files from the metadata directory
		files, err := os.ReadDir(metadataDir)
		if err != nil {
			http.Error(w, "Failed to read metadata directory", http.StatusInternalServerError)
			return
		}

		var allMetadata []MediaMetadata
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
				filePath := filepath.Join(metadataDir, file.Name())
				data, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Failed to read metadata file %s: %v", file.Name(), err)
					continue
				}

				var metadata MediaMetadata
				if err := json.Unmarshal(data, &metadata); err != nil {
					log.Printf("Failed to unmarshal metadata file %s: %v", file.Name(), err)
					continue
				}

				allMetadata = append(allMetadata, metadata)
			}
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
	metadataPath := filepath.Join(metadataDir, filename+".json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Metadata not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to read metadata", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
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

	// Read all JSON files from the metadata directory
	files, err := os.ReadDir(metadataDir)
	if err != nil {
		http.Error(w, "Failed to read metadata directory", http.StatusInternalServerError)
		return
	}

	var allMetadata []MediaMetadata
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(metadataDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Failed to read metadata file %s: %v", file.Name(), err)
				continue
			}

			var metadata MediaMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				log.Printf("Failed to unmarshal metadata file %s: %v", file.Name(), err)
				continue
			}

			allMetadata = append(allMetadata, metadata)
		}
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
