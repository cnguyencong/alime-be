package services

import (
	"alime-be/types"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BuildTTSAudioWithBGM(audioFolderPath string, transcriptsPath string, bgmPath string) (string, error) {
	scriptPath := filepath.Join(".", "scripts/build-audio-with-bgm.py")

	args := []string{
		scriptPath,
		filepath.Join(".", audioFolderPath),
		filepath.Join(".", transcriptsPath),
		filepath.Join(".", bgmPath),
		// filepath.Join(".", videoPath),
	}

	cmd := exec.Command("python", args...)
	output, err := cmd.CombinedOutput()
	println(string(output))
	if err != nil {
		return "", fmt.Errorf("whisper process failed: %v\nError output: %s", err, string(output))
	}

	bgmFolder := filepath.Dir(bgmPath)

	result := filepath.Join(".", bgmFolder, "final_mixed_audio.wav")

	return result, nil

}

func HandleAppendTTS(segments []types.Segment, videoPath string, language string) (string, error) {

	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	jsonPath, error := CreateJSONFile(segments, videoName+".json")
	if error != nil {
		log.Fatal(error)
	}

	bgm, error := GenerateBGMAudio(videoPath)
	if error != nil {
		return "", fmt.Errorf("failed to generate bgm: %v", error)
	}

	log.Println(bgm, "------------------------------------------")

	tts_path, error := BuildTTS(jsonPath, language)
	if error != nil {
		return "", fmt.Errorf("failed to build tts: %v", error)
	}

	log.Println(tts_path, "------------------------------------------")

	mixedAudioPath, error := BuildTTSAudioWithBGM(tts_path, jsonPath, bgm)

	log.Println(mixedAudioPath, "------------------------------------------")

	fullOutputPath, error := ReplaceVideoAudio(videoPath, mixedAudioPath, filepath.Dir(videoPath))

	return fullOutputPath, nil
}

// Replace the audio in a video file with a new audio track.
func ReplaceVideoAudio(videoPath string, audioPath string, outputPath string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate output filename
	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	outputFilename := videoName + "_final.mp4"
	fullOutputPath := filepath.Join(outputPath, outputFilename)

	// Construct FFmpeg command with volume boost
	command := []string{
		"ffmpeg",
		"-i", videoPath,
		"-i", audioPath,
		"-filter_complex", "[1:a]volume=9.0[a_boosted]", // Boost volume by 3x
		"-map", "0:v",
		"-map", "[a_boosted]",
		"-c:v", "copy",
		"-c:a", "aac",
		"-shortest",
		fullOutputPath,
	}

	fmt.Println("Replace Video Audio Command:", strings.Join(command, " "))

	cmd := exec.Command(command[0], command[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to replace audio: %v\nError output: %s", err, string(output))
	}

	// Verify file was created
	if _, err := os.Stat(fullOutputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("output video file was not created")
	}

	return fullOutputPath, nil
}

func CreateJSONFile(data interface{}, filename string) (string, error) {
	// Ensure the output directory exists
	outputDir := filepath.Join(".", "tts-json-output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate full path
	path := filepath.Join(outputDir, filename)

	// Marshal JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Write file with proper permissions
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON file %s: %v", filename, err)
	}

	return path, nil
}

func GenerateBGMAudio(mediaPath string) (string, error) {
	scriptPath := filepath.Join(".", "scripts/split-BGM.py")

	args := []string{
		scriptPath,
		mediaPath,
	}

	cmd := exec.Command("python", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("whisper process failed: %v\nError output: %s", err, string(output))
	}

	audioName := strings.TrimSuffix(filepath.Base(mediaPath), filepath.Ext(mediaPath))

	result := filepath.Join(".", "separated", "htdemucs", audioName+"-audio", "no_vocals.wav")

	fmt.Println("--------------------------------")
	return result, nil
}

func BuildTTS(transcriptsPath string, language string) (string, error) {
	scriptPath := filepath.Join(".", "scripts/tts-file.py")

	args := []string{
		scriptPath,
		transcriptsPath,
		"--language", language,
	}

	cmd := exec.Command("python", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("TTS process failed: %v\nError output: %s", err, string(output))
	}

	name := filepath.Base(transcriptsPath)
	result := filepath.Join(".", "temporary-output", strings.TrimSuffix(name, filepath.Ext(name)))
	fmt.Println("--------------------------------")

	return result, nil
}
