# scripts/split-BGM.py
from demucs.separate import main
import argparse
import subprocess
import os


def extract_audio(media_path):
    file_name = f"{os.path.splitext(os.path.basename(media_path))[0]}-audio.wav"
    output = os.path.join('output', file_name)
    command = (
        f"ffmpeg -i {media_path} -q:a 0 -map a {output}"
    )
    subprocess.run(command, shell=True)
    return output

def split_voice_and_background(media_path):
    main(["--two-stems=vocals", media_path])

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Split audio with BGM"
    )
    parser.add_argument("input", type=str, help="Path to the media file")
    args = parser.parse_args()

    extracted_audio = extract_audio(args.input)

    split_voice_and_background(extracted_audio)
    os.remove(extracted_audio)