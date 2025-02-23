package controllers

import (
	"alime-be/types"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func HandleTTS(c *gin.Context) {
	req := types.TTSRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convert []map[string]interface{} to []types.Segment
	segments := make([]types.TTSSegment, len(req.TTSSegments))
	for i, segment := range req.TTSSegments {
		segments[i] = types.TTSSegment{
			ID:       int(segment["id"].(float64)),
			Start:    segment["start"].(float64),
			End:      segment["end"].(float64),
			Text:     segment["text"].(string),
			Language: segment["language"].(string),
		}
	}

	processId := req.ProcessID
	log.Print(processId)
	log.Println(segments)

	// Call the service to translate
	result, err := ProcessTTS(segments, processId)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func ProcessTTS(segments []types.TTSSegment, id string) (map[string]interface{}, error) {
	segmentsFilePath, err := saveTemporarySegmentFile(segments, id)
	if err != nil {
		return nil, err
	}
	return processTTSSCrip(segmentsFilePath)
}

func processTTSSCrip(segmentsFilePath string) (map[string]interface{}, error) {
	return nil, nil
}

func saveTemporarySegmentFile(segments []types.TTSSegment, fileName string) (string, error) {
	// save segments to json file
	segmentsJSON := struct {
		Segments []types.TTSSegment `json:"segments"`
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

	segmentsFilePath := fmt.Sprintf("temporary-data/tts_segments_%s.json", fileName)
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
