import argparse
from transformers import AutoModelForSeq2SeqLM, AutoTokenizer
import torch
import re
import os
import json
from tqdm import tqdm


def translate_text(text, model, tokenizer, device, target_lang="vi", max_length=512):
    """Translate text to a specific language using a Hugging Face model."""

    # Tokenize the input text
    inputs = tokenizer(
        text, return_tensors="pt", padding=True, truncation=True, max_length=max_length
    )
    inputs = {k: v.to(device) for k, v in inputs.items()}

    # Generate translation
    with torch.no_grad():
        outputs = model.generate(
            **inputs, max_length=max_length, num_beams=5, length_penalty=1.0,
            forced_bos_token_id=tokenizer.lang_code_to_id[f"{target_lang}"]
        )

    # Decode the output
    translation = tokenizer.decode(outputs[0], skip_special_tokens=True)

    return translation


def translate_json(
    input_file, output_file, model_name, batch_size=8, device=None, target_language="vi"
):
    """
    Translate an json file using a Hugging Face translation model

    Parameters:
    - input_file: Path to input JSON file
    - output_file: Path to output translated JSON file
    - model_name: Name of the Hugging Face model to use
    - batch_size: Number of segments to translate at once
    - device: Device to use for translation (cuda/cpu)
    """

    if device is None:
        device = "cuda" if torch.cuda.is_available() else "cpu"

    print(f"\nUsing device: {device}")
    if device == "cuda":
        print(f"GPU: {torch.cuda.get_device_name(0)}")

    # Load model and tokenizer
    print(f"\nLoading model: {model_name}")
    model = AutoModelForSeq2SeqLM.from_pretrained(model_name).to(device)
    tokenizer = AutoTokenizer.from_pretrained(model_name)

    # Read input SRT file
    print("\nReading JSON file...")
    with open(input_file, "r", encoding="utf-8") as f:
        data = json.load(f)
        # content = f.read()

    # Split into blocks
    # blocks = re.split('\n\n+', content.strip())
    blocks = data.get("segments", [])
    translated_blocks = []

    for i in tqdm(range(0, len(blocks), batch_size)):
        batch_blocks = blocks[i : i + batch_size]
        for block in batch_blocks:

            translation = translate_text(
                block.get("text"), model, tokenizer, device, target_lang=target_language
            )
            print(translation)
            # parsed = split_srt_block(block)

            translated_block = {
                "ID": block.get("ID"),
                "text": translation,
                "start": block.get("start"),
                "end": block.get("end"),
            }
            translated_blocks.append(translated_block)

    # Write translated JSON
    print("\nWriting translated JSON file...")
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump({"segments": translated_blocks}, f, indent=4, ensure_ascii=False)

    print(f"\nTranslation completed! Output saved to: {output_file}")

    # if parsed:
    # idx, timestamp, text = parsed

    # translated_block = f"{idx}\n{timestamp}\n{translation}"
    # batch_translations.append(translated_block)

    # translated_blocks.extend(batch_translations)

    # Write translated SRT
    # print("\nWriting translated SRT file...")
    # with open(output_file, 'w', encoding='utf-8') as f:
    #     f.write('\n\n'.join(translated_blocks))

    # print(f"\nTranslation completed! Output saved to: {output_file}")
    # return output_file


def main():
    parser = argparse.ArgumentParser(
        description="Translate JSON files using Hugging Face models"
    )
    parser.add_argument("input_file", help="Path to input JSON file")
    parser.add_argument(
        "--model",
        default="facebook/m2m100_418M",
        help="Hugging Face model to use for translation",
    )
    parser.add_argument(
        "--output-dir",
        default=None,
        help="Directory to save translated file (default: same as input)",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=8,
        help="Number of segments to translate at once",
    )
    parser.add_argument(
        "--device",
        choices=["cuda", "cpu"],
        default=None,
        help="Device to use for translation",
    )
    parser.add_argument(
        "--target-language",
        default="en",
        help="Target language for translation",
    )

    args = parser.parse_args()

    # Set output file path
    if args.output_dir:
        os.makedirs(args.output_dir, exist_ok=True)
        output_file = os.path.join(
            args.output_dir,
            f"{os.path.splitext(os.path.basename(args.input_file))[0]}_translated.json",
        )
    else:
        output_file = os.path.splitext(args.input_file)[0] + "_translated.json"

    print(f"Target language: {args.target_language}")

    # Translate
    translate_json(
        args.input_file,
        output_file,
        args.model,
        args.batch_size,
        args.device,
        args.target_language,
    )


if __name__ == "__main__":
    main()
