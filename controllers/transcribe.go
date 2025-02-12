package controllers

import (
	"alime-be/services"
	"alime-be/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleFileUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"error": "No file uploaded",
		})
		return
	}

	if !utils.IsMediaFileType(file.Header.Get("Content-Type")) {
		c.JSON(400, gin.H{
			"error": "Invalid file type",
		})
		return
	}

	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to create upload directory",
		})
		return
	}

	utils.CleanFileWithTime(uploadDir, 1*time.Hour)

	// Generate unique filename
	filename := filepath.Join(uploadDir, file.Filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	result, err := services.ProcessTranscriptionScript(filename)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	// Return the result
	c.JSON(200, result)
}
