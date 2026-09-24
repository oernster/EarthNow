"""Convert Natural Earth shapefiles to the compact form that
internal/infrastructure/geo embeds. First written for the Phase 0 spike.

Reads the ESRI shapefile (.shp, polygon and point records) and its dBase
table (.dbf) directly from their published layouts, so no package is needed.
Run as `python tools/geodata.py <unpacked layers> <output folder>`: it writes
places.json.gz, countries.json.gz and shelves.json.gz for whichever of those
layers it finds unpacked, leaving the rest alone.

With `--labels` it writes only labels.json from the countries layer: each
country's ISO 3166-1 alpha-2 code to Natural Earth's label point for it, the
point the globe opens facing (REQUIREMENTS.md FR-GLB-015, FR-GLB-017). The
borders are left alone, so a label refresh never changes them.
"""

from __future__ import annotations

import gzip
import itertools
import json
import pathlib
import struct
import sys

SHAPE_POINT = 1
SHAPE_POLYGON = 5
SHP_HEADER_BYTES = 100
COORD_DECIMALS = 4


def read_dbf(path: pathlib.Path) -> list[dict[str, str]]:
    data = path.read_bytes()
    count, header_len, record_len = struct.unpack("<IHH", data[4:12])
    fields = []
    offset = 32
    while data[offset] != 0x0D:
        name = data[offset : offset + 11].split(b"\0", 1)[0].decode("ascii")
        length = data[offset + 16]
        fields.append((name, length))
        offset += 32
    rows = []
    for i in range(count):
        start = header_len + i * record_len + 1
        row = {}
        for name, length in fields:
            raw = data[start : start + length]
            row[name] = raw.decode("utf-8", "replace").strip(" \x00")
            start += length
        rows.append(row)
    return rows


def read_shp(path: pathlib.Path) -> list[object]:
    data = path.read_bytes()
    shapes: list[object] = []
    offset = SHP_HEADER_BYTES
    while offset < len(data):
        _, content_words = struct.unpack(">II", data[offset : offset + 8])
        body = data[offset + 8 : offset + 8 + content_words * 2]
        offset += 8 + content_words * 2
        kind = struct.unpack("<i", body[:4])[0]
        if kind == SHAPE_POINT:
            shapes.append(struct.unpack("<dd", body[4:20]))
        elif kind == SHAPE_POLYGON:
            num_parts, num_points = struct.unpack("<ii", body[36:44])
            parts = list(struct.unpack(f"<{num_parts}i", body[44 : 44 + 4 * num_parts]))
            pts_at = 44 + 4 * num_parts
            pts = struct.unpack(
                f"<{2 * num_points}d", body[pts_at : pts_at + 16 * num_points]
            )
            parts.append(num_points)
            rings = []
            for a, b in itertools.pairwise(parts):
                rings.append(
                    [
                        [
                            round(pts[2 * k], COORD_DECIMALS),
                            round(pts[2 * k + 1], COORD_DECIMALS),
                        ]
                        for k in range(a, b)
                    ]
                )
            shapes.append(rings)
        else:
            shapes.append(None)
    return shapes


def layer(src: pathlib.Path, name: str) -> pathlib.Path | None:
    """The shapefile stem for a Natural Earth layer unpacked under src; None when absent."""
    stem = src / name / name
    return stem if stem.with_suffix(".shp").exists() else None


def polygons(stem: pathlib.Path, name_field: str) -> list[dict[str, object]]:
    return [
        {"n": row[name_field], "r": rings}
        for row, rings in zip(
            read_dbf(stem.with_suffix(".dbf")), read_shp(stem.with_suffix(".shp"))
        )
        if rings
    ]


# NO_CODE is Natural Earth's mark for an area with no ISO code of its own.
NO_CODE = "-99"


def label_points(rows: list[dict[str, str]]) -> dict[str, list[float]]:
    """Each ISO_A2_EH code to [LABEL_Y, LABEL_X] (FR-GLB-017).

    A dependency carries its sovereign's code in ISO_A2_EH (the Coral Sea
    Islands read AU), so a code can name several rows. The row whose own
    ISO_A2 is the code wins; failing one (France and Norway hold -99 there),
    the row that is its own sovereign. A code still naming two rows stops the
    script rather than choosing silently.
    """
    by_code: dict[str, list[dict[str, str]]] = {}
    for row in rows:
        code = row["ISO_A2_EH"].strip()
        if code and code != NO_CODE:
            by_code.setdefault(code, []).append(row)
    table = {}
    for code, found in sorted(by_code.items()):
        if len(found) > 1:
            found = [r for r in found if r["ISO_A2"].strip() == code] or [
                r for r in found if r["ADMIN"].strip() == r["SOVEREIGNT"].strip()
            ]
        if len(found) != 1:
            names = ", ".join(r["ADMIN"].strip() for r in by_code[code])
            sys.exit(f"{code} names {len(found)} rows after the rule: {names}")
        row = found[0]
        table[code] = [
            round(float(row["LABEL_Y"]), COORD_DECIMALS),
            round(float(row["LABEL_X"]), COORD_DECIMALS),
        ]
    return table


def write_labels(src: pathlib.Path, out: pathlib.Path) -> None:
    cc = layer(src, "ne_10m_admin_0_countries")
    if not cc:
        sys.exit(f"ne_10m_admin_0_countries is not unpacked under {src}")
    table = label_points(read_dbf(cc.with_suffix(".dbf")))
    out.mkdir(parents=True, exist_ok=True)
    target = out / "labels.json"
    target.write_text(json.dumps(table, separators=(",", ":")) + "\n", encoding="utf-8")
    print(f"{target.name}: {len(table)} codes, {target.stat().st_size} bytes")


def main(src: pathlib.Path, out: pathlib.Path) -> None:
    """Write each layer whose source is unpacked under src, leaving the others alone.

    The ice shelves (ne_10m_antarctic_ice_shelves_polys) lie outside Antarctica's
    country polygon, which stops at the grounded coast; the geocoder counts them as
    Antarctica so a point on the Ross Ice Shelf does not read as sea.
    """
    out.mkdir(parents=True, exist_ok=True)
    layers: list[tuple[str, list[object]]] = []
    pp = layer(src, "ne_10m_populated_places_simple")
    if pp:
        places = []
        for row, pt in zip(
            read_dbf(pp.with_suffix(".dbf")), read_shp(pp.with_suffix(".shp"))
        ):
            places.append(
                {
                    "n": row["name"],
                    "r": row["adm1name"],
                    "c": row["adm0name"],
                    "lat": round(pt[1], COORD_DECIMALS),
                    "lng": round(pt[0], COORD_DECIMALS),
                }
            )
        layers.append(("places", places))
    cc = layer(src, "ne_10m_admin_0_countries")
    if cc:
        layers.append(("countries", polygons(cc, "NAME")))
    shelves = layer(src, "ne_10m_antarctic_ice_shelves_polys")
    if shelves:
        layers.append(("shelves", polygons(shelves, "name")))
    if not layers:
        sys.exit(f"no Natural Earth layer unpacked under {src}")
    for name, value in layers:
        target = out / f"{name}.json.gz"
        target.write_bytes(
            gzip.compress(
                json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode(
                    "utf-8"
                )
            )
        )
        print(f"{target.name}: {len(value)} items, {target.stat().st_size} bytes")


if __name__ == "__main__":
    if sys.argv[1] == "--labels":
        write_labels(pathlib.Path(sys.argv[2]), pathlib.Path(sys.argv[3]))
    else:
        main(pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2]))
