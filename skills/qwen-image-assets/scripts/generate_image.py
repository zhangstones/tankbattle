#!/usr/bin/env python3
import argparse
import base64
import json
import os
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path


def parse_args():
    parser = argparse.ArgumentParser(
        description="Generate an image asset via a Qwen-compatible image API."
    )
    parser.add_argument("--prompt", required=True, help="Main prompt text.")
    parser.add_argument("--out", required=True, help="Output PNG path.")
    parser.add_argument("--negative-prompt", help="Negative prompt, if supported.")
    parser.add_argument("--size", default="1024x1024", help="Image size, e.g. 1024x1024.")
    parser.add_argument("--model", help="Override model name.")
    parser.add_argument("--style", help="Optional style field.")
    parser.add_argument("--quality", help="Optional quality field.")
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Validate endpoint and payload without sending a network request.",
    )
    return parser.parse_args()


def require_env(name: str) -> str:
    value = os.getenv(name)
    if not value:
        raise SystemExit(f"Missing required environment variable: {name}")
    return value


def detect_api_mode(raw_url: str, model: str) -> str:
    mode = os.getenv("QWEN_IMAGE_API_MODE", "auto").strip().lower()
    if mode in {"openai-images", "dashscope-qwen-image"}:
        return mode

    parsed = urllib.parse.urlparse(raw_url)
    if "dashscope" in parsed.netloc and "qwen-image" in model.lower():
        return "dashscope-qwen-image"

    return "openai-images"


def get_endpoint(mode: str) -> str:
    raw = require_env("QWEN_IMAGE_API_URL").strip()
    parsed = urllib.parse.urlparse(raw)
    if not parsed.scheme or not parsed.netloc:
        raise SystemExit("QWEN_IMAGE_API_URL must be an absolute URL.")

    if mode == "dashscope-qwen-image":
        return urllib.parse.urlunparse(
            parsed._replace(path="/api/v1/services/aigc/multimodal-generation/generation")
        )

    path = parsed.path.rstrip("/")
    last_segment = path.split("/")[-1] if path else ""
    if not path or last_segment == "v1":
        path = f"{path}/images/generations".replace("//", "/")
        parsed = parsed._replace(path=path)
        return urllib.parse.urlunparse(parsed)

    return raw


def normalize_size(mode: str, size: str) -> str:
    return size.replace("x", "*") if mode == "dashscope-qwen-image" else size


def build_payload(args: argparse.Namespace, mode: str) -> dict:
    model = args.model or os.getenv("QWEN_IMAGE_MODEL", "qwen-image")
    size = normalize_size(mode, args.size)

    if mode == "dashscope-qwen-image":
        if "qwen-image" not in model.lower():
            raise SystemExit(
                "DashScope qwen-image mode requires an image model, "
                f"but current model is: {model}"
            )
        payload = {
            "model": model,
            "input": {
                "messages": [
                    {
                        "role": "user",
                        "content": [{"text": args.prompt}],
                    }
                ]
            },
            "parameters": {
                "size": size,
                "watermark": False,
            },
        }
        if args.negative_prompt:
            payload["parameters"]["negative_prompt"] = args.negative_prompt
        if args.quality:
            payload["parameters"]["quality"] = args.quality
        if args.style:
            payload["parameters"]["style"] = args.style
    else:
        prompt_field = os.getenv("QWEN_IMAGE_PROMPT_FIELD", "prompt")
        payload = {
            "model": model,
            prompt_field: args.prompt,
            "size": size,
        }
        if args.negative_prompt:
            payload["negative_prompt"] = args.negative_prompt
        if args.style:
            payload["style"] = args.style
        if args.quality:
            payload["quality"] = args.quality

    extra_json = os.getenv("QWEN_IMAGE_EXTRA_JSON")
    if extra_json:
        try:
            extra = json.loads(extra_json)
        except json.JSONDecodeError as exc:
            raise SystemExit(f"QWEN_IMAGE_EXTRA_JSON is not valid JSON: {exc}") from exc
        if not isinstance(extra, dict):
            raise SystemExit("QWEN_IMAGE_EXTRA_JSON must decode to a JSON object.")
        payload.update(extra)

    return payload


def post_json(url: str, payload: dict, mode: str) -> dict:
    api_key = require_env("QWEN_IMAGE_API_KEY")
    timeout = int(os.getenv("QWEN_IMAGE_TIMEOUT", "120"))
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
        "Accept": "application/json",
    }
    if mode == "dashscope-qwen-image":
        headers["X-DashScope-Async"] = "disable"
    body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=body,
        method="POST",
        headers=headers,
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            charset = resp.headers.get_content_charset() or "utf-8"
            return json.loads(resp.read().decode(charset))
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise SystemExit(f"HTTP {exc.code}: {detail}") from exc
    except urllib.error.URLError as exc:
        raise SystemExit(f"Request failed: {exc}") from exc


def walk_candidates(node):
    if isinstance(node, dict):
        yield node
        for value in node.values():
            yield from walk_candidates(value)
    elif isinstance(node, list):
        for item in node:
            yield from walk_candidates(item)


def extract_image_bytes(payload: dict) -> bytes:
    for node in walk_candidates(payload):
        image = node.get("image")
        if isinstance(image, str) and image.startswith(("http://", "https://")):
            with urllib.request.urlopen(image, timeout=120) as resp:
                return resp.read()

    for node in walk_candidates(payload):
        for key in ("b64_json", "image_base64", "base64", "b64"):
            value = node.get(key)
            if isinstance(value, str) and value:
                return base64.b64decode(value)

    for node in walk_candidates(payload):
        for key in ("url", "image_url"):
            value = node.get(key)
            if isinstance(value, str) and value.startswith(("http://", "https://")):
                with urllib.request.urlopen(value, timeout=120) as resp:
                    return resp.read()

    raise SystemExit(
        "Response did not contain a supported image field. "
        "Expected b64_json, image_base64, url, or image_url."
    )


def write_output(out_path: Path, image_bytes: bytes):
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_bytes(image_bytes)


def main():
    args = parse_args()
    raw_url = require_env("QWEN_IMAGE_API_URL").strip()
    model = args.model or os.getenv("QWEN_IMAGE_MODEL", "qwen-image")
    mode = detect_api_mode(raw_url, model)
    endpoint = get_endpoint(mode)
    payload = build_payload(args, mode)
    out_path = Path(args.out).resolve()
    if args.dry_run:
        print(
            json.dumps(
                {
                    "mode": mode,
                    "endpoint": endpoint,
                    "out": str(out_path),
                    "payload": payload,
                },
                ensure_ascii=False,
                indent=2,
            )
        )
        return

    response = post_json(endpoint, payload, mode)
    image_bytes = extract_image_bytes(response)
    write_output(out_path, image_bytes)
    print(out_path)


if __name__ == "__main__":
    main()
