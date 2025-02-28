import edge_tts
import asyncio
import os
import json
import argparse


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


async def processTTS(input, language="en"):
    output_folder = os.path.join(
        "temporary-output", os.path.splitext(os.path.basename(input))[0]
    )
    if not os.path.exists(output_folder):
        os.makedirs(output_folder)

    # Read JSON input
    print("\nReading JSON file...")
    with open(input, "r", encoding="utf-8") as f:
        blocks = json.load(f)

    # blocks = data.get("segments", [])

    audio_name = os.path.splitext(os.path.basename(input))[0]
    # Create a list of coroutines for concurrent processing
    audio_tasks = []
    for block in blocks:
        print(
            f"Generating audio segment {block['id']}",
            f"{audio_name}_{block['id']}.wav",
        )

        # Create a task for each audio generation
        task = asyncio.create_task(
            generate_audio_segment(block, audio_name, language, output_folder)
        )
        audio_tasks.append(task)

    # Wait for all audio generation tasks to complete concurrently
    await asyncio.gather(*audio_tasks)

    # Return the absolute path of the output folder
    return os.path.abspath(output_folder)


async def generate_audio_segment(block, audio_name, language, output_path):
    tts = edge_tts.Communicate(block["text"], get_language_code(language))
    await tts.save(f"{output_path}/{audio_name}_{block['id']}.wav")


def splitAudio(file_path):
    return


def processAppendTTS(file_path, language="en"):
    return


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

    args = parser.parse_args()

    print(args)
    await processTTS(args.input, args.language)


if __name__ == "__main__":
    asyncio.run(main())
