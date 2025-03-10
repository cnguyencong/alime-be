package utils

import (
	"alime-be/types"
	"os/exec"

	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func IsMediaFileType(fileType string) bool {
	validTypes := []string{
		"video/mp4",
		"video/mpeg",
		"audio/mpeg",
		"audio/wav",
		// Add more types as needed
	}

	for _, t := range validTypes {
		if t == fileType {
			return true
		}
	}
	return false
}

func CleanDir(directory string) {
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Printf("Failed to read directory for cleanup: %v", err)
		return
	}
	for _, file := range files {
		filePath := filepath.Join(directory, file.Name())
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to cleanup old file %s: %v", filePath, err)
		} else {
			log.Printf("Cleaned up old file: %s", filePath)
		}
	}
}

func CleanFileWithTime(directory string, maxAge time.Duration) {
	// Read directory contents
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Printf("Failed to read directory for cleanup: %v", err)
		return
	}

	// Get current time
	now := time.Now()

	// Check each file
	for _, file := range files {
		filePath := filepath.Join(directory, file.Name())

		// Get file info
		info, err := file.Info()
		if err != nil {
			log.Printf("Failed to get file info for %s: %v", filePath, err)
			continue
		}

		// Check if file is older than maxAge
		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to cleanup old file %s: %v", filePath, err)
			} else {
				log.Printf("Cleaned up old file: %s", filePath)
			}
		}
	}
}

// UserSessionInfo ...
type UserSessionInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// JSONRaw ...
type JSONRaw json.RawMessage

// Value ...
func (j JSONRaw) Value() (driver.Value, error) {
	byteArr := []byte(j)
	return driver.Value(byteArr), nil
}

// Scan ...
func (j *JSONRaw) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return error(errors.New("Scan source was not []bytes"))
	}
	err := json.Unmarshal(asBytes, &j)
	if err != nil {
		return error(errors.New("Scan could not unmarshal to []string"))
	}
	return nil
}

// MarshalJSON ...
func (j *JSONRaw) MarshalJSON() ([]byte, error) {
	return *j, nil
}

// UnmarshalJSON ...
func (j *JSONRaw) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// DataList ....
type DataList struct {
	Data JSONRaw `db:"data" json:"data"`
	Meta JSONRaw `db:"meta" json:"meta"`
}

func GenerateSRTFile(segments []types.Segment, outputDir string, mediaData types.MediaStorageData) (string, error) {
	// Create SRT filename
	srtFileOutput := filepath.Join(outputDir, fmt.Sprintf("%s%s", strings.TrimSuffix(mediaData.FileUniqueName, mediaData.FileExt), ".srt"))
	srtFileOutput = filepath.ToSlash(filepath.Clean(srtFileOutput))

	// Check if the file already exists
	if _, err := os.Stat(srtFileOutput); err == nil {
		// If it does, remove it
		if err := os.Remove(srtFileOutput); err != nil {
			return "", fmt.Errorf("failed to remove existing SRT file: %v", err)
		}
	}

	// Check if the output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// Create the directory
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// Create or truncate the SRT file
	file, err := os.Create(srtFileOutput)
	if err != nil {
		return "", fmt.Errorf("failed to create SRT file: %v", err)
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
			return "", fmt.Errorf("failed to write segment number: %v", err)
		}

		// Write timestamp
		start := segment.Start
		end := segment.End
		_, err = fmt.Fprintf(file, "%s --> %s\n", formatTime(start), formatTime(end))
		if err != nil {
			return "", fmt.Errorf("failed to write timestamp: %v", err)
		}

		// Write text and blank line
		_, err = fmt.Fprintf(file, "%s\n\n", segment.Text)
		if err != nil {
			return "", fmt.Errorf("failed to write text: %v", err)
		}
	}

	return srtFileOutput, nil
}

func CreateJSONFile(data interface{}, filename string, outputPath string) (string, error) {
	// Ensure the output directory exists
	outputDir := filepath.Join(".", outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate full path
	path := filepath.Join(outputDir, filename)

	// Marshal JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Write file with UTF-8 encoding
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create JSON file %s: %v", filename, err)
	}
	defer file.Close()

	// Write the UTF-8 BOM (Byte Order Mark) for UTF-8 encoding
	if _, err := file.WriteString("\xEF\xBB\xBF"); err != nil {
		return "", fmt.Errorf("failed to write BOM to JSON file %s: %v", filename, err)
	}

	if _, err := file.Write(jsonData); err != nil {
		return "", fmt.Errorf("failed to write JSON data to file %s: %v", filename, err)
	}

	return path, nil
}

func ExecExternalScript(args []string, cmdName string) ([]byte, error) {
	log.Printf("Executing command: %s %s", cmdName, strings.Join(args, " "))
	cmd := exec.Command(cmdName, args...)
	output, err := cmd.CombinedOutput()

	// Check for errors
	if err != nil {
		// Save the error output to a log file
		errorLogPath := "error/scripts_error_log.txt"
		logEntry := fmt.Sprintf("Time: %s\nFailed to execute command: %s %s\n%s\n--------------------------------------------\n", time.Now().Format(time.RFC3339), cmdName, strings.Join(args, " "), string(output))
		file, err := os.OpenFile(errorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open error log file: %v", err)
			return output, err
		}
		defer file.Close()
		if _, writeErr := file.Write([]byte(logEntry)); writeErr != nil {
			log.Printf("Failed to write error log to file: %v", writeErr)
		}
		return output, err
	}

	return output, nil
}
