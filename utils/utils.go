package utils

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
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

// func generateSRTFile(segments []map[string]interface{}, outputDir, baseFileName string) error {
// 	// Create SRT filename
// 	srtFileName := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))+".srt")

// 	// Create or truncate the SRT file
// 	file, err := os.Create(srtFileName)
// 	if err != nil {
// 		return fmt.Errorf("failed to create SRT file: %v", err)
// 	}
// 	defer file.Close()

// 	// Format time for SRT (HH:MM:SS,mmm)
// 	formatTime := func(seconds float64) string {
// 		duration := time.Duration(seconds * float64(time.Second))
// 		hours := int(duration.Hours())
// 		minutes := int(duration.Minutes()) % 60
// 		secs := int(duration.Seconds()) % 60
// 		milliseconds := int(duration.Milliseconds()) % 1000

// 		return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
// 	}

// 	// Write segments to file
// 	for i, segment := range segments {
// 		// Write segment number
// 		_, err := fmt.Fprintf(file, "%d\n", i+1)
// 		if err != nil {
// 			return fmt.Errorf("failed to write segment number: %v", err)
// 		}

// 		// Write timestamp
// 		start := segment["start"].(float64)
// 		end := segment["end"].(float64)
// 		_, err = fmt.Fprintf(file, "%s --> %s\n", formatTime(start), formatTime(end))
// 		if err != nil {
// 			return fmt.Errorf("failed to write timestamp: %v", err)
// 		}

// 		// Write text and blank line
// 		_, err = fmt.Fprintf(file, "%s\n\n", segment["text"].(string))
// 		if err != nil {
// 			return fmt.Errorf("failed to write text: %v", err)
// 		}
// 	}

// 	return nil
// }
