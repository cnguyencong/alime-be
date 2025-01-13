package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Add new struct for request parameters
type TranscriptionRequest struct {
	Model        string   `form:"model,default=base"`
	Language     string   `form:"language,default=en"`
	OutputFormat []string `form:"output_formats[]"`
}

func HandleFileUpload(c *gin.Context) {
	// Parse form parameters
	var req TranscriptionRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Set defaults if not provided
	if len(req.OutputFormat) == 0 {
		req.OutputFormat = []string{"srt"}
	}
	if req.Model == "" {
		req.Model = "base"
	}

	// Get the file from the request
	log.Println(c.Request.Body)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"error": "No file uploaded",
		})
		return
	}
	// Validate file type (example for audio/video)
	if !isValidFileType(file.Header.Get("Content-Type")) {
		c.JSON(400, gin.H{
			"error": "Invalid file type. Only audio/video files are allowed",
		})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := filepath.Join(uploadDir, file.Filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// Determine if file is video based on content type
	isVideo := strings.HasPrefix(file.Header.Get("Content-Type"), "video/")

	// Process the file with Python script with additional parameters
	result, err := processPythonScript(filename, req.Model, req.Language, req.OutputFormat, isVideo)
	log.Println(req)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	// Return the result
	c.JSON(200, result)
}

func isValidFileType(contentType string) bool {
	validTypes := []string{
		"video/mp4",
		"video/mpeg",
		"audio/mpeg",
		"audio/wav",
		// Add more types as needed
	}

	for _, t := range validTypes {
		if t == contentType {
			return true
		}
	}
	return false
}

func processPythonScript(filename string, model string, language string, outputFormats []string, isVideo bool) (map[string]interface{}, error) {
	// Set default values if empty
	if model == "" {
		model = "base"
	}
	if language == "" {
		language = "en" // default to English
	}
	if len(outputFormats) == 0 {
		outputFormats = []string{"srt"}
	}

	// Create output directory
	outputDir := filepath.Join(".", "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Construct whisper command arguments
	scriptPath := filepath.Join(".", "/controllers/whisper_transcribe.py")
	args := []string{
		scriptPath,
		filename,
		"--model", model,
		"--language", language,
		"--output-dir", outputDir,
		"--output-formats", "json",
	}

	if isVideo {
		args = append(args, "--is-video")
	}

	// Create and configure command
	cmd := exec.Command("python", args...)

	// Detach from stdin and set up output pipes
	cmd.Stdin = nil
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start command
	log.Printf("Executing command: python %s", strings.Join(args, " "))
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %v", err)
	}

	// Handle output in real-time
	var output, errorOutput strings.Builder
	go io.Copy(io.MultiWriter(&output, os.Stdout), stdout)
	go io.Copy(io.MultiWriter(&errorOutput, os.Stderr), stderr)

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("whisper process failed: %v\nError output: %s", err, errorOutput.String())
	}

	// Clean up the media file
	if err := os.Remove(filename); err != nil {
		log.Printf("Failed to cleanup media file: %v", err)
	}

	// Generate output file path
	baseFileName := filepath.Base(filename)
	ext := filepath.Ext(baseFileName)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, ext)+".json")

	type Segment struct {
		ID    int     `json:"id"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Text  string  `json:"text"`
	}

	type WhisperResponse struct {
		Segments []Segment `json:"segments"`
	}

	// Read the output file
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %v", err)
	}

	// Parse the JSON response
	var whisperResp WhisperResponse
	if err := json.Unmarshal(outputContent, &whisperResp); err != nil {
		return nil, fmt.Errorf("failed to parse whisper output: %v", err)
	}

	// Create simplified segments
	simplifiedSegments := make([]map[string]interface{}, len(whisperResp.Segments))
	for i, segment := range whisperResp.Segments {
		simplifiedSegments[i] = map[string]interface{}{
			"id":    segment.ID,
			"start": segment.Start,
			"end":   segment.End,
			"text":  segment.Text,
		}
	}

	return map[string]interface{}{
		"success":  true,
		"segments": simplifiedSegments,
	}, nil
}
