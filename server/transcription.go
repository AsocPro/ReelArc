package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TranscriptionQueue manages the queue of files to be transcribed
type TranscriptionQueue struct {
	Queue     []string
	InProcess map[string]bool
	Completed map[string]bool
	Failed    map[string]string // filename -> error message
	mu        sync.Mutex
}

// TranscriptionStatus represents the status of a transcription job
type TranscriptionStatus struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"` // "queued", "processing", "completed", "failed"
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

const (
	transcriptsDir = "./data/transcripts"
)

var (
	// Global transcription queue
	TQueue = &TranscriptionQueue{
		Queue:     []string{},
		InProcess: make(map[string]bool),
		Completed: make(map[string]bool),
		Failed:    make(map[string]string),
	}
)

// Initialize transcription system
func InitTranscriptionSystem() {
	// Ensure transcripts directory exists
	if err := os.MkdirAll(transcriptsDir, 0755); err != nil {
		log.Fatalf("Failed to create transcripts directory: %v", err)
	}

	// Start the transcription worker
	go transcriptionWorker()

	// Check for existing audio/video files without transcripts
	checkExistingMediaFiles()
}

// Add a file to the transcription queue
func (tq *TranscriptionQueue) AddToQueue(filename string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Check if file is already in queue, in process, completed, or failed
	if tq.isInQueue(filename) || tq.InProcess[filename] || tq.Completed[filename] || tq.Failed[filename] != "" {
		return
	}

	// Add to queue
	tq.Queue = append(tq.Queue, filename)
	log.Printf("Added %s to transcription queue", filename)
}

// Check if a file is in the queue
func (tq *TranscriptionQueue) isInQueue(filename string) bool {
	for _, f := range tq.Queue {
		if f == filename {
			return true
		}
	}
	return false
}

// Get the next file from the queue
func (tq *TranscriptionQueue) GetNext() (string, bool) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if len(tq.Queue) == 0 {
		return "", false
	}

	// Get the first file
	filename := tq.Queue[0]
	
	// Remove from queue
	tq.Queue = tq.Queue[1:]
	
	// Mark as in process
	tq.InProcess[filename] = true
	
	return filename, true
}

// Mark a file as completed
func (tq *TranscriptionQueue) MarkCompleted(filename string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	delete(tq.InProcess, filename)
	tq.Completed[filename] = true
}

// Mark a file as failed with an error message
func (tq *TranscriptionQueue) MarkFailed(filename, errorMsg string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	delete(tq.InProcess, filename)
	tq.Failed[filename] = errorMsg
}

// Get all transcription statuses
func (tq *TranscriptionQueue) GetAllStatuses() []TranscriptionStatus {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	var statuses []TranscriptionStatus

	// Add queued files
	for _, filename := range tq.Queue {
		statuses = append(statuses, TranscriptionStatus{
			Filename:  filename,
			Status:    "queued",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	// Add in-process files
	for filename := range tq.InProcess {
		statuses = append(statuses, TranscriptionStatus{
			Filename:  filename,
			Status:    "processing",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	// Add completed files
	for filename := range tq.Completed {
		statuses = append(statuses, TranscriptionStatus{
			Filename:  filename,
			Status:    "completed",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	// Add failed files
	for filename, errMsg := range tq.Failed {
		statuses = append(statuses, TranscriptionStatus{
			Filename:  filename,
			Status:    "failed",
			Error:     errMsg,
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	return statuses
}

// Worker that processes the transcription queue
func transcriptionWorker() {
	for {
		// Get next file from queue
		filename, ok := TQueue.GetNext()
		if !ok {
			// No files in queue, sleep and try again
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Processing transcription for %s", filename)
		
		// Process the file
		err := processTranscription(filename)
		if err != nil {
			log.Printf("Transcription failed for %s: %v", filename, err)
			TQueue.MarkFailed(filename, err.Error())
			
			// Create a .failed file
			failedFilePath := filepath.Join(transcriptsDir, filename+".failed")
			if err := os.WriteFile(failedFilePath, []byte(err.Error()), 0644); err != nil {
				log.Printf("Failed to write failure file for %s: %v", filename, err)
			}
		} else {
			log.Printf("Transcription completed for %s", filename)
			TQueue.MarkCompleted(filename)
		}
	}
}

// Check for existing media files that don't have transcripts
func checkExistingMediaFiles() {
	files, err := os.ReadDir(mediaDir)
	if err != nil {
		log.Printf("Failed to read media directory: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		lowerFilename := strings.ToLower(filename)
		
		// Check if it's an audio or video file
		if strings.HasSuffix(lowerFilename, ".mp3") || 
		   strings.HasSuffix(lowerFilename, ".wav") || 
		   strings.HasSuffix(lowerFilename, ".mp4") || 
		   strings.HasSuffix(lowerFilename, ".mov") {
			
			// Check if transcript already exists
			transcriptPath := filepath.Join(transcriptsDir, filename+".json")
			failedPath := filepath.Join(transcriptsDir, filename+".failed")
			
			if _, err := os.Stat(transcriptPath); os.IsNotExist(err) {
				if _, err := os.Stat(failedPath); os.IsNotExist(err) {
					// No transcript or failed file exists, add to queue
					TQueue.AddToQueue(filename)
				}
			}
		}
	}
}

// Process a file for transcription
func processTranscription(filename string) error {
	filePath := filepath.Join(mediaDir, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}
	
	// Determine if it's an audio or video file
	lowerFilename := strings.ToLower(filename)
	isVideo := strings.HasSuffix(lowerFilename, ".mp4") || strings.HasSuffix(lowerFilename, ".mov")
	isAudio := strings.HasSuffix(lowerFilename, ".mp3") || strings.HasSuffix(lowerFilename, ".wav")
	
	if !isVideo && !isAudio {
		return fmt.Errorf("unsupported file type: %s", filename)
	}
	
	// For video files, extract audio first
	var audioPath string
	if isVideo {
		// Extract audio using ffmpeg
		audioPath = filepath.Join(transcriptsDir, filename+".wav")
		if err := extractAudioFromVideo(filePath, audioPath); err != nil {
			return fmt.Errorf("failed to extract audio: %v", err)
		}
	} else {
		// For audio files, use the original file
		audioPath = filePath
	}
	
	// Run whisperx on the audio file
	transcriptPath := filepath.Join(transcriptsDir, filename+".json")
	if err := runWhisperX(audioPath, transcriptPath); err != nil {
		return fmt.Errorf("whisperx transcription failed: %v", err)
	}
	
	// Clean up temporary audio file if it was extracted from video
	if isVideo {
		if err := os.Remove(audioPath); err != nil {
			log.Printf("Warning: Failed to remove temporary audio file %s: %v", audioPath, err)
		}
	}
	
	// Update the metadata file with transcript information
	if err := updateMetadataWithTranscript(filename, transcriptPath); err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}
	
	return nil
}

// Extract audio from a video file using ffmpeg
func extractAudioFromVideo(videoPath, audioPath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1", audioPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}
	return nil
}

// Run whisperx on an audio file
func runWhisperX(audioPath, outputPath string) error {
	// Create a temporary directory for whisperx output
	tempDir, err := os.MkdirTemp("", "whisperx")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = os.Chmod(tempDir, 0777)
    if err != nil {
        log.Fatal(err)
	}
	
	audioFileName := filepath.Base(audioPath)
	tempAudioPath := filepath.Join(tempDir, audioFileName)
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return fmt.Errorf("failed to read audio file error: %v", err)
	}

	if err := os.WriteFile(tempAudioPath, audioData, 0666); err != nil {
		return fmt.Errorf("failed to read audio file error: %v", err)
	}

	// Run whisperx
	cmd := exec.Command("podman", "run",  "-v",  tempDir + ":/app:Z", "ghcr.io/jim60105/whisperx:base-en", "--", "--output_format", "json", "--compute_type", "int8", audioFileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("whisperx error: %v, output: %s", err, string(output))
	}
	
	// Find the JSON output file
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read whisperx output directory: %v", err)
	}
	
	var jsonFile string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			jsonFile = filepath.Join(tempDir, file.Name())
			break
		}
	}
	
	if jsonFile == "" {
		return fmt.Errorf("no JSON output found from whisperx")
	}
	
	// Read the whisperx output
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to read whisperx output: %v", err)
	}
	
	// Parse the whisperx output to extract segments
	var whisperOutput map[string]interface{}
	if err := json.Unmarshal(data, &whisperOutput); err != nil {
		return fmt.Errorf("failed to parse whisperx output: %v", err)
	}
	
	// Convert to our transcript format
	segments, ok := whisperOutput["segments"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid whisperx output format")
	}
	
	var transcriptEntries []TranscriptEntry
	for i, seg := range segments {
		segment, ok := seg.(map[string]interface{})
		if !ok {
			continue
		}
		
		start, _ := segment["start"].(float64)
		end, _ := segment["end"].(float64)
		text, _ := segment["text"].(string)
		
		entry := TranscriptEntry{
			Start:   start,
			End:     end,
			Text:    text,
			Segment: i,
		}
		
		transcriptEntries = append(transcriptEntries, entry)
	}
	
	// Write the transcript to the output file
	transcriptData, err := json.MarshalIndent(transcriptEntries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transcript: %v", err)
	}
	
	if err := os.WriteFile(outputPath, transcriptData, 0644); err != nil {
		return fmt.Errorf("failed to write transcript file: %v", err)
	}
	
	return nil
}

// Update metadata file with transcript information
func updateMetadataWithTranscript(filename, transcriptPath string) error {
	// Read the transcript file
	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return fmt.Errorf("failed to read transcript file: %v", err)
	}
	
	var transcriptEntries []TranscriptEntry
	if err := json.Unmarshal(data, &transcriptEntries); err != nil {
		return fmt.Errorf("failed to parse transcript: %v", err)
	}
	
	// Read the metadata file
	metadataPath := filepath.Join(metadataDir, filename+".json")
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata file: %v", err)
	}
	
	var metadata MediaMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %v", err)
	}
	
	// Update metadata with transcript
	metadata.Transcripts = transcriptEntries
	
	// Create a full transcription text
	var fullText strings.Builder
	for _, entry := range transcriptEntries {
		fullText.WriteString(entry.Text)
		fullText.WriteString(" ")
	}
	metadata.Transcription = fullText.String()
	
	// Write updated metadata back to file
	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated metadata: %v", err)
	}
	
	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated metadata: %v", err)
	}
	
	return nil
}
