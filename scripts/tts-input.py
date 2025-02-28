import edge_tts
import asyncio
import os
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

async def processTTS(input_text, name, language="en"):
    if not os.path.exists("temporary-output"):
        os.makedirs("temporary-output")

    tts = edge_tts.Communicate(input_text, get_language_code(language))
    await tts.save(f"temporary-output/{name}.wav")


async def main():
    parser = argparse.ArgumentParser(
        description="Translate JSON files using Hugging Face models"
    )
    parser.add_argument("input")
    parser.add_argument(
        "--name",
        help="Name for the output audio file series",
    )
    parser.add_argument(
        "--language",
        choices=["en", "fr", "de", "ja", "ru", "zh", "es", "vi"],
        help="Language for the input text",
    )
    args = parser.parse_args()
    await processTTS(args.input, args.name, args.language)


if __name__ == "__main__":
    asyncio.run(main())
