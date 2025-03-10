package services

import (
	"alime-be/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func TranslateSegments(transcriptPath string, lang string, id string) (string, error) {
	outputDir := filepath.Join(".", "output/translated", string(id))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	scriptPath := filepath.Join(".", "scripts/translate.py")

	args := []string{
		scriptPath,
		transcriptPath,
		"--target-language", lang,
		"--output-dir", outputDir,
	}

	output, err := utils.ExecExternalScript(args, "python")
	if err != nil {
		return "", fmt.Errorf("translate process failed: %v\nError output: %s", err, string(output))
	}

	baseFileName := filepath.Base(transcriptPath)
	ext := filepath.Ext(baseFileName)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(baseFileName, ext)+fmt.Sprintf("_%s.json", lang))

	return outputFile, nil
}
