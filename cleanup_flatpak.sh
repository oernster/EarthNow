#!/usr/bin/env bash
# Uninstalls and purges the EarthNow Flatpak (DEL-006). Run from the repo root:
#
#   bash cleanup_flatpak.sh
#
# Scoped to flatpak artefacts only. It deliberately does NOT touch what the other build
# paths produce (build/bin, dist-installer, dist-dmg), so the three stay independent. It
# never touches the user's settings or cached events either: forgetting those is something
# the user asks for, not something a cleanup script decides.
set -euo pipefail

APP_ID="uk.codecrafter.EarthNow"
BIN_NAME="earthnow"
APP_NAME="EarthNow"

bold=$(tput bold 2>/dev/null || true)
reset=$(tput sgr0 2>/dev/null || true)
section() { echo; echo "${bold}=== $* ===${reset}"; }

section "Uninstalling ${APP_ID}"
if flatpak list --user | grep -q "${APP_ID}"; then
    flatpak uninstall --user -y "${APP_ID}"
    echo "  Uninstalled."
else
    echo "  Not installed, skipping."
fi

section "Removing flatpak build artefacts"
rm -f "${BIN_NAME}.flatpak"
rm -rf .flatpak-build .flatpak-repo .flatpak-builder
rm -f "${APP_ID}.yml"
rm -rf packaging/
# build_flatpak.sh renders the icons here; no other build path writes build/linux.
rm -rf build/linux
echo "  Done."

echo
echo "${bold}Purge complete.${reset}"
echo "Your settings and cached events were left alone: ~/.var/app/${APP_ID}/cache/${APP_NAME}"
