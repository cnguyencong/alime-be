import argparse
from transformers import AutoModelForSeq2SeqLM, AutoTokenizer
import torch
import re
import os
import json
from tqdm import tqdm


def get_language_id(tokenizer, lang_code):
    try:
        # Try different token formats
        tokens_to_try = [
            f"__{lang_code}__",
            f"__{lang_code}",
            lang_code
        ]
        
        for token in tokens_to_try:
            lang_id = tokenizer.convert_tokens_to_ids(token)
            if lang_id != tokenizer.unk_token_id:
                return lang_id
        
        raise ValueError(f"Could not find language ID for {lang_code}")
    
    except Exception as e:
        print(f"Error finding language ID: {e}")
        print("Available languages:", list(tokenizer.get_vocab().keys()))
        raise


def translate_text(
    batch_texts, model, tokenizer, device, target_lang="vi", max_length=512
):
    """Translate a batch of texts at once."""

    # Tokenize input batch
    inputs = tokenizer(
        batch_texts,
        return_tensors="pt",
        padding=True,
        truncation=True,
        max_length=max_length,
    )
    inputs = {k: v.to(device) for k, v in inputs.items()}

    # Generate translations
    with torch.no_grad():
        outputs = model.generate(
            **inputs,
            max_length=max_length,
            num_beams=4,
            length_penalty=1.0,
            # forced_bos_token_id=tokenizer.lang_code_to_id[target_lang],
            forced_bos_token_id=get_language_id(tokenizer, target_lang),
        )

    # Decode the batch
    return tokenizer.batch_decode(outputs, skip_special_tokens=True)

    
def translate_json(
    input_file, output_file, model_name, batch_size=8, device=None, target_language="Vie_Latn"
):
    """Optimized translation using batched processing"""

    if device is None:
        device = "cuda" if torch.cuda.is_available() else "cpu"

    print(f"\nUsing device: {device}")
    if device == "cuda":
        print(f"GPU: {torch.cuda.get_device_name(0)}")
        model = AutoModelForSeq2SeqLM.from_pretrained(model_name).to(device).half()
    else:
        model = AutoModelForSeq2SeqLM.from_pretrained(model_name).to(device)


    # Load model and tokenizer
    print(f"\nLoading model: {model_name}")
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    # Read JSON input
    print("\nReading JSON file...")
    with open(input_file, "r", encoding="utf-8") as f:
        data = json.load(f)

    blocks = data.get("segments", [])
    translated_blocks = []

    for i in tqdm(range(0, len(blocks), batch_size)):
        batch = blocks[i : i + batch_size]
        batch_texts = [b["text"] for b in batch]

        # 🔥 Batch translation 🔥
        translations = translate_text(
            batch_texts, model, tokenizer, device, target_lang=target_language
        )

        # Store results
        for block, translation in zip(batch, translations):
            translated_blocks.append(
                {
                    "id": block.get("id"),
                    "text": translation,
                    "start": block.get("start"),
                    "end": block.get("end"),
                }
            )

    # Save translated JSON
    print("\nWriting translated JSON file...")
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump({"segments": translated_blocks}, f, indent=4, ensure_ascii=False)

    print(f"\nTranslation completed! Output saved to: {output_file}")


def main():
    parser = argparse.ArgumentParser(
        description="Translate JSON files using Hugging Face models"
    )
    parser.add_argument("input_file", help="Path to input JSON file")
    parser.add_argument(
        "--model",
        default="facebook/nllb-200-distilled-600M",
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
