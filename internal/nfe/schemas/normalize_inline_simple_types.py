#!/usr/bin/env python3

from __future__ import annotations

from collections import defaultdict
from copy import deepcopy
from pathlib import Path
from xml.etree import ElementTree as ET


XS_NS = "http://www.w3.org/2001/XMLSchema"
ET.register_namespace("xs", XS_NS)


def xs(tag: str) -> str:
    return f"{{{XS_NS}}}{tag}"


def normalize_schema(path: Path) -> tuple[int, int]:
    tree = ET.parse(path)
    root = tree.getroot()

    type_insert_at = 0
    for idx, child in enumerate(list(root)):
        if child.tag in {xs("annotation"), xs("import"), xs("include"), xs("redefine"), xs("simpleType")}:
            type_insert_at = idx + 1
            continue
        break

    generated_simple_types = []
    generated_complex_types = []
    counters: dict[str, int] = defaultdict(int)
    complex_counters: dict[str, int] = defaultdict(int)
    flattened_optional_sequences = 0

    for parent in root.iter():
        children = list(parent)
        index = 0
        while index < len(children):
            child = children[index]
            if child.tag != xs("sequence") or child.get("minOccurs") != "0":
                index += 1
                continue

            direct_elements = [node for node in list(child) if node.tag == xs("element")]
            compositors = [node for node in list(child) if node.tag in {xs("sequence"), xs("choice"), xs("group")}]
            if not direct_elements or compositors:
                index += 1
                continue

            for element in direct_elements:
                element.set("minOccurs", "0")

            parent.remove(child)
            for offset, element in enumerate(direct_elements):
                parent.insert(index + offset, element)

            flattened_optional_sequences += 1
            children = list(parent)
            index += len(direct_elements)

    for element in root.findall(f".//{xs('element')}"):
        inline_complex_type = element.find(xs("complexType"))
        if inline_complex_type is not None:
            name = element.get("name")
            if name:
                complex_counters[name] += 1
                type_name = f"TAnonComplex_{name}_{complex_counters[name]}"

                complex_type = deepcopy(inline_complex_type)
                complex_type.set("name", type_name)

                element.remove(inline_complex_type)
                element.set("type", type_name)

                generated_complex_types.append(complex_type)

        inline_simple_type = element.find(xs("simpleType"))
        if inline_simple_type is None:
            continue

        name = element.get("name")
        if not name:
            continue

        counters[name] += 1
        type_name = f"TAnon_{name}_{counters[name]}"

        simple_type = deepcopy(inline_simple_type)
        simple_type.set("name", type_name)

        element.remove(inline_simple_type)
        element.set("type", type_name)

        generated_simple_types.append(simple_type)

    generated_types = generated_simple_types + generated_complex_types
    for offset, generated_type in enumerate(generated_types):
        root.insert(type_insert_at + offset, generated_type)

    tree.write(path, encoding="utf-8", xml_declaration=True)
    return len(generated_simple_types) + len(generated_complex_types), flattened_optional_sequences


def main() -> None:
    schema_dir = Path(__file__).with_name("v4_0").joinpath("nfe_proc")
    for schema in sorted(schema_dir.glob("*.xsd")):
        changed, flattened = normalize_schema(schema)
        print(f"normalized {changed} inline simpleType elements in {schema}")
        print(f"flattened {flattened} optional direct-element sequences in {schema}")


if __name__ == "__main__":
    main()
