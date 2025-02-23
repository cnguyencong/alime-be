import edge_tts
import asyncio
import os
import json
import argparse


language_codes = {
    'en' : 'en-US-GuyNeural',
    'fr' : 'fr-FR-HenriNeural',
    'de': 'de-DE-ConradNeural',
    'ja': 'ja-JP-KeitaNeural',
    'ru': 'ru-RU-DmitryNeural',
    'zh': 'zh-CN-YunxiNeural',
    'es': 'es-ES-AlvaroNeural',
    'vi': 'vi-VN-NamMinhNeural'
}

def get_language_code(language):
    return language_codes.get(language, 'en-US-GuyNeural')
    

# async def generate_adio_block(text,  output, language):
#   tts = edge_tts.Communicate(text, get_language_code(language))
#   print(text, get_language_code(language))
#   tts.save(output)

async def processTTS(input_file, name):
    # Read JSON input
    print("\nReading JSON file...")
    with open(input_file, "r", encoding="utf-8") as f:
        data = json.load(f)


    audio_name = name

    if not os.path.exists("temporary-output"):
        os.makedirs("temporary-output")
    

    blocks = data.get("segments", [])
    for block in blocks:
        print(f"Generating audio segment {block['id']}", f"{audio_name}_{block['id']}.wav")

        tts = edge_tts.Communicate(block['text'], get_language_code(block['language'])) 
        await tts.save(f"temporary-output/{audio_name}_{block['id']}.wav")

        # await generate_adio_block(block['text'],  f"{audio_name}_{block['id']}.wav", block['language'])


async def main():
    # tts = edge_tts.Communicate(" But in reality, the dog's name was Terry, and when Terry died in 1945, her owner and", "en-US-GuyNeural") 
    # await tts.save("output.wav")

    parser = argparse.ArgumentParser(
        description="Translate JSON files using Hugging Face models"
    )
    parser.add_argument("input_file", help="Path to input JSON file")
    parser.add_argument(
        "--name",
        help="Name for the output audio file series",
    )
    args = parser.parse_args()
    await processTTS(args.input_file, args.name)

if __name__ == "__main__":
    asyncio.run(main())

# os.system("start output.mp3")