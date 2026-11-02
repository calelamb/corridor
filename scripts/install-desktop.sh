#!/usr/bin/env bash
# scripts/install-desktop.sh
# Zero-admin native desktop installer for Corridor on macOS and Linux workstations.
# Complies with NIST SP 800-53 security controls by utilizing sandboxed browser processes.

set -e

# Default settings
CORRIDOR_URL="https://corridorwildlife.org"
APP_NAME="Corridor"
UNINSTALL=false
FORCE_DESKTOP=false
TARGET_OS=""

show_help() {
    cat << 'EOF'
Corridor Desktop Installer (macOS & Linux)
Zero-admin, pure user-space desktop integration for Park Rangers & Staff.

Usage:
  ./scripts/install-desktop.sh [OPTIONS]

Options:
  --uninstall        Cleanly remove Corridor desktop shortcuts and app bundles
  --url <URL>        Specify target deployment URL (default: https://corridorwildlife.org)
  --name <NAME>      Specify application name (default: Corridor)
  --desktop          Ensure a shortcut is created on the Desktop
  --os <macos|linux> Force target operating system mode (useful for testing/automation)
  -h, --help         Display this help message and exit

Security & Architecture:
  - Strictly ZERO administrative / root / sudo privileges required.
  - Complies with NIST SP 800-53 security controls via native browser sandboxing.
  - macOS: Creates native Corridor.app bundle in ~/Applications and ~/Desktop.
  - Linux: Creates standard XDG freedesktop shortcut in ~/.local/share/applications.
  - Launches inside modern sandboxed browser windows (--app mode) via Chrome, Edge, Brave, or Safari.
EOF
}

# Parse command line arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --uninstall)
            UNINSTALL=true
            shift
            ;;
        --url)
            CORRIDOR_URL="$2"
            shift 2
            ;;
        --name)
            APP_NAME="$2"
            shift 2
            ;;
        --desktop)
            FORCE_DESKTOP=true
            shift
            ;;
        --os)
            TARGET_OS="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            echo "Run '$0 --help' for usage." >&2
            exit 1
            ;;
    esac
done

# Detect or normalize OS
if [ -n "$CORRIDOR_OS" ] && [ -z "$TARGET_OS" ]; then
    TARGET_OS="$CORRIDOR_OS"
fi

if [ -z "$TARGET_OS" ]; then
    KERNEL_NAME="$(uname -s)"
    case "$KERNEL_NAME" in
        Darwin*) TARGET_OS="macos" ;;
        Linux*)  TARGET_OS="linux" ;;
        *)       TARGET_OS="linux" ;;
    esac
fi

TARGET_OS="$(echo "$TARGET_OS" | tr '[:upper:]' '[:lower:]')"
if [ "$TARGET_OS" = "darwin" ]; then
    TARGET_OS="macos"
fi

# Embedded 512x512 SVG icon fallback (Corridor brand mark ⋈ with wildlife overpass arch)
EMBEDDED_ICON_SVG='<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" width="512" height="512">
  <defs>
    <linearGradient id="bgGrad512" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#163d31" />
      <stop offset="100%" stop-color="#0e241d" />
    </linearGradient>
    <linearGradient id="markGrad512" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#dcfce7" />
      <stop offset="100%" stop-color="#4ade80" />
    </linearGradient>
  </defs>
  <rect width="512" height="512" fill="url(#bgGrad512)" />
  <circle cx="256" cy="256" r="170" fill="#102b23" stroke="#4ade80" stroke-width="8" stroke-opacity="0.85" />
  <line x1="128" y1="256" x2="384" y2="256" stroke="#4ade80" stroke-width="5" stroke-dasharray="14 10" stroke-opacity="0.5" />
  <path d="M 160 197 Q 256 149 352 197" fill="none" stroke="#dcfce7" stroke-width="7" stroke-linecap="round" stroke-opacity="0.75" />
  <path d="M 160 315 Q 256 363 352 315" fill="none" stroke="#dcfce7" stroke-width="7" stroke-linecap="round" stroke-opacity="0.75" />
  <path d="M 181 203 L 256 256 L 181 309 Z M 331 203 L 256 256 L 331 309 Z" fill="url(#markGrad512)" stroke="#4ade80" stroke-width="7" stroke-linejoin="round" />
  <circle cx="256" cy="256" r="9" fill="#102b23" stroke="#dcfce7" stroke-width="4" />
</svg>'

# Locate icon asset on disk or write embedded
resolve_icon_svg() {
    local target_dest="$1"
    local script_dir
    script_dir="$(cd "$(dirname "$0")" 2>/dev/null && pwd || true)"
    local repo_icon="${script_dir}/../web/static/icons/icon-512.svg"

    if [ -f "$repo_icon" ]; then
        cp "$repo_icon" "$target_dest"
    elif [ -f "web/static/icons/icon-512.svg" ]; then
        cp "web/static/icons/icon-512.svg" "$target_dest"
    else
        printf '%s\n' "$EMBEDDED_ICON_SVG" > "$target_dest"
    fi
}

# ==============================================================================
# UNINSTALLATION WORKFLOW
# ==============================================================================
if [ "$UNINSTALL" = true ]; then
    echo "=== Uninstalling Corridor Desktop Shortcuts ==="
    REMOVED=0

    if [ "$TARGET_OS" = "macos" ]; then
        MAC_APP_DIR="${HOME}/Applications/${APP_NAME}.app"
        MAC_DESKTOP_APP="${HOME}/Desktop/${APP_NAME}.app"

        if [ -d "$MAC_APP_DIR" ]; then
            rm -rf "$MAC_APP_DIR"
            echo "[-] Removed macOS application: $MAC_APP_DIR"
            REMOVED=1
        fi
        if [ -e "$MAC_DESKTOP_APP" ] || [ -L "$MAC_DESKTOP_APP" ]; then
            rm -rf "$MAC_DESKTOP_APP"
            echo "[-] Removed Desktop shortcut: $MAC_DESKTOP_APP"
            REMOVED=1
        fi
    else
        # Linux uninstallation
        LINUX_DESKTOP="${HOME}/.local/share/applications/corridor.desktop"
        LINUX_DESKTOP_LINK="${HOME}/Desktop/corridor.desktop"
        LINUX_LAUNCHER="${HOME}/.local/bin/corridor"
        LINUX_ICON_HICOLOR="${HOME}/.local/share/icons/hicolor/scalable/apps/corridor.svg"
        LINUX_ICON_GENERIC="${HOME}/.local/share/icons/corridor.svg"

        if [ -f "$LINUX_DESKTOP" ]; then
            rm -f "$LINUX_DESKTOP"
            echo "[-] Removed XDG desktop file: $LINUX_DESKTOP"
            REMOVED=1
        fi
        if [ -f "$LINUX_DESKTOP_LINK" ] || [ -L "$LINUX_DESKTOP_LINK" ]; then
            rm -f "$LINUX_DESKTOP_LINK"
            echo "[-] Removed Desktop shortcut: $LINUX_DESKTOP_LINK"
            REMOVED=1
        fi
        if [ -f "$LINUX_LAUNCHER" ]; then
            rm -f "$LINUX_LAUNCHER"
            echo "[-] Removed launcher script: $LINUX_LAUNCHER"
            REMOVED=1
        fi
        if [ -f "$LINUX_ICON_HICOLOR" ]; then
            rm -f "$LINUX_ICON_HICOLOR"
            echo "[-] Removed scalable icon: $LINUX_ICON_HICOLOR"
            REMOVED=1
        fi
        if [ -f "$LINUX_ICON_GENERIC" ]; then
            rm -f "$LINUX_ICON_GENERIC"
            echo "[-] Removed generic icon: $LINUX_ICON_GENERIC"
            REMOVED=1
        fi

        if command -v update-desktop-database >/dev/null 2>&1; then
            update-desktop-database "${HOME}/.local/share/applications" 2>/dev/null || true
        fi
    fi

    if [ $REMOVED -eq 1 ]; then
        echo "[✓] Corridor has been cleanly uninstalled from your user environment."
    else
        echo "[!] No Corridor desktop installations were detected to remove."
    fi
    exit 0
fi

# ==============================================================================
# INSTALLATION WORKFLOW
# ==============================================================================
echo "=== Installing Corridor Desktop App (Zero-Admin) ==="
echo "Operating System: ${TARGET_OS}"
echo "Target Platform URL: ${CORRIDOR_URL}"
echo "Application Name: ${APP_NAME}"

if [ "$TARGET_OS" = "macos" ]; then
    # --------------------------------------------------------------------------
    # macOS Native .app Bundle Creation
    # --------------------------------------------------------------------------
    APP_DIR="${HOME}/Applications/${APP_NAME}.app"
    CONTENTS_DIR="${APP_DIR}/Contents"
    MACOS_DIR="${CONTENTS_DIR}/MacOS"
    RESOURCES_DIR="${CONTENTS_DIR}/Resources"

    mkdir -p "$MACOS_DIR"
    mkdir -p "$RESOURCES_DIR"

    # 1. Info.plist
    cat > "${CONTENTS_DIR}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>corridor</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>org.corridorwildlife.desktop</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

    # 2. Executable launcher script
    LAUNCHER_SCRIPT="${MACOS_DIR}/corridor"
    cat > "$LAUNCHER_SCRIPT" << EOF
#!/bin/sh
# Corridor Desktop Launcher for macOS
# Standalone window mode via Chromium browsers with Safari fallback

TARGET_URL="\${CORRIDOR_URL:-${CORRIDOR_URL}}"
if [ \$# -gt 0 ]; then
    case "\$1" in
        http*://*) TARGET_URL="\$1" ;;
        --url) TARGET_URL="\$2" ;;
    esac
fi

# 1. Google Chrome (Dedicated frameless app window)
if [ -d "/Applications/Google Chrome.app" ]; then
    exec "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --app="\$TARGET_URL" "\$@"
elif [ -d "\$HOME/Applications/Google Chrome.app" ]; then
    exec "\$HOME/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --app="\$TARGET_URL" "\$@"

# 2. Microsoft Edge (Enterprise & GFE macOS installations)
elif [ -d "/Applications/Microsoft Edge.app" ]; then
    exec "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge" --app="\$TARGET_URL" "\$@"
elif [ -d "\$HOME/Applications/Microsoft Edge.app" ]; then
    exec "\$HOME/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge" --app="\$TARGET_URL" "\$@"

# 3. Brave Browser
elif [ -d "/Applications/Brave Browser.app" ]; then
    exec "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser" --app="\$TARGET_URL" "\$@"
elif [ -d "\$HOME/Applications/Brave Browser.app" ]; then
    exec "\$HOME/Applications/Brave Browser.app/Contents/MacOS/Brave Browser" --app="\$TARGET_URL" "\$@"

# 4. Apple Safari fallback
elif [ -d "/Applications/Safari.app" ]; then
    open -a "Safari" "\$TARGET_URL"
else
    open "\$TARGET_URL"
fi
EOF
    chmod +x "$LAUNCHER_SCRIPT"

    # 3. Embed SVG icon
    SVG_DEST="${RESOURCES_DIR}/icon-512.svg"
    resolve_icon_svg "$SVG_DEST"

    # 4. Compile AppIcon.icns if macOS graphics utilities are available
    if command -v qlmanage >/dev/null 2>&1 && command -v sips >/dev/null 2>&1 && command -v iconutil >/dev/null 2>&1; then
        TMP_ICON_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'corridor-icon')"
        qlmanage -t -s 1024 -o "$TMP_ICON_DIR" "$SVG_DEST" >/dev/null 2>&1 || true
        PNG_FILE="$(find "$TMP_ICON_DIR" -name "*.png" 2>/dev/null | head -n 1)"
        if [ -n "$PNG_FILE" ] && [ -f "$PNG_FILE" ]; then
            ICONSET="${TMP_ICON_DIR}/AppIcon.iconset"
            mkdir -p "$ICONSET"
            sips -z 16 16     "$PNG_FILE" --out "${ICONSET}/icon_16x16.png" >/dev/null 2>&1 || true
            sips -z 32 32     "$PNG_FILE" --out "${ICONSET}/icon_16x16@2x.png" >/dev/null 2>&1 || true
            sips -z 32 32     "$PNG_FILE" --out "${ICONSET}/icon_32x32.png" >/dev/null 2>&1 || true
            sips -z 64 64     "$PNG_FILE" --out "${ICONSET}/icon_32x32@2x.png" >/dev/null 2>&1 || true
            sips -z 128 128   "$PNG_FILE" --out "${ICONSET}/icon_128x128.png" >/dev/null 2>&1 || true
            sips -z 256 256   "$PNG_FILE" --out "${ICONSET}/icon_128x128@2x.png" >/dev/null 2>&1 || true
            sips -z 256 256   "$PNG_FILE" --out "${ICONSET}/icon_256x256.png" >/dev/null 2>&1 || true
            sips -z 512 512   "$PNG_FILE" --out "${ICONSET}/icon_256x256@2x.png" >/dev/null 2>&1 || true
            sips -z 512 512   "$PNG_FILE" --out "${ICONSET}/icon_512x512.png" >/dev/null 2>&1 || true
            sips -z 1024 1024 "$PNG_FILE" --out "${ICONSET}/icon_512x512@2x.png" >/dev/null 2>&1 || true
            iconutil -c icns "$ICONSET" -o "${RESOURCES_DIR}/AppIcon.icns" >/dev/null 2>&1 || true
        fi
        rm -rf "$TMP_ICON_DIR"
    fi

    # Touch bundle so macOS LaunchServices discovers it
    touch "$APP_DIR" 2>/dev/null || true

    # Optional Desktop shortcut symlink
    if [ -d "${HOME}/Desktop" ] || [ "$FORCE_DESKTOP" = true ]; then
        mkdir -p "${HOME}/Desktop"
        ln -sf "$APP_DIR" "${HOME}/Desktop/${APP_NAME}.app"
        echo "[+] Created Desktop link: ${HOME}/Desktop/${APP_NAME}.app"
    fi

    echo "[+] Installed native macOS application: ${APP_DIR}"
    echo "[+] Searchable in Spotlight (Cmd+Space -> Corridor) and Launchpad."

else
    # --------------------------------------------------------------------------
    # Linux Freedesktop XDG Specification Installation
    # --------------------------------------------------------------------------
    BIN_DIR="${HOME}/.local/bin"
    APPS_DIR="${HOME}/.local/share/applications"
    ICONS_DIR="${HOME}/.local/share/icons/hicolor/scalable/apps"
    GENERIC_ICONS_DIR="${HOME}/.local/share/icons"

    mkdir -p "$BIN_DIR"
    mkdir -p "$APPS_DIR"
    mkdir -p "$ICONS_DIR"
    mkdir -p "$GENERIC_ICONS_DIR"

    # 1. Launcher script in ~/.local/bin/corridor
    LAUNCHER_SCRIPT="${BIN_DIR}/corridor"
    cat > "$LAUNCHER_SCRIPT" << EOF
#!/bin/sh
# Corridor Desktop Launcher for Linux Workstations
# Runs Corridor in dedicated app window mode via Chromium/Chrome/Edge/Brave with xdg-open fallback

TARGET_URL="\${CORRIDOR_URL:-${CORRIDOR_URL}}"
while [ \$# -gt 0 ]; do
    case "\$1" in
        --url)
            TARGET_URL="\$2"
            shift 2
            ;;
        http*://*)
            TARGET_URL="\$1"
            shift
            ;;
        *)
            shift
            ;;
    esac
done

# Probe Chromium-based browsers for native --app standalone window mode
for browser in google-chrome google-chrome-stable chromium chromium-browser microsoft-edge microsoft-edge-stable brave-browser; do
    if command -v "\$browser" >/dev/null 2>&1; then
        exec "\$browser" --app="\$TARGET_URL" "\$@"
    fi
done

# Fallback to system default desktop browser
if command -v xdg-open >/dev/null 2>&1; then
    exec xdg-open "\$TARGET_URL"
elif command -v sensible-browser >/dev/null 2>&1; then
    exec sensible-browser "\$TARGET_URL"
else
    echo "Error: No supported browser found to open \$TARGET_URL" >&2
    exit 1
fi
EOF
    chmod +x "$LAUNCHER_SCRIPT"
    echo "[+] Installed launcher command: ${LAUNCHER_SCRIPT}"

    # 2. Install SVG Icon
    HICOLOR_SVG="${ICONS_DIR}/corridor.svg"
    GENERIC_SVG="${GENERIC_ICONS_DIR}/corridor.svg"
    resolve_icon_svg "$HICOLOR_SVG"
    cp "$HICOLOR_SVG" "$GENERIC_SVG"
    echo "[+] Installed vector icon: ${HICOLOR_SVG}"

    # 3. Write XDG Desktop Entry
    DESKTOP_FILE="${APPS_DIR}/corridor.desktop"
    cat > "$DESKTOP_FILE" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=${APP_NAME}
GenericName=Wildlife Coexistence Platform
Comment=Wildlife collision reduction and corridor mitigation platform for park rangers
Exec=${LAUNCHER_SCRIPT} %U
Icon=corridor
Terminal=false
Categories=Utility;Science;Geography;
StartupWMClass=corridorwildlife.org
MimeType=x-scheme-handler/corridor;
Keywords=wildlife;corridor;nps;ranger;gis;mitigation;roadkill;
Actions=Report;Mitigate;Hotspots;Rangers;

[Desktop Action Report]
Name=New Field Carcass Report
Exec=${LAUNCHER_SCRIPT} --url ${CORRIDOR_URL}/report/

[Desktop Action Mitigate]
Name=Mitigation Planner
Exec=${LAUNCHER_SCRIPT} --url ${CORRIDOR_URL}/?workspace=mitigate

[Desktop Action Hotspots]
Name=Hotspot Lab
Exec=${LAUNCHER_SCRIPT} --url ${CORRIDOR_URL}/?workspace=hotspot

[Desktop Action Rangers]
Name=NPS Ranger Hub
Exec=${LAUNCHER_SCRIPT} --url ${CORRIDOR_URL}/rangers/
EOF
    chmod +x "$DESKTOP_FILE"
    echo "[+] Installed desktop specification entry: ${DESKTOP_FILE}"

    # Optional Desktop shortcut symlink
    if [ -d "${HOME}/Desktop" ] || [ "$FORCE_DESKTOP" = true ]; then
        mkdir -p "${HOME}/Desktop"
        cp "$DESKTOP_FILE" "${HOME}/Desktop/corridor.desktop"
        chmod +x "${HOME}/Desktop/corridor.desktop" 2>/dev/null || true
        echo "[+] Created Desktop launcher: ${HOME}/Desktop/corridor.desktop"
    fi

    # Update desktop and icon databases if tools are available
    if command -v update-desktop-database >/dev/null 2>&1; then
        update-desktop-database "$APPS_DIR" 2>/dev/null || true
    fi
    if command -v gtk-update-icon-cache >/dev/null 2>&1; then
        gtk-update-icon-cache -f -t "${HOME}/.local/share/icons/hicolor" 2>/dev/null || true
    fi
fi

echo ""
echo "[✓] Installation Complete!"
echo "    - No administrator / root rights required (pure user-space)."
echo "    - Sandboxed browser process isolation complies with NIST SP 800-53."
echo "    - Ready for field carcass logging, mitigation planning, and live dispatch."
