#!/usr/bin/env python3
# Copyright (C) 2026 Aliaksandr Shynkevich
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program. If not, see <https://www.gnu.org/licenses/>.

"""
genicons.py — generates TimeSyncer icon files from a programmatic clock drawing.

Outputs:
  assets/icon.png       — 32×32 tray icon (committed, embedded via go:embed)
  res/timesyncer.ico    — Windows multi-size icon (16, 32, 48, 256 px)
  res/timesyncer.icns   — macOS icon bundle (via iconutil on macOS)

Requires: Pillow  (pip3 install Pillow)
"""

import math
import os
import struct
import subprocess
import sys
import tempfile
from io import BytesIO
from pathlib import Path

try:
    from PIL import Image, ImageDraw
except ImportError:
    sys.exit(
        "genicons: Pillow is required.\n"
        "  pip3 install Pillow"
    )

# Resolve project root regardless of cwd: tools/genicons/genicons.py → ../../
ROOT   = Path(__file__).resolve().parent.parent.parent
ASSETS = ROOT / "assets"
RES    = ROOT / "res"


# ── Clock renderer ────────────────────────────────────────────────────────────

def render(size: int) -> Image.Image:
    """Render a black-and-white clock icon at the given pixel size."""
    img  = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    cx       = size / 2.0
    cy       = size / 2.0
    radius   = cx * 0.88
    border_w = max(1.5, radius * 0.22)
    line_w   = max(1.5, radius * 0.18)

    # White filled clock face
    draw.ellipse(
        [cx - radius, cy - radius, cx + radius, cy + radius],
        fill=(255, 255, 255, 255),
    )
    # Black border ring
    draw.ellipse(
        [cx - radius, cy - radius, cx + radius, cy + radius],
        outline=(0, 0, 0, 255),
        width=int(round(border_w)),
    )
    # Minute hand — 12 o'clock (straight up)
    draw.line(
        [(cx, cy), (cx, cy - radius * 0.60)],
        fill=(0, 0, 0, 255),
        width=int(round(line_w)),
    )
    # Hour hand — ~10 o'clock
    angle = 2 * math.pi * (10.0 / 12.0) - math.pi / 2
    draw.line(
        [(cx, cy),
         (cx + radius * 0.40 * math.cos(angle),
          cy + radius * 0.40 * math.sin(angle))],
        fill=(0, 0, 0, 255),
        width=int(round(line_w)),
    )
    return img


def to_png(img: Image.Image) -> bytes:
    buf = BytesIO()
    img.save(buf, format="PNG")
    return buf.getvalue()


# ── ICO (Windows) ─────────────────────────────────────────────────────────────

def write_ico(path: Path) -> None:
    write_ico_sizes(path, [16, 32, 48, 256])


def write_ico_sizes(path: Path, sizes: list) -> None:
    pngs = [to_png(render(s)) for s in sizes]

    with open(path, "wb") as f:
        # ICONDIR header
        f.write(struct.pack("<HHH", 0, 1, len(sizes)))
        # ICONDIRENTRY × n
        offset = 6 + len(sizes) * 16
        for s, png in zip(sizes, pngs):
            w = 0 if s == 256 else s   # 0 encodes 256 in ICO format
            f.write(struct.pack("<BBBBHHII", w, w, 0, 0, 1, 32, len(png), offset))
            offset += len(png)
        for png in pngs:
            f.write(png)


# ── ICNS (macOS) ──────────────────────────────────────────────────────────────

def write_icns_iconutil(path: Path) -> None:
    """Use Apple's iconutil — produces a correctly structured ICNS with @2x
    Retina entries so list-view icons render correctly at small sizes."""
    entries = [
        ("icon_16x16.png",      16),
        ("icon_16x16@2x.png",   32),   # 16pt @2x
        ("icon_32x32.png",      32),
        ("icon_32x32@2x.png",   64),   # 32pt @2x
        ("icon_128x128.png",   128),
        ("icon_128x128@2x.png",256),   # 128pt @2x
        ("icon_256x256.png",   256),
        ("icon_256x256@2x.png",512),   # 256pt @2x
        ("icon_512x512.png",   512),
    ]
    with tempfile.TemporaryDirectory(suffix=".iconset") as tmp:
        for name, size in entries:
            (Path(tmp) / name).write_bytes(to_png(render(size)))
        result = subprocess.run(
            ["iconutil", "-c", "icns", "--output", str(path), tmp],
            capture_output=True, text=True,
        )
        if result.returncode != 0:
            sys.exit(f"genicons: iconutil failed:\n{result.stderr}")


def write_icns_manual(path: Path) -> None:
    """Portable hand-written ICNS for non-macOS hosts (e.g. Linux CI).
    Produces a valid file but without @2x entries."""
    entries = [
        (b"icp4",  16),
        (b"icp5",  32),
        (b"ic07", 128),
        (b"ic08", 256),
        (b"ic09", 512),
    ]
    chunks = [(code, to_png(render(size))) for code, size in entries]
    total = 8 + sum(8 + len(data) for _, data in chunks)

    with open(path, "wb") as f:
        f.write(b"icns")
        f.write(struct.pack(">I", total))
        for code, data in chunks:
            f.write(code)
            f.write(struct.pack(">I", 8 + len(data)))
            f.write(data)


def write_icns(path: Path) -> None:
    if sys.platform == "darwin":
        write_icns_iconutil(path)
    else:
        write_icns_manual(path)


# ── Entry point ───────────────────────────────────────────────────────────────

def main() -> None:
    ASSETS.mkdir(parents=True, exist_ok=True)
    RES.mkdir(parents=True, exist_ok=True)

    tray_png = ASSETS / "icon.png"
    tray_png.write_bytes(to_png(render(32)))
    print("assets/icon.png")

    # Small single-size ICO for the Windows system tray (systray expects ICO format).
    # Committed to git so `go build` works without running this script.
    tray_ico = ASSETS / "icon.ico"
    write_ico_sizes(tray_ico, [32])
    print("assets/icon.ico")

    write_ico(RES / "timesyncer.ico")
    print("res/timesyncer.ico")

    write_icns(RES / "timesyncer.icns")
    print("res/timesyncer.icns")


if __name__ == "__main__":
    main()
