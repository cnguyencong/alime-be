package controllers

import (
	"alime-be/db"
	"alime-be/services"
	"alime-be/types"
	"alime-be/utils"
	"log"

	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleExportVideo(c *gin.Context) {
	req := types.ExportVideoRequest{}
	log.Printf("HandleExportVideo req: %v", req)

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
			Id:    int(segment["id"].(float64)),
			Start: segment["start"].(float64),
			End:   segment["end"].(float64),
			Text:  segment["text"].(string),
		}
	}

	var mediaData types.MediaStorageData
	err := db.GetItem(req.ProcessId, &mediaData)
	if err != nil {
		log.Fatal(err)
	}
	videoFilePath := mediaData.FilePath

	if req.IsShowCaption {
		srtOutputPath := filepath.Join(".", "output/srt")
		srtOutput, err := utils.GenerateSRTFile(segments, srtOutputPath, mediaData)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to generate SRT file: %v", err)})
			return
		}

		newFile, err := MergeSubtitleToVideo(videoFilePath, mediaData, srtOutput)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
			return
		}
		videoFilePath = newFile
	}

	if req.IsAppendTTS {
		output, error := services.HandleAppendTTS(segments, videoFilePath, req.Language)

		if error != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
			return
		}

		videoFilePath = output
	}

	if req.IsTrimVideo {
		trimedVideoPath, error := TrimVideo(videoFilePath, req.TrimStart, req.TrimEnd)
		if error != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
			return
		}

		videoFilePath = trimedVideoPath
	}

	c.JSON(200, gin.H{
		"file_path": videoFilePath,
	})
}

func MergeSubtitleToVideo(mediaPath string, mediaData types.MediaStorageData, srtPath string) (string, error) {
	outputSubtitlePath := filepath.Join(".", "output/exported", fmt.Sprintf("%s_subtitled%s", mediaData.FileName, mediaData.FileExt))

	// Check if the file already exists
	if _, err := os.Stat(outputSubtitlePath); err == nil {
		// If it does, remove it
		if err := os.Remove(outputSubtitlePath); err != nil {
			return "", fmt.Errorf("failed to remove existing SRT file: %v", err)
		}
	}

	// Ensure paths are absolute and use forward slashes
	mediaPath = filepath.ToSlash(filepath.Clean(mediaPath))
	srtPath = filepath.ToSlash(filepath.Clean(srtPath))
	outputSubtitlePath = filepath.ToSlash(filepath.Clean(outputSubtitlePath))

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputSubtitlePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output subtitled directory: %v", err)
	}

	args := []string{"-i", mediaPath, "-vf", "subtitles=" + srtPath, "-c:a", "copy", outputSubtitlePath}
	output, err := utils.ExecExternalScript(args, "ffmpeg")
	if err != nil {
		return "", fmt.Errorf("failed to merge subtitles: %v. Output: %s", err, string(output))
	}

	// Verify file was created

	if _, err := os.Stat(outputSubtitlePath); os.IsNotExist(err) {
		return "", fmt.Errorf("output video file was not created")
	}

	// Get relative path from current working directory
	relPath, err := filepath.Rel(".", outputSubtitlePath)
	if err != nil {
		relPath = outputSubtitlePath
	}

	return relPath, nil
}

func TrimVideo(videoPath string, trimStart float64, trimEnd float64) (string, error) {
	// Generate a unique output filename
	outputPath := fmt.Sprintf("%s_trimmed%s", strings.TrimSuffix(videoPath, filepath.Ext(videoPath)), filepath.Ext(videoPath))

	// Construct FFmpeg command to trim video
	command := []string{
		"-i", videoPath,
		"-ss", fmt.Sprintf("%.2f", trimStart),
		"-to", fmt.Sprintf("%.2f", trimEnd),
		"-c", "copy", outputPath,
	}

	output, err := utils.ExecExternalScript(command, "ffmpeg")

	// Run the command and capture any potential errors
	if err != nil {
		return "", fmt.Errorf("failed to trim video: %v. Output: %s", err, string(output))
	}

	return outputPath, nil
}

func DownloadVideo(c *gin.Context) {
	req := types.GetMediaRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	path := req.FilePath

	// Validate file path
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file specified"})
		return
	}

	// Sanitize and validate the file path
	cleanPath := filepath.Clean(path)
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

func HandleStreamAudio(c *gin.Context) {
	// Get the name of the audio file from the request parameters or query
	req := types.GetMediaRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	path := req.FilePath // Assuming the filename is passed as a URL parameter

	// Construct the full path to the audio file
	audioFilePath := filepath.Join(".", path)

	// Set the content type to audio/wav
	c.Header("Content-Type", "audio/wav")

	// Serve the audio file
	c.File(audioFilePath)
}
