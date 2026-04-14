#!/usr/bin/env bash
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

set -euo pipefail

APP="timesyncer"
OUT_DIR="dist"

# Read version from git tag, fallback to "dev"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

TARGETS=(
  "windows amd64 .exe"
  "windows arm64 .exe"
  "linux   amd64 "
  "linux   arm64 "
  "linux   arm   "
  "darwin  amd64 "
  "darwin  arm64 "
)

# ── Resource generation ──────────────────────────────────────────────────────

# generate_resources runs the Python icon generator to produce:
#   assets/icon.png   — committed tray icon (go:embed)
#   res/timesyncer.ico / res/timesyncer.icns — packaging icons
# Requires: pip3 install Pillow
generate_resources() {
  echo "Generating icon resources…"
  python3 tools/genicons/genicons.py
}

# generate_windows_syso embeds the manifest + icon + version info into a .syso
# file that the Go toolchain picks up when building the Windows binary.
generate_windows_syso() {
  local goarch="${1:-amd64}"
  local arch_flag
  case "${goarch}" in
    amd64) arch_flag="-64" ;;
    arm64) arch_flag="-arm" ;;
    *)     arch_flag="" ;;
  esac
  echo "Generating Windows resource file (${goarch}; manifest + icon + version info)…"
  go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest \
    ${arch_flag} \
    -o timesyncer_windows.syso \
    res/versioninfo.json
}

# create_app_bundle wraps the darwin binary in a proper .app bundle so the
# clock icon appears in Finder. The raw binary is kept alongside.
create_app_bundle() {
  local goarch="$1"
  local platform_dir="${OUT_DIR}/darwin-${goarch}"
  local binary="${platform_dir}/timesyncer"
  local app="${platform_dir}/TimeSyncer.app"

  mkdir -p "${app}/Contents/MacOS"
  mkdir -p "${app}/Contents/Resources"

  cp "${binary}" "${app}/Contents/MacOS/timesyncer"
  chmod +x "${app}/Contents/MacOS/timesyncer"

  cat > "${app}/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleExecutable</key>         <string>timesyncer</string>
  <key>CFBundleIdentifier</key>         <string>com.alex-and-soft.timesyncer</string>
  <key>CFBundleName</key>               <string>TimeSyncer</string>
  <key>CFBundleDisplayName</key>        <string>TimeSyncer</string>
  <key>CFBundleVersion</key>            <string>${VERSION}</string>
  <key>CFBundleShortVersionString</key> <string>${VERSION}</string>
  <key>CFBundlePackageType</key>        <string>APPL</string>
  <key>CFBundleIconFile</key>           <string>AppIcon</string>
  <key>NSHumanReadableCopyright</key>   <string>Copyright © 2026 Aliaksandr Shynkevich. Licensed under GPL v3.</string>
  <key>LSUIElement</key>                <true/>
  <key>NSHighResolutionCapable</key>    <true/>
  <key>LSMinimumSystemVersion</key>     <string>12.0</string>
</dict></plist>
PLIST

  if [[ -f "res/timesyncer.icns" ]]; then
    cp "res/timesyncer.icns" "${app}/Contents/Resources/AppIcon.icns"
  fi
}

usage() {
  echo "Usage: $0 [platform]"
  echo ""
  echo "Platforms:"
  echo "  all          Build for all platforms (default)"
  echo "  windows      Build for Windows amd64 + arm64"
  echo "  linux        Build for Linux amd64 + arm64 + arm"
  echo "  darwin       Build for macOS amd64 + arm64"
  echo "  windows/amd64, linux/amd64, darwin/arm64, ...  Single target"
  echo ""
  echo "Examples:"
  echo "  $0            # build all"
  echo "  $0 windows    # build Windows only"
  echo "  $0 linux/arm64"
}

build_linux_docker() {
  local goarch="$1"
  local platform_dir="${OUT_DIR}/linux-${goarch}"
  local out="${platform_dir}/${APP}"

  printf "  %-30s" "linux/${goarch}"

  if ! command -v docker &>/dev/null; then
    echo "SKIPPED (Docker not found — install Rancher Desktop or Docker Desktop)"
    return 0
  fi

  # Verify the Docker daemon is actually running before calling docker info
  if ! docker info &>/dev/null; then
    echo "SKIPPED (Docker daemon not running — start Rancher Desktop or Docker Desktop)"
    return 0
  fi

  # Detect the real Docker daemon architecture (not the Go toolchain, which may run via Rosetta)
  local docker_arch
  docker_arch=$(docker info --format '{{.Architecture}}' 2>/dev/null | sed 's/x86_64/amd64/;s/aarch64/arm64/')

  # If target arch differs from docker arch, QEMU emulation is required
  if [[ "${goarch}" != "${docker_arch}" ]]; then
    # Quick check: can Docker actually run the target platform?
    if ! docker run --rm --platform "linux/${goarch}" --entrypoint uname alpine:latest -m &>/dev/null; then
      echo "SKIPPED (QEMU emulation for linux/${goarch} not enabled)"
      echo ""
      echo "        Your Docker daemon runs on ${docker_arch}."
      echo "        To build linux/${goarch} from this machine, enable QEMU:"
      echo "        Rancher Desktop → Preferences → Virtual Machine → Emulation → QEMU"
      return 0
    fi
  fi

  mkdir -p "${platform_dir}"

  docker build \
    --platform "linux/${goarch}" \
    --build-arg VERSION="${VERSION}" \
    --build-arg GOARCH="${goarch}" \
    -f Dockerfile.linux \
    --output "type=local,dest=${OUT_DIR}/__linux_tmp_${goarch}" \
    . 2>&1

  local bin
  bin=$(find "${OUT_DIR}/__linux_tmp_${goarch}" -type f | head -1)
  if [[ -n "${bin}" ]]; then
    mv "${bin}" "${out}"
    chmod +x "${out}"
    rm -rf "${OUT_DIR}/__linux_tmp_${goarch}"
    cp scripts/install_linux.sh "${platform_dir}/install.sh"
    cp scripts/run_linux.sh "${platform_dir}/run.sh"
    chmod +x "${platform_dir}/install.sh" "${platform_dir}/run.sh"
    local size
    size=$(du -sh "${out}" | cut -f1)
    echo "→ ${platform_dir}/ (${size})"
  else
    echo "FAILED (check Docker output above)"
    return 1
  fi
}

build_one() {
  local goos="$1"
  local goarch="$2"
  local ext="${3:-}"

  local platform_dir="${OUT_DIR}/${goos}-${goarch}"
  local out="${platform_dir}/${APP}${ext}"
  local host_os
  host_os=$(go env GOOS)

  # Windows: embed manifest + icon into .syso before go build
  if [[ "${goos}" == "windows" ]]; then
    generate_windows_syso "${goarch}"
  fi

  printf "  %-30s" "${goos}/${goarch}${ext}"

  # Linux systray requires full CGo + GTK headers — cannot cross-compile from non-Linux
  if [[ "${goos}" == "linux" && "${host_os}" != "linux" ]]; then
    build_linux_docker "${goarch}"
    return
  fi

  mkdir -p "${platform_dir}"

  (
    export GOOS="${goos}"
    export GOARCH="${goarch}"

    if [[ "${goos}" == "darwin" ]]; then
      # macOS always needs CGo (AppKit/Objective-C via systray)
      export CGO_ENABLED=1
      if [[ "${goarch}" == "arm64" && "${host_os}" == "darwin" ]]; then
        export CC="clang"
        export CFLAGS="-arch arm64"
        export LDFLAGS="-arch arm64"
      fi
    elif [[ "${goos}" == "windows" ]]; then
      # Windows systray uses Win32 API via syscall — no CGo required
      export CGO_ENABLED=0
    elif [[ "${goos}" == "linux" ]]; then
      export CGO_ENABLED=1
    fi

    local ldflags="-s -w -X main.version=${VERSION}"
    if [[ "${goos}" == "windows" ]]; then
      ldflags="${ldflags} -H windowsgui"
    fi

    go build \
      -trimpath \
      -ldflags="${ldflags}" \
      -o "${out}" \
      . 2>&1
  )

  # macOS: wrap binary in .app bundle (icon shown in Finder) then remove the
  # raw binary — users drag TimeSyncer.app to /Applications; autostart is
  # handled by the in-app toggle (writes its own LaunchAgent via os.Executable).
  if [[ "${goos}" == "darwin" ]]; then
    create_app_bundle "${goarch}"
    rm -f "${out}"
  fi

  # Copy platform install script (Linux + Windows only; macOS uses .app)
  case "${goos}" in
    linux)
      cp scripts/install_linux.sh "${platform_dir}/install.sh"
      cp scripts/run_linux.sh "${platform_dir}/run.sh"
      chmod +x "${platform_dir}/install.sh" "${platform_dir}/run.sh" ;;
    windows)
      cp scripts/install_windows.ps1 "${platform_dir}/install.ps1" ;;
  esac

  local size
  if [[ "${goos}" == "darwin" ]]; then
    size=$(du -sh "${platform_dir}/TimeSyncer.app" | cut -f1)
  else
    size=$(du -sh "${out}" | cut -f1)
  fi
  echo "→ ${platform_dir}/ (${size})"
}

main() {
  local filter="${1:-all}"

  echo "TimeSyncer build — version: ${VERSION}"
  echo "Output directory: ${OUT_DIR}/"
  echo ""

  mkdir -p "${OUT_DIR}"

  generate_resources

  for target in "${TARGETS[@]}"; do
    read -r goos goarch ext <<< "${target}"
    goarch="${goarch// /}"
    ext="${ext// /}"

    case "${filter}" in
      all)                        ;;
      windows|linux|darwin)
        [[ "${goos}" == "${filter}" ]] || continue ;;
      */*)
        [[ "${goos}/${goarch}" == "${filter}" ]] || continue ;;
      *)
        echo "Unknown platform: ${filter}"; usage; exit 1 ;;
    esac

    build_one "${goos}" "${goarch}" "${ext}"
  done

  echo ""
  echo "Done. Artifacts in ${OUT_DIR}/:"
  for d in "${OUT_DIR}"/*/; do
    [[ -d "$d" ]] || continue
    echo "  $d"
    ls "$d" | sed 's/^/      /'
  done
}

case "${1:-}" in
  -h|--help) usage; exit 0 ;;
  *)         main "${1:-all}" ;;
esac
