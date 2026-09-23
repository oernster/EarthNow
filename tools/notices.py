"""Write THIRD_PARTY_NOTICES from the dependencies actually shipped (NFR-LEG-001).

The Go modules are those linked into the application and the setup program
(go list -deps); the page packages are the page's production tree
(npm ls --omit=dev --all). Each component is listed with its licence, then every
licence text follows in full, read from the component's own licence file: MIT,
BSD and ISC all ask for the text itself to travel with the binary.

    python tools/notices.py           rewrite the file
    python tools/notices.py --check   exit 1 when the file is not what would be written

Run from anywhere; paths are taken from this file's place in the repository.
"""

from __future__ import annotations

import json
import pathlib
import re
import shutil
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
NOTICES = ROOT / "THIRD_PARTY_NOTICES"
FRONTEND = ROOT / "frontend"
GO_PACKAGES = (".", "./installer")
LICENCE_FILE = re.compile(
    r"^(licen[cs]e|copying)(\.(md|txt|markdown))?$", re.IGNORECASE
)
NAME_COLUMN = 43
RULE = "=" * 79

HEADER = """\
EarthNow: third-party notices

EarthNow's own code is licensed under the GNU General Public License, version 3
(see LICENSE). It ships the components and data below, each under its own
licence or terms. The lists are the modules linked into the application and the
setup program (go list -deps) and the page's production dependency tree
(npm ls --omit=dev --all). The full text of every licence follows the lists.
This file is written by tools/notices.py; edit that, not this.


DATA AND IMAGERY BUNDLED IN THE APPLICATION

Earth imagery: NASA Earth Observatory, Blue Marble Next Generation
  Used under the NASA media usage guidelines:
  https://www.nasa.gov/nasa-brand-center/images-and-media/
  Credited as NASA Earth Observatory. EarthNow is not affiliated with NASA,
  NASA does not endorse it and it carries no NASA insignia.

Place names and country boundaries: Natural Earth
  Public domain. Terms of use:
  https://www.naturalearthdata.com/about/terms-of-use/


DATA FETCHED WHILE RUNNING (not bundled; a local copy is cached)

Natural events: NASA Earth Observatory Natural Event Tracker (EONET)
  https://eonet.gsfc.nasa.gov/
Earthquakes: USGS Earthquake Hazards Program
  https://earthquake.usgs.gov/
Volcanoes: Smithsonian / USGS Weekly Volcanic Activity Report
  Global Volcanism Program, Smithsonian Institution: https://volcano.si.edu/
  with the USGS Volcano Hazards Program.
  A product of employees of the United States; used for non-commercial purposes
  under the Smithsonian Terms of Use, which ask for this citation:
  https://volcano.si.edu/gvp_termsofuse.cfm
  Neither NASA, the USGS nor the Smithsonian endorses EarthNow.
"""

FOOTER = """
The MIT, ISC, BSD, Apache-2.0, Unlicense and 0BSD licences all permit shipping
inside a GPL-3.0 application.
"""

# How a Go module's licence is named from its text, which Go modules do not
# state anywhere else. Checked in order; the first phrase found names it.
GO_LICENCES = (
    ("Apache License", "Apache-2.0"),
    ("Permission is hereby granted, free of charge", "MIT"),
    ("Neither the name", "BSD-3-Clause"),
    ("Redistribution and use in source and binary forms", "BSD-2-Clause"),
)


class NoticeError(Exception):
    """A component whose licence cannot be found or named."""


class Component:
    def __init__(self, name: str, version: str, licence: str, text: str) -> None:
        self.name = name
        self.version = version
        self.licence = licence
        self.text = text

    def row(self) -> str:
        label = f"{self.name} {self.version}"
        return f"{label.ljust(NAME_COLUMN - 1)} {self.licence}".rstrip()


def run(args: list[str], cwd: pathlib.Path) -> str:
    # npm is a .cmd script on Windows, which only a resolved path can start.
    program = shutil.which(args[0])
    if program is None:
        raise NoticeError(f"{args[0]} is not on the path")
    done = subprocess.run(
        [program, *args[1:]],
        cwd=cwd,
        capture_output=True,
        text=True,
        encoding="utf-8",
        check=False,
    )
    if done.returncode != 0:
        raise NoticeError(f"{' '.join(args)} failed: {done.stderr.strip()}")
    return done.stdout


def licence_text(folder: pathlib.Path, name: str) -> str:
    found = sorted(p for p in folder.iterdir() if LICENCE_FILE.match(p.name))
    if not found:
        raise NoticeError(f"{name} has no licence file in {folder}")
    text = found[0].read_text(encoding="utf-8", errors="replace")
    return text.replace("\r\n", "\n").strip() + "\n"


def go_licence(text: str, name: str) -> str:
    for phrase, licence in GO_LICENCES:
        if phrase in text:
            return licence
    raise NoticeError(f"the licence of {name} is not one this tool can name")


def go_components() -> list[Component]:
    goroot = pathlib.Path(run(["go", "env", "GOROOT"], ROOT).strip())
    version = run(["go", "env", "GOVERSION"], ROOT).strip()
    stdlib = licence_text(goroot, "the Go standard library")
    found = [Component("The Go standard library", version, "BSD-3-Clause", stdlib)]
    template = (
        "{{with .Module}}{{if not .Main}}"
        "{{.Path}}\t{{.Version}}\t{{.Dir}}"
        "{{end}}{{end}}"
    )
    listed = run(["go", "list", "-deps", "-f", template, *GO_PACKAGES], ROOT)
    for line in sorted(set(filter(None, listed.splitlines()))):
        path, version, folder = line.split("\t")
        text = licence_text(pathlib.Path(folder), path)
        found.append(Component(path, version, go_licence(text, path), text))
    return found


def page_components() -> list[Component]:
    listed = run(
        ["npm", "ls", "--omit=dev", "--all", "--parseable", "--long"], FRONTEND
    )
    seen: dict[str, Component] = {}
    for line in listed.splitlines()[1:]:
        folder, _, spec = line.rpartition(":")
        name, _, version = spec.rpartition("@")
        if f"{name}@{version}" in seen:
            continue
        where = pathlib.Path(folder)
        manifest = json.loads((where / "package.json").read_text(encoding="utf-8"))
        licence = manifest.get("license")
        if not isinstance(licence, str):
            raise NoticeError(f"{name} {version} states no licence in its package.json")
        seen[f"{name}@{version}"] = Component(
            name, version, licence, licence_text(where, name)
        )
    return sorted(seen.values(), key=lambda c: c.name)


def render(go: list[Component], page: list[Component]) -> str:
    parts = [HEADER, "", "GO MODULES (the application and the setup program)", ""]
    parts += [c.row() for c in go]
    parts += ["", "", "PAGE PACKAGES (bundled into the application's page)", ""]
    parts += [c.row() for c in page]
    parts += [FOOTER, "", "LICENCE TEXTS", ""]
    for c in go + page:
        parts += [RULE, f"{c.name} {c.version} ({c.licence})", RULE, "", c.text]
    return "\n".join(parts).rstrip() + "\n"


def main(argv: list[str]) -> int:
    try:
        go, page = go_components(), page_components()
        wanted = render(go, page)
    except NoticeError as refused:
        print(f"notices: {refused}", file=sys.stderr)
        return 2
    if "--check" in argv:
        held = NOTICES.read_text(encoding="utf-8") if NOTICES.exists() else ""
        if held != wanted:
            print(
                "THIRD_PARTY_NOTICES is not what the dependencies call for; "
                "run python tools/notices.py",
                file=sys.stderr,
            )
            return 1
        return 0
    NOTICES.write_text(wanted, encoding="utf-8", newline="\n")
    print(f"wrote {NOTICES.name}: {len(go)} Go modules, {len(page)} page packages")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
