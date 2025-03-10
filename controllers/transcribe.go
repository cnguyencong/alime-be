package controllers

import (
	"alime-be/db"
	"alime-be/services"
	"alime-be/types"
	"alime-be/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HandleGenerateTranscribe(c *gin.Context) {
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

	utils.CleanFileWithTime(uploadDir, 8*time.Hour)

	//generate unique file name
	processId := uuid.New().String()
	fileExt := filepath.Ext(file.Filename)
	fileName := strings.TrimSuffix(file.Filename, fileExt)
	fileUniqueName := processId + fileExt
	filePath := filepath.Join(uploadDir, fileUniqueName)

	data := types.MediaStorageData{
		FileName:       fileName,
		FileExt:        fileExt,
		FileFullName:   fileName + fileExt,
		FileUniqueName: fileUniqueName,
		FilePath:       filePath,
	}

	// log.Printf("%v", data)

	err = db.SetItem(string(processId), data)

	// Save the file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	transcriptPath, err := services.ProcessTranscriptionScript(filePath, fileName)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	//Read the output file
	outputContent, err := os.ReadFile(transcriptPath)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Errorf("failed to read output file: %v", err).Error(),
		})
		return
	}

	var result types.WhisperResponse
	if err := json.Unmarshal(outputContent, &result); err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Errorf("failed to parse output: %v", err).Error(),
		})
		return
	}

	// Return the result
	c.JSON(200, gin.H{
		"success":   true,
		"processId": processId,
		"segments":  result.Segments,
	})
}
