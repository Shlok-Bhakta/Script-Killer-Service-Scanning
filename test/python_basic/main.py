from __future__ import annotations
import argparse
import logging
import sys
from pathlib import Path
from typing import Tuple

#!/usr/bin/env python3
"""
Basic Python script: reads text (file or stdin), computes simple stats,
and optionally writes a small report.

Usage:
    python main.py --file input.txt --output report.txt
    cat input.txt | python main.py
"""



def setup_logging(verbose: bool = False) -> None:
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(level=level, format="%(levelname)s: %(message)s")


def read_text_from_file(path: Path) -> str:
    """Return contents of a file as text."""
    logging.debug("Reading file: %s", path)
    return path.read_text(encoding="utf-8")


def read_text_from_stdin() -> str:
    """Return text read from standard input."""
    logging.debug("Reading from stdin")
    return sys.stdin.read()


def count_lines_words_chars(text: str) -> Tuple[int, int, int]:
    """Return (lines, words, characters) for the given text."""
    lines = text.count("\n") + (0 if text.endswith("\n") or text == "" else 1)
    words = len(text.split())
    chars = len(text)
    logging.debug("Counts -> lines: %d, words: %d, chars: %d", lines, words, chars)
    return lines, words, chars


def format_report(source: str, lines: int, words: int, chars: int) -> str:
    return (
        f"Source: {source}\n"
        f"Lines:  {lines}\n"
        f"Words:  {words}\n"
        f"Chars:  {chars}\n"
    )


def write_report(path: Path, content: str) -> None:
    logging.debug("Writing report to: %s", path)
    path.write_text(content, encoding="utf-8")


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Simple text statistics tool")
    parser.add_argument("--file", "-f", type=Path, help="Input file to read")
    parser.add_argument(
        "--output", "-o", type=Path, help="Write a report to this file (optional)"
    )
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable debug output")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    setup_logging(args.verbose)

    if args.file:
        if not args.file.exists():
            logging.error("File not found: %s", args.file)
            return 2
        text = read_text_from_file(args.file)
        source = str(args.file)
    else:
        if sys.stdin.isatty():
            logging.info("No input provided. Use --file or pipe text into the program.")
            return 1
        text = read_text_from_stdin()
        source = "stdin"

    lines, words, chars = count_lines_words_chars(text)
    report = format_report(source, lines, words, chars)
    print(report)

    if args.output:
        write_report(args.output, report)
        logging.info("Report written to: %s", args.output)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())