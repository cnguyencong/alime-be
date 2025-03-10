package services

import (
	"alime-be/utils"

	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ProcessTranscriptionScript(filePath string, fileName string) (string, error) {
	outputDir := filepath.Join(".", "output/transcripts")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	scriptPath := filepath.Join(".", "scripts/transcribe.py")
	baseFileName := filepath.Base(filePath)
	ext := filepath.Ext(baseFileName)

	args := []string{
		scriptPath,
		filePath,
		"--output-path", outputDir,
		"--output-name", baseFileName,
	}
	output, err := utils.ExecExternalScript(args, "python")
	if err != nil {
		return "", fmt.Errorf("whisper process failed: %v\nError output: %s", err, string(output))
	}

	outputFile := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, ext)+".json")
	return string(outputFile), nil

	// //Read the output file with UTF-8 encoding
	// outputContent, err := os.ReadFile(outputFile)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read output file: %v", err)
	// }

	// // Ensure UTF-8 decoding
	// outputContent = []byte(string(outputContent))

	// // Parse the JSON response with UTF-8 support
	// var whisperResp types.WhisperResponse
	// decoder := json.NewDecoder(bytes.NewReader(outputContent))
	// decoder.UseNumber() // Preserve number precision
	// decoder.Decode(&whisperResp)

	// // Optional: Explicit UTF-8 validation
	// for i, segment := range whisperResp.Segments {
	// 	if !utf8.ValidString(segment.Text) {
	// 		// Sanitize or handle invalid UTF-8 characters
	// 		whisperResp.Segments[i].Text = sanitizeUTF8(segment.Text)
	// 	}
	// }

	// return map[string]interface{}{
	// 	"segments": whisperResp.Segments,
	// }, nil
}

// func sanitizeUTF8(s string) string {
// 	return strings.Map(func(r rune) rune {
// 		if r == utf8.RuneError {
// 			return -1
// 		}
// 		return r
// 	}, s)
// }
