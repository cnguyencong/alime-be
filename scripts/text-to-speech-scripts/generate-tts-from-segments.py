import edge_tts
import asyncio
import os
import json
import argparse
import soundfile as sf


language_codes = {
    "en": "en-US-GuyNeural",
    "fr": "fr-FR-HenriNeural",
    "de": "de-DE-ConradNeural",
    "ja": "ja-JP-KeitaNeural",
    "ru": "ru-RU-DmitryNeural",
    "zh": "zh-CN-YunxiNeural",
    "es": "es-ES-AlvaroNeural",
    "vi": "vi-VN-NamMinhNeural",
}


def get_language_code(language):
    return language_codes.get(language, "en-US-GuyNeural")


async def processTTS(input, language="en", output_folder="output/tts"):
    if not os.path.exists(output_folder):
        os.makedirs(output_folder)

    # Read JSON input
    print("\nReading JSON file...")
    with open(input, "r", encoding="utf-8-sig") as f:
        data = json.load(f)

    blocks = data.get("segments", [])
    audio_name = os.path.splitext(os.path.basename(input))[0]
    # Create a list of coroutines for concurrent processing

    audio_info_data = []

    audio_tasks = []
    for block in blocks:
        print(
            f"Generating audio segment {block['id']}",
            f"{output_folder}/{audio_name}_{block['id']}.wav",
        )

        # Create a task for each audio generation
        task = asyncio.create_task(
            generate_audio_segment(
                block, audio_name, language, output_folder, audio_info_data
            )
        )
        audio_tasks.append(task)

    # Wait for all audio generation tasks to complete concurrently
    await asyncio.gather(*audio_tasks)

    # Write audio info to JSON file
    with open(f"{output_folder}/audio_info.json", "w") as f:
        json.dump(audio_info_data, f, indent=4)

    # Return the absolute path of the output folder
    return os.path.abspath(output_folder)


async def generate_audio_segment(
    block, audio_name, language, output_path, audio_info_data
):
    tts = edge_tts.Communicate(block["text"], get_language_code(language))
    output_path = f"{output_path}/{audio_name}_{block['id']}.wav"
    await tts.save(output_path)

    audio_file = sf.SoundFile(output_path)
    audio_length = len(audio_file) / audio_file.samplerate

    audio_info = {
        "id": block["id"],
        "audioLength": audio_length,
        "audioPath": output_path,
    }
    audio_info_data.append(audio_info)


async def main():
    parser = argparse.ArgumentParser(
        description="Translate JSON files using Hugging Face models"
    )
    parser.add_argument("input")
    parser.add_argument(
        "--language",
        choices=["en", "fr", "de", "ja", "ru", "zh", "es", "vi"],
        help="Language for the input text",
    )
    parser.add_argument("--output", default="output/tts")

    args = parser.parse_args()

    await processTTS(args.input, args.language, args.output)


if __name__ == "__main__":
    asyncio.run(main())
