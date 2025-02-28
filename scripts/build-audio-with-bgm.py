import argparse
from ast import arg
import os
import soundfile as sf
import json
import subprocess



def buildTTSAudioWithBGM(tts_audio_folder_path, script_path, bgm_path):
     # Get the base folder name to match the naming convention
    folder_name = os.path.basename(tts_audio_folder_path)
    
    # List to store audio file details
    tts_audio_files = []
    
    # Iterate through files in the TTS path
    for filename in sorted(os.listdir(tts_audio_folder_path)):
        # Check if the file matches the naming convention
        if filename.startswith(f"{folder_name}_") and filename.endswith(".wav"):
            full_path = os.path.join(tts_audio_folder_path, filename)
            
            # Get audio length using soundfile
            with sf.SoundFile(full_path) as audio_file:
                audio_length = len(audio_file) / audio_file.samplerate
            
            # Append file details to the list
            tts_audio_files.append({"path": full_path, "length": audio_length})

    # Read the JSON file from the script_path
    with open(script_path, 'r', encoding='utf-8') as script_file:
        blocks= json.load(script_file)

    
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

        # if tts_file:
        #     # Only speed up if there's a significant difference
        #     if (tts_file["length"] - length) > 0.05:

        #         # Create output path for speed-adjusted audio
        #         output_path = os.path.join(tts_audio_folder_path, f"{folder_name}_{id}_adjusted.wav")
                
        #         # Speed up/slow down the audio
        #         speedUpAudio(1+(length / tts_file["length"]), tts_file["path"], output_path)

        #         # Update the path to the new adjusted audio file
        #         tts_file["path"] = output_path
                

    # Append TTS audio to BGM
    output_bgm_path = os.path.join(os.path.dirname(bgm_path), 'final_mixed_audio.wav')
    appendAudioToBGM(bgm_path, tts_audio_files, output_bgm_path)


def appendAudioToBGM(bgm_path, audio_list, output_path):
    print("---------------------------------------------------")
    print(bgm_path)
    print("---------------------------------------------------")
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

# def speedUpAudio(ratio, input_path, output_path):
#     ratio = min(ratio, 1.15)
    
#     command = f'ffmpeg -i {input_path} -filter:a "atempo={ratio}" {output_path}'
#     subprocess.run(command, shell=True)


# def calculateAudioSpeed(original_length, target_length):
#     return target_length / original_length

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Build audio with BGM"
    )
    parser.add_argument(
        "tts_audio_folder_path", type=str, help="Path to the TTS audio files folder"
    )
    parser.add_argument("script_path", type=str, help="Path to the script file")
    parser.add_argument("bgm_path", type=str, help="Path to the background music file")

    args = parser.parse_args()

    final_video = buildTTSAudioWithBGM(
        args.tts_audio_folder_path, 
        args.script_path, 
        args.bgm_path
    )