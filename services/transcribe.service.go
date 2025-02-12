package services

import (
	"alime-be/types"

	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ProcessTranscriptionScript(filename string) (map[string]interface{}, error) {
	model := "base"

	outputDir := filepath.Join(".", "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	scriptPath := filepath.Join(".", "scripts/faster-whisper.py")

	args := []string{
		scriptPath,
		filename,
		"--model", model,
		"--output-path", outputDir,
		"--output-name", filename,
	}

	cmd := exec.Command("python", args...)
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

	// Generate SRT file if requested
	// if contains(outputFormats, "srt") {
	// 	if err := generateSRTFile(simplifiedSegments, outputDir, baseFileName); err != nil {
	// 		log.Printf("Failed to generate SRT file: %v", err)
	// 		// Continue execution even if SRT generation fails
	// 	}
	// }

	return map[string]interface{}{
		"success":  true,
		"segments": whisperResp.Segments,
	}, nil

}

// func execScriptWithDebug(args ...string) (map[string]interface{}, error) {
// 	cmd := exec.Command("python", args...)

// 	// Detach from stdin and set up output pipes
// 	cmd.Stdin = nil
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
// 	}
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
// 	}

// 	// Start command
// 	log.Printf("Executing command: python %s", strings.Join(args, " "))
// 	if err := cmd.Start(); err != nil {
// 		return nil, fmt.Errorf("failed to start command: %v", err)
// 	}

// 	// Handle output in real-time
// 	var output, errorOutput strings.Builder
// 	go io.Copy(io.MultiWriter(&output, os.Stdout), stdout)
// 	go io.Copy(io.MultiWriter(&errorOutput, os.Stderr), stderr)

// 	// Wait for completion
// 	if err := cmd.Wait(); err != nil {
// 		return nil, fmt.Errorf("whisper process failed: %v\nError output: %s", err, errorOutput.String())
// 	}

// }

// // func generateSRTFile(segments []map[string]interface{}, outputDir, baseFileName string) error {
// // 	// Create SRT filename
// // 	srtFileName := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))+".srt")

// // 	// Create or truncate the SRT file
// // 	file, err := os.Create(srtFileName)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to create SRT file: %v", err)
// // 	}
// // 	defer file.Close()

// // 	// Format time for SRT (HH:MM:SS,mmm)
// // 	formatTime := func(seconds float64) string {
// // 		duration := time.Duration(seconds * float64(time.Second))
// // 		hours := int(duration.Hours())
// // 		minutes := int(duration.Minutes()) % 60
// // 		secs := int(duration.Seconds()) % 60
// // 		milliseconds := int(duration.Milliseconds()) % 1000

// // 		return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
// // 	}

// // 	// Write segments to file
// // 	for i, segment := range segments {
// // 		// Write segment number
// // 		_, err := fmt.Fprintf(file, "%d\n", i+1)
// // 		if err != nil {
// // 			return fmt.Errorf("failed to write segment number: %v", err)
// // 		}

// // 		// Write timestamp
// // 		start := segment["start"].(float64)
// // 		end := segment["end"].(float64)
// // 		_, err = fmt.Fprintf(file, "%s --> %s\n", formatTime(start), formatTime(end))
// // 		if err != nil {
// // 			return fmt.Errorf("failed to write timestamp: %v", err)
// // 		}

// // 		// Write text and blank line
// // 		_, err = fmt.Fprintf(file, "%s\n\n", segment["text"].(string))
// // 		if err != nil {
// // 			return fmt.Errorf("failed to write text: %v", err)
// // 		}
// // 	}

// // 	return nil
// // }
