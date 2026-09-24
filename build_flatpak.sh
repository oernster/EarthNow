#!/usr/bin/env bash
# Builds the EarthNow Flatpak for Linux (DEL-005, release 2). Run from the repo root:
#
#   bash build_flatpak.sh
#
# Flow: install flatpak tooling and Pillow if missing, add flathub, install the GNOME runtime
# and the golang and node SDK extensions, render the hicolor icons, self-generate the desktop
# file, metainfo and manifest, build the app inside the sandbox (npm front end, then a CGO Go
# build against the runtime's webkit2gtk-4.1), install it for the current user and export a
# distributable bundle.
#
# The GNOME runtime is required because Wails v2 renders through webkit2gtk-4.1, which the
# freedesktop runtime does not carry. The Go build therefore uses -tags webkit2_41.
#
# Ported from PigeonPost's build_flatpak.sh, with SymDiary's version stamp and licence. The
# differences, each deliberate:
#
#   1. There is no `wails generate module` step. EarthNow's wire contract is the hand-written
#      frontend/src/api.ts and types.ts; the generated bindings are gitignored and imported
#      by nothing, so nothing in the sandbox needs them.
#   2. The icons come from tools/genicons.py on the host, not from a generator inside the
#      sandbox: Pillow is on the host and not in the SDK. It writes into the gitignored
#      build/linux/icons; the manifest installs whatever it wrote, so the sizes have one
#      home, in genicons.py.
#   3. The finished application keeps --share=network, since fetching three public sources and
#      the cloud image is the whole of what it does (NFR-PRIV-001). It needs no --filesystem at all: its
#      settings, cache and log live under the sandbox's own cache folder.
#   4. The globe needs WebGL. main.go passes options.Linux with the GPU policy set to Always,
#      since Wails v2 turns acceleration off when that option is absent (RSK-002). Whether the
#      webkit2gtk in the runtime then offers WebGL2 on a given machine is measured there, not
#      promised here; where it does not, the page says so rather than showing a blank window.
#
# Outputs: earthnow.flatpak (installable anywhere) and a user install of ${APP_ID}.
set -euo pipefail

APP_ID="uk.codecrafter.EarthNow"
BIN_NAME="earthnow"
APP_NAME="EarthNow"
APP_SUMMARY="See what is happening on Earth right now"
HOMEPAGE="https://earthnow.world/"
RUNTIME="org.gnome.Platform"
SDK="org.gnome.Sdk"
RUNTIME_VERSION="50"
SDK_EXT_VERSION="25.08"   # the freedesktop base of GNOME 50; extensions pair with it
GOLANG_EXT="org.freedesktop.Sdk.Extension.golang"
NODE_EXT="org.freedesktop.Sdk.Extension.node22"
BUILD_DIR=".flatpak-build"
REPO_DIR=".flatpak-repo"
BUNDLE="${BIN_NAME}.flatpak"
MANIFEST="${APP_ID}.yml"
PACKAGING_DIR="packaging"
ICON_DIR="build/linux/icons"
VERSION="$(tr -d '[:space:]' < VERSION)"
RELEASE_DATE="$(date +%F)"

section() { printf '\n\033[1m== %s ==\033[0m\n' "$1"; }

install_if_missing() {
    local tool="$1"
    command -v "$tool" > /dev/null 2>&1 && return 0
    section "Installing missing tool: $tool"
    if command -v apt-get > /dev/null 2>&1; then sudo apt-get install -y "$tool"
    elif command -v dnf > /dev/null 2>&1; then sudo dnf install -y "$tool"
    elif command -v pacman > /dev/null 2>&1; then sudo pacman -S --noconfirm "$tool"
    elif command -v zypper > /dev/null 2>&1; then sudo zypper install -y "$tool"
    else echo "error: install $tool with your package manager and re-run" >&2; exit 1
    fi
}

# install_pillow puts Pillow where the system python3 can import it. Each distribution
# names the package differently; a distribution python also refuses pip installs into
# itself, so the package manager is asked rather than pip.
install_pillow() {
    python3 -c "import PIL" > /dev/null 2>&1 && return 0
    section "Installing missing Python library: Pillow"
    if command -v apt-get > /dev/null 2>&1; then sudo apt-get install -y python3-pil
    elif command -v dnf > /dev/null 2>&1; then sudo dnf install -y python3-pillow
    elif command -v pacman > /dev/null 2>&1; then sudo pacman -S --noconfirm python-pillow
    elif command -v zypper > /dev/null 2>&1; then sudo zypper install -y python3-Pillow
    else echo "error: install Pillow for python3 with your package manager and re-run" >&2; exit 1
    fi
}

section "Tooling"
install_if_missing flatpak
install_if_missing flatpak-builder
install_if_missing python3
install_pillow

section "Flathub remote and runtimes"
flatpak remote-add --if-not-exists --user flathub https://dl.flathub.org/repo/flathub.flatpakrepo
flatpak install --user --noninteractive flathub \
    "${RUNTIME}//${RUNTIME_VERSION}" \
    "${SDK}//${RUNTIME_VERSION}" \
    "${GOLANG_EXT}//${SDK_EXT_VERSION}" \
    "${NODE_EXT}//${SDK_EXT_VERSION}"

section "Rendering the hicolor icons"
# Rendered fresh every run, like the icon build.ps1 copies into build/ for Windows, so a
# changed master can never ship beside an icon made from the old one.
rm -rf "${ICON_DIR}"
python3 tools/genicons.py --hicolor "${ICON_DIR}"

section "Writing packaging files"
mkdir -p "${PACKAGING_DIR}"

# No MimeType is claimed. EarthNow opens no file of any kind, so a file manager handing it
# one is not a route the application has.
cat > "${PACKAGING_DIR}/${APP_ID}.desktop" << DESKTOP
[Desktop Entry]
Name=${APP_NAME}
Comment=${APP_SUMMARY}
Exec=${BIN_NAME}
Icon=${APP_ID}
Terminal=false
Type=Application
Categories=Education;Science;Geography;
DESKTOP

cat > "${PACKAGING_DIR}/${APP_ID}.metainfo.xml" << METAINFO
<?xml version="1.0" encoding="UTF-8"?>
<component type="desktop-application">
  <id>${APP_ID}</id>
  <name>${APP_NAME}</name>
  <summary>${APP_SUMMARY}</summary>
  <metadata_license>CC0-1.0</metadata_license>
  <project_license>GPL-3.0-only</project_license>
  <description>
    <p>
      A slowly turning globe carrying the natural events of the past week at the places
      they happened: earthquakes from the USGS, wildfires, storms, floods and ice from
      NASA EONET and the volcanoes in the Smithsonian and USGS Weekly Volcanic Activity
      Report. Every event names the source that reported it and how old that report is.
      The clouds from EUMETSAT's world cloud map can be laid over the globe.
      No account, no telemetry and no forecasts.
    </p>
  </description>
  <launchable type="desktop-id">${APP_ID}.desktop</launchable>
  <url type="homepage">${HOMEPAGE}</url>
  <content_rating type="oars-1.1"/>
  <releases>
    <release version="${VERSION}" date="${RELEASE_DATE}"/>
  </releases>
</component>
METAINFO

section "Writing manifest"
ICON_INSTALL_CMDS=""
for icon in "${ICON_DIR}"/${BIN_NAME}_*.png; do
    size="$(basename "${icon}" .png)"
    size="${size#${BIN_NAME}_}"
    ICON_INSTALL_CMDS="${ICON_INSTALL_CMDS}      - install -Dm644 ${icon} /app/share/icons/hicolor/${size}x${size}/apps/${APP_ID}.png
"
done
[ -n "${ICON_INSTALL_CMDS}" ] || { echo "error: no icons in ${ICON_DIR}" >&2; exit 1; }

cat > "${MANIFEST}" << MANIFEST_EOF
app-id: ${APP_ID}
runtime: ${RUNTIME}
runtime-version: '${RUNTIME_VERSION}'
sdk: ${SDK}
sdk-extensions:
  - ${GOLANG_EXT}
  - ${NODE_EXT}
command: ${BIN_NAME}
# finish-args is the permission the FINISHED application runs with: a display, a GPU for the
# globe and the network for its three sources and the cloud image. No filesystem permission: its settings, cache
# and log live in the sandbox's own cache folder. Opening a source page or the donate page
# hands the address to the desktop portal, so that needs nothing either.
finish-args:
  - --share=ipc
  - --socket=fallback-x11
  - --socket=wayland
  - --device=dri
  - --share=network
build-options:
  append-path: /usr/lib/sdk/golang/bin:/usr/lib/sdk/node22/bin
  build-args:
    - --share=network
  env:
    GOPATH: /run/build/${BIN_NAME}/gopath
    GOCACHE: /run/build/${BIN_NAME}/gocache
    GOFLAGS: -buildvcs=false
    npm_config_cache: /run/build/${BIN_NAME}/npm-cache
modules:
  - name: ${BIN_NAME}
    buildsystem: simple
    build-commands:
      # npm install rather than npm ci, matching wails.json frontend:install, which is what
      # every other platform runs. npm ci aborts whenever package-lock.json drifts even
      # slightly from package.json, for example over esbuild's optional platform packages,
      # and that strict check is why a Linux build can break while Windows and macOS keep
      # working. Keep this in lockstep with wails.json.
      # NOTE: this heredoc is unquoted, so backticks and dollar-parens here are shell
      # evaluated into the manifest. Never put either in these comments.
      - cd frontend && npm install --no-audit --no-fund && npm run build
      # The version reaches the binary the same way build.ps1 sends it, through -X against
      # a var: a const would be silently ignored.
      - go build -tags desktop,production,webkit2_41 -ldflags "-s -w -X main.appVersion=${VERSION}" -o ${BIN_NAME} .
      - install -Dm755 ${BIN_NAME} /app/bin/${BIN_NAME}
      - chmod -R u+w gopath gocache 2>/dev/null || true
      - install -Dm644 packaging/${APP_ID}.desktop /app/share/applications/${APP_ID}.desktop
      - install -Dm644 packaging/${APP_ID}.metainfo.xml /app/share/metainfo/${APP_ID}.metainfo.xml
      # Section 4 of the GNU GPL asks that a copy of the licence reach every recipient
      # along with the program, so it travels inside the bundle.
      - install -Dm644 LICENSE /app/share/licenses/${APP_ID}/LICENSE
${ICON_INSTALL_CMDS}    sources:
      - type: dir
        path: .
        skip:
          - .git
          - ${BUILD_DIR}
          - ${REPO_DIR}
          - .flatpak-builder
          - ${BUNDLE}
          - frontend/node_modules
          - spike
          - spike-assets
          - dist-installer
MANIFEST_EOF

section "Building ${APP_NAME} ${VERSION} with flatpak-builder"
flatpak-builder --user --install --force-clean \
    --install-deps-from=flathub \
    --repo="${REPO_DIR}" \
    "${BUILD_DIR}" "${MANIFEST}"

section "Exporting bundle"
flatpak build-bundle \
    --runtime-repo=https://dl.flathub.org/repo/flathub.flatpakrepo \
    "${REPO_DIR}" "${BUNDLE}" "${APP_ID}"

section "Done"
echo "Installed for the current user: flatpak run ${APP_ID}"
echo "Distributable bundle: ${BUNDLE}"
echo "Settings, cache and log live under ~/.var/app/${APP_ID}/cache/${APP_NAME}"
