package controllers

import (
	"alime-be/db"
	"alime-be/services"
	"alime-be/types"
	"alime-be/utils"
	"fmt"
	"log"
	"net/http"
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

	utils.CleanFileWithTime(uploadDir, 1*time.Hour)

	// Generate unique filename
	filename := filepath.Join(uploadDir, file.Filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	processId := uuid.New().String()
	data := map[string]interface{}{
		"filename": filename,
	}
	err = db.SetItem(string(processId), data)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	result, err := services.ProcessTranscriptionScript(filename, processId)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	// Return the result
	c.JSON(200, result)
}

func HandleDownloadVideoWithCaption(c *gin.Context) {
	req := types.ExportVideoRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convert []map[string]interface{} to []types.Segment
	segments := make([]types.Segment, len(req.Segments))
	for i, segment := range req.Segments {
		segments[i] = types.Segment{
			ID:    int(segment["id"].(float64)),
			Start: segment["start"].(float64),
			End:   segment["end"].(float64),
			Text:  segment["text"].(string),
		}
	}

	var data map[string]interface{}
	err := db.GetItem(req.ProcessId, &data)
	if err != nil {
		log.Fatal(err)
	}
	filename, ok := data["filename"].(string)
	if !ok {
		c.JSON(400, gin.H{
			"error": "Invalid processId",
		})
		return
	}

	result, err := services.MergeSubtitleToVideo(filename, segments)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}
	c.JSON(200, result)
}

func DownloadSubtitledVideo(c *gin.Context) {
	// Get the file path from query parameter
	filePath := c.Query("file")

	// Validate file path
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file specified"})
		return
	}

	// Sanitize and validate the file path
	// Ensure the file is within the output_subtitled directory
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, "output_subtitled") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Construct full file path
	fullPath := filepath.Join(".", cleanPath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Set appropriate headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(fullPath))
	c.Header("Content-Type", "application/octet-stream")

	// Serve the file
	c.File(fullPath)
}
