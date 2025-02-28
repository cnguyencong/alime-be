package services

import (
	"alime-be/types"
	"log"
	"time"

	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func ProcessTranscriptionScript(filename string, processID string) (map[string]interface{}, error) {
	model := "medium"

	outputDir := filepath.Join(".", "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	scriptPath := filepath.Join(".", "scripts/faster-whisper.py")

	args := []string{
		scriptPath,
		filename,
		"--model", model,
		"--process-id", processID,
		"--output-path", outputDir,
		"--output-name", filename,
	}

	cmd := exec.Command("python", args...)

	log.Printf("Executing command: python %s", strings.Join(args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper process failed: %v\nError output: %s", err, string(output))
	}

	baseFileName := filepath.Base(filename)
	ext := filepath.Ext(baseFileName)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, ext)+".json")

	// Read the output file
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %v", err)
	}

	// Parse the JSON response
	var whisperResp types.WhisperResponse
	if err := json.Unmarshal(outputContent, &whisperResp); err != nil {
		return nil, fmt.Errorf("failed to parse whisper output: %v", err)
	}

	return map[string]interface{}{
		"success":   true,
		"processID": processID,
		"segments":  whisperResp.Segments,
	}, nil
}

func MergeSubtitleToVideo(filename string, segments []types.Segment) (string, error) {
	outputDir := filepath.Dir(filename)
	baseFileName := filepath.Base(filename)

	// Generate SRT file
	err := generateSRTFile(segments, outputDir, filename)
	if err != nil {
		return "", fmt.Errorf("failed to generate SRT file: %v", err)
	}

	videoPath := filename
	srtPath := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))+".srt")
	outputVideoPath := filepath.Join(".", "output_subtitled", "subtitled_"+string(uuid.New().String())+"_"+baseFileName)

	// Ensure paths are absolute and use forward slashes
	videoPath = filepath.ToSlash(filepath.Clean(videoPath))
	srtPath = filepath.ToSlash(filepath.Clean(srtPath))
	outputVideoPath = filepath.ToSlash(filepath.Clean(outputVideoPath))

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputVideoPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output subtitled directory: %v", err)
	}

	// Prepare FFmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "subtitles="+srtPath,
		"-c:a", "copy",
		outputVideoPath)

	// Use CombinedOutput with Wait to ensure command fully completes
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to merge subtitles: %v. Output: %s", err, string(output))
	}

	// Verify file was created
	if _, err := os.Stat(outputVideoPath); os.IsNotExist(err) {
		return "", fmt.Errorf("output video file was not created")
	}

	// Get relative path from current working directory
	relPath, err := filepath.Rel(".", outputVideoPath)
	if err != nil {
		relPath = outputVideoPath
	}

	// return map[string]interface{}{
	// 	"file_path": relPath,
	// }, nil

	return relPath, nil
}

func generateSRTFile(segments []types.Segment, outputDir, baseFileName string) error {
	// Create SRT filename
	srtFileName := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(baseFileName), filepath.Ext(baseFileName))+".srt")

	// Check if the file already exists
	if _, err := os.Stat(srtFileName); err == nil {
		// If it does, remove it
		if err := os.Remove(srtFileName); err != nil {
			return fmt.Errorf("failed to remove existing SRT file: %v", err)
		}
	}

	// Create or truncate the SRT file
	file, err := os.Create(srtFileName)
	if err != nil {
		return fmt.Errorf("failed to create SRT file: %v", err)
	}
	defer file.Close()

	// Format time for SRT (HH:MM:SS,mmm)
	formatTime := func(seconds float64) string {
		duration := time.Duration(seconds * float64(time.Second))
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		secs := int(duration.Seconds()) % 60
		milliseconds := int(duration.Milliseconds()) % 1000

		return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
	}

	// Write segments to file
	for i, segment := range segments {
		// Write segment number
		_, err := fmt.Fprintf(file, "%d\n", i+1)
		if err != nil {
			return fmt.Errorf("failed to write segment number: %v", err)
		}

		// Write timestamp
		start := segment.Start
		end := segment.End
		_, err = fmt.Fprintf(file, "%s --> %s\n", formatTime(start), formatTime(end))
		if err != nil {
			return fmt.Errorf("failed to write timestamp: %v", err)
		}

		// Write text and blank line
		_, err = fmt.Fprintf(file, "%s\n\n", segment.Text)
		if err != nil {
			return fmt.Errorf("failed to write text: %v", err)
		}
	}

	return nil
}
