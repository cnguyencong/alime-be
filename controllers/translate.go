package controllers

import (
	"alime-be/services"
	"alime-be/types"

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

	// Call the service to translate
	result, err := services.TranslateSegments(segments, req.TargetLanguage)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, result)
}
