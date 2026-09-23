"""Generate the rail icons from the master artwork in assets/.

Ported from ED Voyage Companion's tools/genicons.py.

The masters are 1254 pixels square and a megabyte or two each, which is right
for artwork and wrong for a rail that draws them at 48. Each is trimmed to its
own content then centred on a square canvas, so every icon carries the same
optical weight, then written small enough to embed.

The stop-rotating icon has no master. It is DERIVED here by laying negative.png
over rotate.png, so the two states of the rotation button cannot drift apart:
every pixel the overlay does not cover is the rotate artwork's own pixel
(REQUIREMENTS.md Appendix D.2, NFR-UX-004).

Run it when a master changes:

    python tools/genicons.py

It is not part of the build. The output is committed, so a clone needs neither
Python nor Pillow to build the application.
"""

from __future__ import annotations

import pathlib
import sys

try:
    from PIL import Image
except ImportError:  # pragma: no cover - a plain message beats a traceback
    sys.exit("Pillow is required: python -m pip install pillow")

# SIZE is Appendix D.2's generated size: over four times the rail's 48 px glyph,
# so the artwork stays crisp on a high-density display.
SIZE = 208

# PAD keeps the trimmed artwork off the edge of its square.
PAD = 2

# ROTATE_MASTER is the rotation artwork and OVERLAY_MASTER the mark laid over it
# to make STOP_ICON. The overlay is never shown alone, so it is not rendered.
ROTATE_MASTER = "rotate.png"
OVERLAY_MASTER = "negative.png"
STOP_ICON = "rotate-stop.png"

# NOT_RAIL are masters with another destination: the application icon (the .ico
# and the About crest, Phase 4), the donate mark (cropped by height, FR-DON) and
# the setup program's theme pair.
NOT_RAIL = (
    "application-icon.png",
    OVERLAY_MASTER,
    "donate.png",
    "light-mode.png",
    "dark-mode.png",
)

REPO = pathlib.Path(__file__).resolve().parent.parent
MASTERS = REPO / "assets"
OUTPUT = REPO / "frontend" / "src" / "assets" / "icons"


def overlaid(base_path: pathlib.Path, overlay_path: pathlib.Path) -> Image.Image:
    """Return the base artwork with the overlay fitted inside its bounding box.

    Nothing underneath is touched, so the stop state is the rotate state plus a
    mark rather than a second drawing. Fitting the overlay to the artwork's own
    box keeps both states on one canvas AND one bounding box, so the trim in
    render_image treats them alike.
    """
    base = Image.open(base_path).convert("RGBA")
    mark = Image.open(overlay_path).convert("RGBA")
    box = mark.getbbox()
    if box is not None:
        mark = mark.crop(box)
    left, top, right, bottom = base.getbbox() or (0, 0, base.width, base.height)
    width, height = right - left, bottom - top
    fit = min(width / mark.width, height / mark.height)
    mark = mark.resize(
        (max(1, round(mark.width * fit)), max(1, round(mark.height * fit))),
        Image.LANCZOS,
    )
    out = base.copy()
    out.alpha_composite(
        mark,
        (left + (width - mark.width) // 2, top + (height - mark.height) // 2),
    )
    return out


def render_image(image: Image.Image, target: pathlib.Path) -> int:
    """Write one downscaled icon from artwork in memory; return its byte size."""
    box = image.getbbox()
    art = image.crop(box) if box is not None else image
    inner = SIZE - 2 * PAD
    scale = min(inner / art.width, inner / art.height)
    scaled = art.resize(
        (max(1, round(art.width * scale)), max(1, round(art.height * scale))),
        Image.LANCZOS,
    )
    canvas = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    canvas.paste(
        scaled,
        ((SIZE - scaled.width) // 2, (SIZE - scaled.height) // 2),
        scaled,
    )
    canvas.save(target, "PNG", optimize=True)
    return target.stat().st_size


def main() -> int:
    masters = sorted(p for p in MASTERS.glob("*.png") if p.name not in NOT_RAIL)
    if not masters:
        sys.exit(f"no master artwork found in {MASTERS}")
    OUTPUT.mkdir(parents=True, exist_ok=True)

    total_in = total_out = 0
    for master in masters:
        source = master.stat().st_size
        written = render_image(Image.open(master).convert("RGBA"), OUTPUT / master.name)
        total_in += source
        total_out += written
        print(f"{master.name:<22} {source:>9,} -> {written:>7,} bytes")

    rotate = MASTERS / ROTATE_MASTER
    overlay = MASTERS / OVERLAY_MASTER
    for needed in (rotate, overlay):
        if not needed.exists():
            sys.exit(f"no {needed.name} in {MASTERS}; the stop icon is made from it")
    total_in += overlay.stat().st_size
    written = render_image(overlaid(rotate, overlay), OUTPUT / STOP_ICON)
    total_out += written
    print(f"{STOP_ICON:<22} {'derived':>9} -> {written:>7,} bytes")

    print(f"\n{len(masters) + 1} rail icons, {total_in:,} -> {total_out:,} bytes")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
