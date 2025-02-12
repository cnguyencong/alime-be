import argparse
from faster_whisper import WhisperModel
import argparse
import json
import os

def transcribe_media(media_path, model_size='base', device='auto', output_path="", output_name=""):
    model = WhisperModel(model_size, device=device, compute_type="float16")
    
    # Ensure output directory exists
    if output_path:
        os.makedirs(output_path, exist_ok=True)
    
    # Construct output file path
    if output_path and output_name:
        output_file = os.path.join(output_path, f"{os.path.splitext(os.path.basename(output_name))[0]}.json")
    else:
        output_file = f"{os.path.splitext(media_path)[0]}.json"
    
    # Transcribe
    segments, info = model.transcribe(media_path, beam_size=5)
    
    # Prepare segments data
    captions = []
    for index, segment in enumerate(segments):
        captions.append({
            "ID": index,
            "start": segment.start,
            "end": segment.end,
            "text": segment.text
        })
    
    # Write to JSON file
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump({
            "language": info.language,
            "language_probability": info.language_probability,
            "segments": captions
        }, f, ensure_ascii=False, indent=2)
    return captions


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Transcribe audio using Faster-Whisper")
    parser.add_argument("audio", type=str, help="Path to the audio file")
    parser.add_argument("--model", type=str, default="medium", help="Model size (tiny, base, small, medium, large-v2)")
    parser.add_argument("--device", type=str, default="auto", help="Device (cpu, cuda, auto)")
    parser.add_argument("--output-path", type=str, default="", help="Output directory")
    parser.add_argument("--output-name", type=str, default="captions.json", help="Output file name")

    args = parser.parse_args()
    transcribe_media(args.audio, args.model, args.device, args.output_path, args.output_name)