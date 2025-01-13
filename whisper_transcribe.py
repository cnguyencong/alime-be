import whisper
import argparse
import os
import torch
import warnings
from datetime import timedelta
from tqdm import tqdm
import time
import subprocess
import tempfile
import sys

warnings.filterwarnings("ignore", category=FutureWarning)

# Set UTF-8 encoding for stdout
if sys.platform.startswith('win'):
    import locale
    sys.stdout.reconfigure(encoding='utf-8')
    sys.stderr.reconfigure(encoding='utf-8')

def format_timestamp(seconds):
    """Convert seconds to SRT timestamp format"""
    td = timedelta(seconds=seconds)
    hours = td.seconds//3600
    minutes = (td.seconds//60)%60
    seconds = td.seconds%60
    milliseconds = td.microseconds//1000
    return f"{hours:02d}:{minutes:02d}:{seconds:02d},{milliseconds:03d}"

def create_srt_content(segments):
    """Create SRT formatted content from segments"""
    srt_content = []
    for i, segment in enumerate(segments, start=1):
        start_time = format_timestamp(segment['start'])
        end_time = format_timestamp(segment['end'])
        text = segment['text'].strip()
        srt_content.append(f"{i}\n{start_time} --> {end_time}\n{text}\n")
    return "\n".join(srt_content)

def extract_audio(video_path):
    """Extract audio from video file using ffmpeg"""
    temp_audio = tempfile.NamedTemporaryFile(suffix='.wav', delete=False)
    try:
        subprocess.run([
            'ffmpeg', '-i', video_path,
            '-vn',  # Disable video
            '-acodec', 'pcm_s16le',  # Audio codec
            '-ar', '16000',  # Sample rate
            '-ac', '1',  # Mono audio
            '-y',  # Overwrite output file
            temp_audio.name
        ], check=True, capture_output=True)
        return temp_audio.name
    except subprocess.CalledProcessError as e:
        print(f"Error extracting audio: {e.stderr.decode()}")
        if os.path.exists(temp_audio.name):
            os.unlink(temp_audio.name)
        raise

def transcribe_audio(
    media_path,
    model_size='base',
    language=None,
    output_formats=None,
    output_dir=None,
    is_video=False
):
    """
    Transcribe audio/video using Whisper

    Parameters:
    - media_path: Path to audio/video file
    - model_size: Whisper model size ('tiny', 'base', 'small', 'medium', 'large')
    - language: Language code (e.g., 'en', 'ja', 'auto' for auto-detection)
    - output_formats: List of output formats ('srt', 'txt', 'vtt', 'json')
    - output_dir: Directory to save output files (default: same as input file)
    - is_video: Whether the input is a video file
    """

    # Check for GPU
    device = "cuda" if torch.cuda.is_available() else "cpu"
    print(f"\nUsing device: {device}")
    if device == "cuda":
        print(f"GPU: {torch.cuda.get_device_name(0)}")

    # Extract audio if input is video
    audio_path = None
    try:
        if is_video:
            print("Extracting audio from video...")
            audio_path = extract_audio(media_path)
            input_path = audio_path
        else:
            input_path = media_path

        # Load model
        print(f"\nLoading model: {model_size}")
        model = whisper.load_model(model_size).to(device)

        # Get audio duration for progress bar
        import ffmpeg
        try:
            probe = ffmpeg.probe(input_path)
            duration = float(probe['streams'][0]['duration'])
            print(f"\nAudio duration: {timedelta(seconds=int(duration))}")
        except:
            duration = None
            print("\nCouldn't determine audio duration")

        # Transcribe with progress indicator
        print("\nTranscribing...")
        start_time = time.time()

        options = {
            "task": "transcribe",
            "language": language,
            "verbose": True,  # Enable whisper's built-in progress
        }

        result = model.transcribe(input_path, **options)

        end_time = time.time()
        process_duration = end_time - start_time

        if duration:
            speed_factor = duration / process_duration
            print(f"\nProcessed at {speed_factor:.2f}x real-time speed")
        print(f"Total processing time: {timedelta(seconds=int(process_duration))}")

        # Prepare output directory
        if output_dir is None:
            output_dir = os.path.dirname(media_path)
        os.makedirs(output_dir, exist_ok=True)

        base_name = os.path.splitext(os.path.basename(media_path))[0]
        outputs = {}

        # Save in requested formats
        if output_formats:
            print("\nSaving outputs...")
            for fmt in tqdm(output_formats, desc="Saving files"):
                output_path = os.path.join(output_dir, f"{base_name}.{fmt}")

                if fmt == 'srt':
                    content = create_srt_content(result['segments'])
                elif fmt == 'vtt':
                    # Convert SRT to VTT
                    content = "WEBVTT\n\n" + create_srt_content(result['segments']).replace(',', '.')
                elif fmt == 'txt':
                    content = result['text']
                elif fmt == 'json':
                    import json
                    content = json.dumps(result, indent=2, ensure_ascii=False)

                # Open file with UTF-8 encoding
                with open(output_path, 'w', encoding='utf-8', errors='ignore') as f:
                    f.write(content)
                outputs[fmt] = output_path
                print(f"Saved {fmt.upper()}: {output_path}")

        # Remove temporary audio file if it was created
        if audio_path and os.path.exists(audio_path):
            os.unlink(audio_path)

        return outputs

    except Exception as e:
        print(f"Error: {str(e)}")

def main():
    parser = argparse.ArgumentParser(description="Transcribe audio/video using Whisper")
    parser.add_argument("input_file", help="Path to the input audio/video file")
    parser.add_argument("--model", default="base", help="Model size (tiny, base, small, medium, large)")
    parser.add_argument("--language", default=None, help="Language code (e.g., en, ja, auto)")
    parser.add_argument("--output-formats", default="srt", help="Comma-separated list of output formats")
    parser.add_argument("--output-dir", default=None, help="Output directory")
    parser.add_argument("--is-video", action="store_true", help="Input is a video file")

    args = parser.parse_args()

    formats = args.output_formats.split(",")
    transcribe_audio(
        args.input_file,
        model_size=args.model,
        language=args.language,
        output_formats=formats,
        output_dir=args.output_dir,
        is_video=args.is_video
    )

if __name__ == "__main__":
    main()
