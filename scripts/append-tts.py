import argparse
from ast import arg
import os
import soundfile as sf
import json
import subprocess


def appendTTSToBgm(tts_path, script_path, bgm_path, video_path):
    # Get the base folder name to match the naming convention
    folder_name = os.path.basename(tts_path)
    
    # List to store audio file details
    tts_audio_files = []
    
    # Iterate through files in the TTS path
    for filename in sorted(os.listdir(tts_path)):
        # Check if the file matches the naming convention
        if filename.startswith(f"{folder_name}_") and filename.endswith(".wav"):
            full_path = os.path.join(tts_path, filename)
            
            # Get audio length using soundfile
            with sf.SoundFile(full_path) as audio_file:
                audio_length = len(audio_file) / audio_file.samplerate
            
            # Append file details to the list
            tts_audio_files.append({"path": full_path, "length": audio_length})

    # Read the JSON file from the script_path
    with open(script_path, 'r', encoding='utf-8') as script_file:
        script_data = json.load(script_file)

    blocks = script_data.get("segments", [])
    
    for block in blocks:
        start = block.get("start", 0)
        end = block.get("end", 0)
        length = end - start

        id = block.get("id", 0)

        # Find the corresponding TTS file
        tts_file = next(
            (
                i
                for i in tts_audio_files
                if i["path"].endswith(f"{folder_name}_{id}.wav")
            ),
            None,
        )
        tts_file["start"] = start

        if tts_file:
            # Only speed up if there's a significant difference
            if (tts_file["length"] - length) > 0.05:

                # Create output path for speed-adjusted audio
                output_path = os.path.join(tts_path, f"{folder_name}_{id}_adjusted.wav")
                
                # Speed up/slow down the audio
                # speedUpAudio(1+(length / tts_file["length"]), tts_file["path"], output_path)

                # Update the path to the new adjusted audio file
                # tts_file["path"] = output_path
                # tts_file["length"] = length  # Update length to match target

    # Append TTS audio to BGM
    output_bgm_path = os.path.join(os.path.dirname(bgm_path), 'final_mixed_audio.wav')
    appendAudioToBGM(bgm_path, tts_audio_files, output_bgm_path)

    # Generate output video path
    output_video_path = os.path.join(
        os.path.dirname(video_path), 
        f'video_with_new_audio_{os.path.basename(video_path)}'
    )

    # Replace video audio with mixed audio
    final_video_path = replaceVideoAudio(video_path, output_bgm_path, output_video_path)

    # Delete the temporary adjusted audio files
    # os.remove(tts_path)
    # os.remove(os.path.dirname(bgm_path))

    return final_video_path


def appendAudioToBGM(bgm_path, audio_list, output_path):
    """
    Append multiple audio files to a background music track at specific start times.
    
    :param bgm_path: Path to the background music file
    :param audio_list: List of dictionaries with 'path' and 'start' keys
    :param output_path: Path to save the final mixed audio
    """
    # Sort audio list by start time
    sorted_audio_list = sorted(audio_list, key=lambda x: x.get('start', 0))
    
    # Method 1: Using filter complex with adelay
    filter_inputs = []
    filter_complex = []
    
    # Add background music as the first input
    filter_inputs.append(f'-i {bgm_path}')
    
    # Add each audio file and create delay filters
    for i, audio in enumerate(sorted_audio_list, 1):
        filter_inputs.append(f'-i {audio["path"]}')
        
        # Calculate delay in milliseconds
        delay_ms = int(audio.get('start', 0) * 1000)
        
        # Create adelay filter for each input
        filter_complex.append(f'[{i}:a]adelay={delay_ms}|{delay_ms}[a{i}]')
    
    # Combine all inputs
    if len(sorted_audio_list) > 0:
        input_streams = ''.join([f'[a{i+1}]' for i in range(len(sorted_audio_list))])
        filter_complex.append(f'{input_streams}[0:a]amix=inputs={len(sorted_audio_list)+1}:duration=longest[aout]')
    
    # Construct the full FFmpeg command
    command = (
        f'ffmpeg '
        f'{" ".join(filter_inputs)} '
        f'-filter_complex "{";".join(filter_complex)}" '
        f'-map "[aout]" '
        f'{output_path}'
    )
    
    print("FFmpeg Command:", command)
    subprocess.run(command, shell=True)


def replaceVideoAudio(video_path, audio_path, output_path):
    """
    Replace the audio in a video file with a new audio track.
    
    :param video_path: Path to the original video file
    :param audio_path: Path to the new audio file to replace the original audio
    :param output_path: Path to save the new video with replaced audio
    """
    # FFmpeg command to replace audio
    command = (
        f'ffmpeg -i {video_path} -i {audio_path} '
        f'-map 0:v -map 1:a '  # Map video from first input, audio from second input
        f'-c:v copy '  # Copy video codec without re-encoding
        f'-c:a aac '  # Re-encode audio to AAC 
        f'-shortest '  # Stop encoding when the shortest input stream ends
        f'{output_path}'
    )
    
    print("Replace Video Audio Command:", command)
    subprocess.run(command, shell=True)
    
    return output_path


def speedUpAudio(ratio, input_path, output_path):
    command = f'ffmpeg -i {input_path} -filter:a "atempo={ratio}" {output_path}'
    subprocess.run(command, shell=True)


def calculateAudioSpeed(original_length, target_length):
    return target_length / original_length


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Transcribe audio using Faster-Whisper"
    )
    parser.add_argument(
        "tts_audio_folder_path", type=str, help="Path to the TTS audio files folder"
    )
    parser.add_argument("script_path", type=str, help="Path to the script file")
    parser.add_argument("bgm_path", type=str, help="Path to the background music file")
    parser.add_argument("video_path", type=str, help="Path to the video file")

    args = parser.parse_args()

    final_video = appendTTSToBgm(
        args.tts_audio_folder_path, 
        args.script_path, 
        args.bgm_path, 
        args.video_path
    )
    print(f"Final video created at: {final_video}")
