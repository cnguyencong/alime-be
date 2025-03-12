import argparse
from ast import arg
import os
import json
import subprocess


def buildTTSAudioWithBGM(tts_audio_folder_path, script_path, bgm_path):
    # Get the base folder name to match the naming convention
    folder_name = os.path.basename(tts_audio_folder_path)

    # List to store audio file details
    tts_audio_file_path_list = []

    print(tts_audio_folder_path)

    for filename in sorted(os.listdir(tts_audio_folder_path)):
        if filename.endswith(".wav"):
            full_path = os.path.join(tts_audio_folder_path, filename)
            tts_audio_file_path_list.append(full_path)

    # Read the JSON file from the script_path
    with open(script_path, "r", encoding="utf-8-sig") as script_file:
        data = json.load(script_file)

    tts_audio = []

    blocks = data.get("segments", [])
    for block in blocks:
        id = block.get("id", 0)

        # Find the corresponding TTS file
        tts_file = next(
            (i for i in tts_audio_file_path_list if i.endswith(f"_{id}.wav")),
            None,
        )
        tts_audio.append({"path": tts_file, "start": block.get("start", 0)})

    # Append TTS audio to BGM
    output_bgm_path = os.path.join(os.path.dirname(bgm_path), "final_mixed_audio.wav")
    appendAudioToBGM(bgm_path, tts_audio, output_bgm_path)


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
    sorted_audio_list = sorted(audio_list, key=lambda x: x.get("start", 0))

    # Method 1: Using filter complex with adelay
    filter_inputs = []
    filter_complex = []

    # Add background music as the first input
    filter_inputs.append(f"-i {bgm_path}")

    # Add each audio file and create delay filters
    for i, audio in enumerate(sorted_audio_list, 1):
        filter_inputs.append(f'-i {audio["path"]}')

        # Calculate delay in milliseconds
        delay_ms = int(audio.get("start", 0) * 1000)

        # Create adelay filter for each input
        filter_complex.append(f"[{i}:a]adelay={delay_ms}|{delay_ms}[a{i}]")

    # Combine all inputs
    if len(sorted_audio_list) > 0:
        input_streams = "".join([f"[a{i+1}]" for i in range(len(sorted_audio_list))])
        filter_complex.append(
            f"{input_streams}[0:a]amix=inputs={len(sorted_audio_list)+1}:duration=longest[aout]"
        )

    # Construct the full FFmpeg command
    command = (
        f"ffmpeg "
        f'{" ".join(filter_inputs)} '
        f'-filter_complex "{";".join(filter_complex)}" '
        f'-map "[aout]" '
        f"{output_path}"
    )

    print("FFmpeg Command:", command)
    subprocess.run(command, shell=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Build audio with BGM")
    parser.add_argument(
        "tts_audio_folder_path", type=str, help="Path to the TTS audio files folder"
    )
    parser.add_argument("script_path", type=str, help="Path to the script file")
    parser.add_argument("bgm_path", type=str, help="Path to the background music file")

    args = parser.parse_args()

    final_video = buildTTSAudioWithBGM(
        args.tts_audio_folder_path, args.script_path, args.bgm_path
    )
