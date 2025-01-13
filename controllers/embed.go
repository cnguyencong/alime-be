package controllers

import (
	"embed"
	"os"
)

//go:embed whisper_transcribe.py
var embeddedFiles embed.FS

// ExtractEmbeddedFiles extracts embedded files to the working directory
func ExtractEmbeddedFiles() error {
	pythonScript, err := embeddedFiles.ReadFile("whisper_transcribe.py")
	if err != nil {
		return err
	}

	// Write the Python script to the working directory
	err = os.WriteFile("whisper_transcribe.py", pythonScript, 0644)
	if err != nil {
		return err
	}

	return nil
}
