package services

import (
	"alime-be/types"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func TranslateSegments(segments []types.Segment, lang string) (map[string]interface{}, error) {
	log.Printf("lang: %s", lang)

	langCode := loadLanguageCode(lang)
	if langCode == "" {
		return nil, fmt.Errorf("can't find target language")
	}

	segmentsFilePath, err := saveTemporarySegmentFile(segments, langCode)
	if err != nil {
		return nil, fmt.Errorf("can't save segments to file: %v", err)
	}

	return processTranslateScript(segmentsFilePath, langCode)
}

func saveTemporarySegmentFile(segments []types.Segment, langCode string) (string, error) {
	// save segments to json file
	segmentsJSON := struct {
		Segments []types.Segment `json:"segments"`
	}{
		Segments: segments,
	}

	jsonData, err := json.Marshal(segmentsJSON)
	if err != nil {
		return "", fmt.Errorf("can't convert segments to json: %v", err)
	}

	err = os.MkdirAll("temporary-data", os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("can't create folder: %v", err)
	}

	segmentsFilePath := fmt.Sprintf("temporary-data/segments_%s.json", langCode)
	segmentsFile, err := os.Create(segmentsFilePath)
	if err != nil {
		return "", fmt.Errorf("can't create json file: %v", err)
	}
	defer segmentsFile.Close()

	_, err = segmentsFile.Write(jsonData)
	if err != nil {
		return "", fmt.Errorf("can't write segments to json file: %v", err)
	}

	return segmentsFilePath, nil
}

func processTranslateScript(filename string, langCode string) (map[string]interface{}, error) {

	outputDir := filepath.Join(".", "translated_output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	scriptPath := filepath.Join(".", "scripts/translate.py")

	args := []string{
		scriptPath,
		filename,
		"--target-language", langCode,
		"--output-dir", outputDir,
	}

	cmd := exec.Command("python", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("translate process failed: %v\nError output: %s", err, string(output))
	}

	// cmd := exec.Command("python", args...)

	// // Detach from stdin and set up output pipes
	// cmd.Stdin = nil
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	// }
	// stderr, err := cmd.StderrPipe()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	// }

	// // Start command
	// log.Printf("Executing command: python %s", strings.Join(args, " "))
	// if err := cmd.Start(); err != nil {
	// 	return nil, fmt.Errorf("failed to start command: %v", err)
	// }

	// // Handle output in real-time
	// var output, errorOutput strings.Builder
	// go io.Copy(io.MultiWriter(&output, os.Stdout), stdout)
	// go io.Copy(io.MultiWriter(&errorOutput, os.Stderr), stderr)

	// // Wait for completion
	// if err := cmd.Wait(); err != nil {
	// 	return nil, fmt.Errorf("whisper process failed: %v\nError output: %s", err, errorOutput.String())
	// }

	baseFileName := filepath.Base(filename)
	ext := filepath.Ext(baseFileName)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, ext)+"_translated.json")

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
		"success":  true,
		"segments": whisperResp.Segments,
	}, nil
}

func loadLanguageCode(lang string) string {
	jsonFile, err := os.Open(filepath.Join(".", "/services/translating_code.json"))
	if err != nil {
		log.Printf("can't open json file: %v", err)
		return ""
	}
	defer jsonFile.Close()

	var languages map[string]string
	jsonParser := json.NewDecoder(jsonFile)
	jsonParser.Decode(&languages)

	if lang, ok := languages[lang]; ok {
		return lang
	}

	return ""
}

func TrimVideo(videoPath string, trimStart float64, trimEnd float64) (string, error) {
	// Generate a unique output filename
	outputPath := fmt.Sprintf("%s_trimmed_%d.mp4", strings.TrimSuffix(videoPath, filepath.Ext(videoPath)), time.Now().UnixNano())

	// Construct FFmpeg command to trim video
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", fmt.Sprintf("%.2f", trimStart),
		"-to", fmt.Sprintf("%.2f", trimEnd),
		"-c", "copy",
		outputPath)

	// Run the command and capture any potential errors
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to trim video: %v", err)
	}

	return outputPath, nil
}

// func execScriptWithDebug(args ...string) {
// 	cmd := exec.Command("python", args...)

// 	// Detach from stdin and set up output pipes
// 	cmd.Stdin = nil
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		log.Fatalf("failed to create stdout pipe: %v", err)
// 	}
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		log.Fatalf("failed to create stderr pipe: %v", err)
// 	}

// 	// Start command
// 	log.Printf("Executing command: python %s", strings.Join(args, " "))
// 	if err := cmd.Start(); err != nil {
// 		log.Fatalf("failed to start command: %v", err)
// 	}

// 	// Handle output in real-time
// 	var output, errorOutput strings.Builder
// 	go io.Copy(io.MultiWriter(&output, os.Stdout), stdout)
// 	go io.Copy(io.MultiWriter(&errorOutput, os.Stderr), stderr)

// 	// Wait for completion
// 	if err := cmd.Wait(); err != nil {
// 		log.Fatalf("whisper process failed: %v\nError output: %s", err, errorOutput.String())
// 	}

// 	return output.String(), nil

// }
