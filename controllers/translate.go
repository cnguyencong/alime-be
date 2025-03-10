package controllers

import (
	"alime-be/services"
	"alime-be/types"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func HandleTranslate(c *gin.Context) {
	req := types.TranslateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	transcriptPath := filepath.Join(".", "output/transcripts", fmt.Sprintf("%v.json", req.ProcessId))

	// Call the service to translate
	translatedScriptJsonPath, err := services.TranslateSegments(transcriptPath, req.TargetLanguage, req.ProcessId)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	//Read the output file
	outputContent, err := os.ReadFile(translatedScriptJsonPath)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Errorf("failed to read output file: %v", err).Error(),
		})
		return
	}

	var translated types.WhisperResponse
	if err := json.Unmarshal(outputContent, &translated); err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Errorf("failed to parse output: %v", err).Error(),
		})
		return
	}

	//Using the result to generate TTS audio
	tts_path, error := services.BuildTTS(translatedScriptJsonPath, req.TargetLanguage)
	if error != nil {
		c.JSON(500, gin.H{
			"error": error.Error(),
		})
		return
	}

	audioInfoPath := filepath.Join(tts_path, "audio_info.json")
	audioInfoContent, err := os.ReadFile(audioInfoPath)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Errorf("failed to read audio info file: %v", err).Error(),
		})
		return
	}

	var audioInfo []interface{}
	if err := json.Unmarshal(audioInfoContent, &audioInfo); err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Errorf("failed to parse audio info: %v", err).Error(),
		})
		return
	}

	audioInfoMap := make(map[string]map[string]interface{})
	for _, info := range audioInfo {
		if infoMap, ok := info.(map[string]interface{}); ok {
			audioInfoMap[fmt.Sprintf("%v", infoMap["id"])] = infoMap
		}
	}

	var mappedSegments []interface{}
	for _, segment := range translated.Segments {
		audioInfoData, exists := audioInfoMap[fmt.Sprintf("%v", segment.Id)]
		if !exists {
			// Handle the case where the segment ID is not found in audioInfo
			c.JSON(500, gin.H{
				"error": fmt.Sprintf("audio info not found for segment ID: %v", segment.Id),
			})
			return
		}

		mappedSegment := map[string]interface{}{
			"id":          segment.Id,
			"start":       segment.Start,
			"end":         segment.End,
			"text":        segment.Text,
			"audioLength": audioInfoData["audioLength"],
			"audioPath":   audioInfoData["audioPath"],
		}
		mappedSegments = append(mappedSegments, mappedSegment)
	}

	c.JSON(200, gin.H{
		"segments": mappedSegments,
	})
}
