#!/usr/bin/env python3

from pathlib import Path
import re


OLD = '`xml:"ds:Signature"`'
NEW = '`xml:"http://www.w3.org/2000/09/xmldsig# Signature"`'
ANON_COMPLEX_XMLNAME = re.compile(r'`xml:"TAnonComplex_([^"_]+(?:_[^"_]+)*)_\d+"`')


def main() -> None:
    root = Path(__file__).resolve().parent / "v4_0"
    for path in root.rglob("*.go"):
        text = path.read_text()
        updated = text.replace(OLD, NEW)
        updated = ANON_COMPLEX_XMLNAME.sub(lambda m: f'`xml:"{m.group(1)}"`', updated)
        if updated == text:
            continue
        path.write_text(updated)
        print(f"postprocessed generated xml tags in {path}")


if __name__ == "__main__":
    main()
