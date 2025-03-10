package controllers

import (
	"alime-be/types"
	"alime-be/utils"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleTTSText(c *gin.Context) {
	req := types.TTSRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	result, err := processTTSText(req.Text, req.Language)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func processTTSText(text string, language string) (map[string]interface{}, error) {
	scriptPath := filepath.Join(".", "scripts/text-to-speech-scripts/tts-input.py")
	name := time.Now().Format("20060102150405")

	args := []string{
		scriptPath,
		text,
		"--name", name,
		"--language", language,
	}

	output, err := utils.ExecExternalScript(args, "python")
	if err != nil {
		return nil, fmt.Errorf("TTS process failed: %v\nError output: %s", err, string(output))
	}

	outputStr := strings.TrimSpace(string(output))
	outputNum, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		return nil, fmt.Errorf("can't convert output to number: %v", err)
	}

	outputFile := name + ".wav"

	return map[string]interface{}{
		"outputFile": filepath.Join("output/tts/temporary-output", outputFile),
		"length":     outputNum,
	}, nil

}
