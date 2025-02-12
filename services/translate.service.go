package services

import (
	"alime-be/types"
	"log"
)

var languageAlias = map[string]string{
	"en": "",
	"vi": "",
	"fr": "",
}

func TranslateSegments(segments []types.Segment, lang string) (map[string]interface{}, error) {
	log.Printf("lang: %s", lang)
	// log.Printf("lang: %s", translatedSegments)

	return map[string]interface{}{
		"success":  true,
		"segments": "",
	}, nil
}

// func getModelName(lang string) string {
// 	return LanguagesModels[lang]
// }

// func generateSRTFile(segments []map[string]interface{}, outputDir, baseFileName string) error {
// 	// Create SRT filename
// 	srtFileName := filepath.Join(".", strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))+".srt")

// 	// Create or truncate the SRT file
// 	file, err := os.Create(srtFileName)
// 	if err != nil {
// 		return fmt.Errorf("failed to create SRT file: %v", err)
// 	}
// 	defer file.Close()

// 	// Format time for SRT (HH:MM:SS,mmm)
// 	formatTime := func(seconds float64) string {
// 		duration := time.Duration(seconds * float64(time.Second))
// 		hours := int(duration.Hours())
// 		minutes := int(duration.Minutes()) % 60
// 		secs := int(duration.Seconds()) % 60
// 		milliseconds := int(duration.Milliseconds()) % 1000

// 		return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
// 	}

// 	// Write segments to file
// 	for i, segment := range segments {
// 		// Write segment number
// 		_, err := fmt.Fprintf(file, "%d\n", i+1)
// 		if err != nil {
// 			return fmt.Errorf("failed to write segment number: %v", err)
// 		}

// 		// Write timestamp
// 		start := segment["start"].(float64)
// 		end := segment["end"].(float64)
// 		_, err = fmt.Fprintf(file, "%s --> %s\n", formatTime(start), formatTime(end))
// 		if err != nil {
// 			return fmt.Errorf("failed to write timestamp: %v", err)
// 		}

// 		// Write text and blank line
// 		_, err = fmt.Fprintf(file, "%s\n\n", segment["text"].(string))
// 		if err != nil {
// 			return fmt.Errorf("failed to write text: %v", err)
// 		}
// 	}

// 	return nil
// }
