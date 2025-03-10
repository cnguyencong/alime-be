package services

import (
	"alime-be/types"
	"alime-be/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BuildTTSAudioWithBGM(audioFolderPath string, transcriptsPath string, bgmPath string) (string, error) {
	scriptPath := filepath.Join(".", "scripts/text-to-speech-scripts/build-audio-with-bgm.py")

	args := []string{
		scriptPath,
		filepath.Join(".", audioFolderPath),
		filepath.Join(".", transcriptsPath),
		filepath.Join(".", bgmPath),
	}

	output, err := utils.ExecExternalScript(args, "python")
	if err != nil {
		return "", fmt.Errorf("build tts with bgm failed: %v\nError output: %s", err, string(output))
	}

	bgmFolder := filepath.Dir(bgmPath)

	result := filepath.Join(".", bgmFolder, "final_mixed_audio.wav")

	return result, nil

}

func HandleAppendTTS(segments []types.Segment, videoPath string, language string) (string, error) {

	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	jsonPath, error := utils.CreateJSONFile(map[string]interface{}{"segments": segments}, videoName+".json", "output/json")
	if error != nil {
		log.Fatal(error)
	}

	bgm, error := GenerateBGMAudio(videoPath)
	if error != nil {
		return "", fmt.Errorf("failed to generate bgm: %v", error)
	}

	tts_path, error := BuildTTS(jsonPath, language)
	if error != nil {
		return "", fmt.Errorf("failed to build tts: %v", error)
	}

	mixedAudioPath, error := BuildTTSAudioWithBGM(tts_path, jsonPath, bgm)
	if error != nil {
		return "", fmt.Errorf("failed to build tts with bgm: %v", error)
	}

	fullOutputPath, error := ReplaceVideoAudio(videoPath, mixedAudioPath, filepath.Dir(videoPath))
	if error != nil {
		return "", fmt.Errorf("failed to replace video audio: %v", error)
	}

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

func GenerateBGMAudio(mediaPath string) (string, error) {
	scriptPath := filepath.Join(".", "scripts/text-to-speech-scripts/split-BGM.py")

	args := []string{
		scriptPath,
		mediaPath,
	}
	output, err := utils.ExecExternalScript(args, "python")
	if err != nil {
		return "", fmt.Errorf("BGM process failed: %v\nError output: %s", err, string(output))
	}

	audioName := strings.TrimSuffix(filepath.Base(mediaPath), filepath.Ext(mediaPath))

	result := filepath.Join(".", "separated", "htdemucs", audioName+"-audio", "no_vocals.wav")

	log.Printf(result)

	return result, nil
}

func BuildTTS(transcriptsPath string, language string) (string, error) {
	scriptPath := filepath.Join(".", "scripts/text-to-speech-scripts/generate-tts-from-segments.py")
	name := strings.TrimSuffix(filepath.Base(transcriptsPath), filepath.Ext(transcriptsPath))
	name = strings.TrimSuffix(name, "_"+language)
	outputDir := filepath.Join(".", fmt.Sprintf("output/tts/%s/%s", name, language))

	args := []string{
		scriptPath,
		transcriptsPath,
		"--language", language,
		"--output", outputDir,
	}

	output, err := utils.ExecExternalScript(args, "python")

	if err != nil {
		return "", fmt.Errorf("TTS process failed: %v\nError output: %s", err, string(output))
	}

	result := filepath.Join(".", outputDir)

	return result, nil
}
