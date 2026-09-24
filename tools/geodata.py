"""Convert Natural Earth shapefiles to the compact form that
internal/infrastructure/geo embeds. First written for the Phase 0 spike.

Reads the ESRI shapefile (.shp, polygon and point records) and its dBase
table (.dbf) directly from their published layouts, so no package is needed.
Run as `python tools/geodata.py <unpacked layers> <output folder>`: it writes
places.json.gz, countries.json.gz and shelves.json.gz for whichever of those
layers it finds unpacked, leaving the rest alone.
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
    main(pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2]))
