import argparse
import subprocess
import soundfile as sf
from pathlib import Path


class FFmpegError(Exception):
    """Custom exception for FFmpeg-related errors."""

    pass


def run_ffmpeg_command(cmd: str) -> None:
    """Execute FFmpeg command and handle errors."""
    try:
        subprocess.run(
            cmd,
            shell=True,
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            universal_newlines=True,
        )
    except subprocess.CalledProcessError as e:
        raise FFmpegError(f"FFmpeg command failed: {e.stderr}") from e


def process_video_with_transition(
    input_path: str, output_path: str, start: float, end: float
) -> None:

    fading_time = end - start

    temp_audio = "temp_audio.wav"
    cmd = f"ffmpeg -i {input_path} -q:a 0 -map a -y {temp_audio}"
    run_ffmpeg_command(cmd)

    with sf.SoundFile(temp_audio) as audio_file:
        duration = len(audio_file) / audio_file.samplerate
        print(duration)
    # mid_point = start + fading_time / 2

    # Single filter complex combining split, fade, and concat
    filter_complex = (
        f"[0:v]split=2[v1][v2];"  # Split video into two streams
        f"[v1]trim=start=0:end={start+fading_time/2},setpts=PTS-STARTPTS,fade=out:st={start}:d={fading_time/2}:color=white[v1f];"  # First part with corrected timestamps
        f"[v2]trim=start={start+fading_time/2}:end={duration},setpts=PTS-STARTPTS,fade=in:st=0:d={fading_time/2}:color=white[v2f];"  # Second part with corrected timestamps
        f"[v1f][v2f]concat=n=2:v=1:a=0[outv]"  # Concatenate the parts
    )

    cmd = (
        f"ffmpeg -i {input_path} -i {temp_audio} "
        f'-filter_complex "{filter_complex}" '
        f'-map "[outv]" -map 1:a '
        f"-c:a copy "
        f"-c:v libx264 -crf 18 "
        f"-c:a aac -b:a 192k "
        f"-y {output_path}"
    )
    run_ffmpeg_command(cmd)
    Path(temp_audio).unlink(missing_ok=True)


def main():
    """Parse arguments and process video."""
    parser = argparse.ArgumentParser(description="Insert fade transition in video")
    parser.add_argument("--input", required=True, help="Input media path")
    parser.add_argument("--output", required=True, help="Output path")
    parser.add_argument(
        "--start", type=float, required=True, help="Transition start time (seconds)"
    )
    parser.add_argument(
        "--end", type=float, required=True, help="Transition end time (seconds)"
    )

    args = parser.parse_args()

    try:
        process_video_with_transition(args.input, args.output, args.start, args.end)
        print(f"Video processing completed successfully. Output saved to {args.output}")
    except FFmpegError as e:
        print(f"Error processing video: {e}")
    except Exception as e:
        print(f"Unexpected error: {e}")


if __name__ == "__main__":
    main()
