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

The application icon: assets/application-icon.png becomes a multi-size Windows
.ico beside it (DEL-004, Appendix D.1). That one file is the whole identity:
build.ps1 places it on both executables; the shortcuts and the taskbar button
read it out of the binary rather than carrying a copy of their own.

The setup program's artwork: its header mark (256 px) from the application
icon plus its theme toggle's sun and moon (128 px) from light-mode.png and
dark-mode.png, written straight into installer/frontend/dist. That page has
no bundler, so it loads each file as it finds it; shipping a master there
would put a megabyte and more behind one badge.

The donate mark: cropped to its artwork and scaled by height to four times the
rail glyph, never squared, since a wide picture on a square canvas spends its
height on nothing (FR-DON-002). The same render goes to the site (FR-DON-010).

The GitHub Pages site under docs/ also takes the application's mark as its icon
and a small copy of the page's NASA texture for its turning globe (CON-009).

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

# APP_MASTER is the application's own identity rather than a rail icon.
APP_MASTER = "application-icon.png"

# ICO_SIZES are the sizes Windows chooses between: the small tray and menu sizes,
# the taskbar and shortcut sizes, then the large one Explorer uses in its biggest
# view. Leaving one out makes Windows scale a neighbour, which looks soft.
ICO_SIZES = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]

# HEADER_SIZE is the setup window's header mark and SETUP_ICON_SIZE its theme
# button (Appendix D.1). Each is roughly two and a half times the size it is drawn
# at, crisp on a high-density display without carrying detail nothing shows.
HEADER_SIZE = 256
SETUP_ICON_SIZE = 128

SETUP = REPO / "installer" / "frontend" / "dist"
HEADER = SETUP / "icon.png"

# The donate mark (FR-DON-002, Appendix D.1) is a wide picture, not an icon, so
# it is cropped to its artwork and scaled by height alone rather than squared.
# DONATE_HEIGHT is four times the rail's 48 px glyph, crisp under display scaling.
DONATE_MASTER = "donate.png"
RAIL_GLYPH_PX = 48
DONATE_HEIGHT = 4 * RAIL_GLYPH_PX

# SITE is the GitHub Pages site. It has no build step, so it carries its own
# small copies of the artwork, written here beside the application's.
SITE = REPO / "docs"

# DONATE_OUTPUTS receive the same render in one loop, so no copy can drift from
# another: the rail's button and the site's (FR-DON-010).
DONATE_OUTPUTS = (
    REPO / "frontend" / "src" / "assets" / DONATE_MASTER,
    SITE / DONATE_MASTER,
)

# APP_MARK_OUTPUTS are the application's mark at the rail icon size: beside the
# heading in the key column and as the site's icon.
APP_MARK_OUTPUTS = (OUTPUT / APP_MASTER, SITE / "icon.png")

# The site's globe turns the application's own NASA texture (CON-009: Earth
# imagery is observational data, never drawn), scaled down to twice the height
# the site draws it at: its orb is at most SITE_ORB_PX across (docs/styles.css).
EARTH_TEXTURE = REPO / "frontend" / "src" / "assets" / "earth.jpg"
SITE_EARTH = SITE / "earth.jpg"
SITE_ORB_PX = 380
SITE_EARTH_HEIGHT = 2 * SITE_ORB_PX
SITE_EARTH_QUALITY = 85

# SETUP_ICONS are the theme toggle's pair; each shows the mode it switches TO
# (NFR-UX-004), so the sun shows while the page is dark.
SETUP_ICONS = ("light-mode.png", "dark-mode.png")


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


def render_image(image: Image.Image, target: pathlib.Path, size: int = SIZE) -> int:
    """Write one downscaled icon from artwork in memory; return its byte size."""
    box = image.getbbox()
    art = image.crop(box) if box is not None else image
    inner = size - 2 * PAD
    scale = min(inner / art.width, inner / art.height)
    scaled = art.resize(
        (max(1, round(art.width * scale)), max(1, round(art.height * scale))),
        Image.LANCZOS,
    )
    canvas = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    canvas.paste(
        scaled,
        ((size - scaled.width) // 2, (size - scaled.height) // 2),
        scaled,
    )
    canvas.save(target, "PNG", optimize=True)
    return target.stat().st_size


def trimmed(master: pathlib.Path) -> Image.Image:
    """Open a master and crop away the transparent margin around its artwork."""
    image = Image.open(master).convert("RGBA")
    box = image.getbbox()
    return image.crop(box) if box is not None else image


def squared(image: Image.Image) -> Image.Image:
    """Centre artwork on a transparent square, padding once rather than stretching."""
    side = max(image.width, image.height)
    square = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    square.paste(image, ((side - image.width) // 2, (side - image.height) // 2), image)
    return square


def render_ico(master: pathlib.Path, target: pathlib.Path) -> int:
    """Write the multi-size Windows icon; return its byte size.

    It is squared before saving: an .ico entry is square by definition, so a
    source that is not would be stretched into every size rather than padded once.
    """
    squared(trimmed(master)).save(target, "ICO", sizes=ICO_SIZES)
    return target.stat().st_size


def render_setup() -> None:
    """Write the application .ico and the setup program's three pictures."""
    app = MASTERS / APP_MASTER
    if not app.exists():
        sys.exit(f"\nno application icon at {app}")
    ico = app.with_suffix(".ico")
    written = render_ico(app, ico)
    source = app.stat().st_size
    print(f"\n{app.name:<22} {source:>9,} -> {written:>7,} bytes  ({ico.name})")

    SETUP.mkdir(parents=True, exist_ok=True)
    squared(trimmed(app)).resize((HEADER_SIZE, HEADER_SIZE), Image.LANCZOS).save(
        HEADER, "PNG", optimize=True
    )
    print(f"{'':<22} {'':>9} -> {HEADER.stat().st_size:>7,} bytes  (setup header)")

    for name in SETUP_ICONS:
        master = MASTERS / name
        if not master.exists():
            sys.exit(f"no {name} in {MASTERS}; the setup theme toggle wears it")
        target = SETUP / name
        written = render_image(
            Image.open(master).convert("RGBA"), target, SETUP_ICON_SIZE
        )
        source = master.stat().st_size
        print(f"{name:<22} {source:>9,} -> {written:>7,} bytes  (setup {name})")


def render_donate() -> None:
    """Write the donate mark to every destination from one render."""
    master = MASTERS / DONATE_MASTER
    if not master.exists():
        sys.exit(f"no {DONATE_MASTER} in {MASTERS}; the donate button wears it")
    art = trimmed(master)
    width = max(1, round(art.width * DONATE_HEIGHT / art.height))
    mark = art.resize((width, DONATE_HEIGHT), Image.LANCZOS)
    for target in DONATE_OUTPUTS:
        target.parent.mkdir(parents=True, exist_ok=True)
        mark.save(target, "PNG", optimize=True)
        source, size = master.stat().st_size, target.stat().st_size
        drawn = f"{width} x {DONATE_HEIGHT}"
        print(f"{DONATE_MASTER:<22} {source:>9,} -> {size:>7,} bytes  ({drawn})")


def render_site_earth() -> None:
    """Write the site's copy of the NASA texture, keeping its aspect ratio."""
    if not EARTH_TEXTURE.exists():
        sys.exit(f"no {EARTH_TEXTURE.name} beside the page; the site's globe wears it")
    texture = Image.open(EARTH_TEXTURE).convert("RGB")
    width = round(texture.width * SITE_EARTH_HEIGHT / texture.height)
    SITE.mkdir(parents=True, exist_ok=True)
    texture.resize((width, SITE_EARTH_HEIGHT), Image.LANCZOS).save(
        SITE_EARTH, "JPEG", quality=SITE_EARTH_QUALITY, optimize=True
    )
    source, size = EARTH_TEXTURE.stat().st_size, SITE_EARTH.stat().st_size
    drawn = f"{width} x {SITE_EARTH_HEIGHT}"
    print(f"{EARTH_TEXTURE.name:<22} {source:>9,} -> {size:>7,} bytes  ({drawn}, site)")


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

    render_setup()
    render_donate()
    render_site_earth()

    # The front end wants the application icon too, beside the heading in the key
    # column (and as the About crest to come), as the reference writes it. The
    # site wears the same mark, rendered by the same call.
    app = Image.open(MASTERS / APP_MASTER).convert("RGBA")
    for target in APP_MARK_OUTPUTS:
        target.parent.mkdir(parents=True, exist_ok=True)
        written = render_image(app, target)
        where = target.relative_to(REPO).as_posix()
        print(f"{APP_MASTER:<22} {'':>9} -> {written:>7,} bytes  ({where})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
