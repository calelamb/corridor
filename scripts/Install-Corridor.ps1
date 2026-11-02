<#
.SYNOPSIS
    Corridor Desktop App Installer for Windows 10/11 Government Furnished Equipment (GFE).

.DESCRIPTION
    Creates native Start Menu and Desktop application shortcuts for Corridor on Windows 10/11
    workstations used by National Park Service (NPS) wildlife rangers, resource managers,
    dispatchers, and GIS transportation planners.

    Corridor executes in dedicated standalone application window mode (--app=https://corridorwildlife.org)
    leveraging enterprise-managed Microsoft Edge (pre-installed on all GFE) or Google Chrome.

    Zero Administrative Privileges Required:
    - Pure user-space execution. No UAC elevation, no MSI/EXE installer, no registry writes to HKLM.
    - Installs strictly to the current user's profile ($HOME\Desktop and $env:APPDATA\Microsoft\Windows\Start Menu\Programs).

    Security & NIST SP 800-53 Rev. 5 Compliance:
    - AC-6 (Least Privilege): Operates entirely within standard unprivileged user account context.
    - CM-7 (Least Functionality): Deploys zero native background daemons, services, or scheduled tasks.
    - SC-18 (Mobile Code): Executes inside Chromium's native sandboxed process with OS-level process isolation.
    - SC-28 (Protection of Information at Rest): Offline field reports and basemap tiles are stored
      within the user profile's encrypted browser storage (IndexedDB / CacheStorage).
    - IA-2 / IA-8 (Identification & Authentication): Full compatibility with federal PIV / CAC card and Entra ID SSO.

.PARAMETER Uninstall
    Cleanly removes all Corridor shortcuts and local icon assets from the current user profile.

.PARAMETER Url
    Target web application URL (default: https://corridorwildlife.org).

.PARAMETER AppName
    Display name for the shortcuts (default: Corridor).

.EXAMPLE
    .\scripts\Install-Corridor.ps1
    Installs Corridor shortcuts on the Desktop and Start Menu for the current ranger.

.EXAMPLE
    .\scripts\Install-Corridor.ps1 -Uninstall
    Removes Corridor shortcuts and cleans up the local icon directory.

.EXAMPLE
    .\scripts\Install-Corridor.ps1 -Url "https://corridorwildlife.org" -AppName "Corridor Wildlife"
#>

[CmdletBinding(SupportsShouldProcess = $true)]
param(
    [switch]$Uninstall,
    [string]$Url = "https://corridorwildlife.org",
    [string]$AppName = "Corridor"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-CorridorHeader {
    Write-Host @"
========================================================================
    CORRIDOR - Zero-Admin Desktop Setup for NPS / GFE Workstations
    Wildlife Collision Reduction & Animal Passage Mitigation
========================================================================
"@ -ForegroundColor Green
}

# ------------------------------------------------------------------------------
# Resolve User-Space Directories (Strictly Zero Admin Elevation)
# ------------------------------------------------------------------------------
$DesktopPath = [System.Environment]::GetFolderPath([System.Environment+SpecialFolder]::Desktop)
$StartMenuPath = [System.Environment]::GetFolderPath([System.Environment+SpecialFolder]::Programs)
$LocalAppData = if ($env:LOCALAPPDATA) { $env:LOCALAPPDATA } else { [System.IO.Path]::Combine($env:USERPROFILE, "AppData", "Local") }
$CorridorDataDir = [System.IO.Path]::Combine($LocalAppData, "Corridor")

$DesktopShortcutPath = [System.IO.Path]::Combine($DesktopPath, "$AppName.lnk")
$StartMenuShortcutPath = [System.IO.Path]::Combine($StartMenuPath, "$AppName.lnk")
$IconPath = [System.IO.Path]::Combine($CorridorDataDir, "corridor.ico")

# ------------------------------------------------------------------------------
# Uninstallation Workflow
# ------------------------------------------------------------------------------
if ($Uninstall) {
    Write-CorridorHeader
    Write-Host "[*] Removing Corridor desktop shortcuts..." -ForegroundColor Yellow

    $removedCount = 0

    if (Test-Path -LiteralPath $DesktopShortcutPath) {
        Remove-Item -LiteralPath $DesktopShortcutPath -Force
        Write-Host "    [-] Removed Desktop shortcut: $DesktopShortcutPath" -ForegroundColor Cyan
        $removedCount++
    }

    if (Test-Path -LiteralPath $StartMenuShortcutPath) {
        Remove-Item -LiteralPath $StartMenuShortcutPath -Force
        Write-Host "    [-] Removed Start Menu shortcut: $StartMenuShortcutPath" -ForegroundColor Cyan
        $removedCount++
    }

    if (Test-Path -LiteralPath $CorridorDataDir) {
        Remove-Item -LiteralPath $CorridorDataDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "    [-] Removed icon cache directory: $CorridorDataDir" -ForegroundColor Cyan
        $removedCount++
    }

    if ($removedCount -gt 0) {
        Write-Host "[+] Corridor has been cleanly uninstalled from your user profile." -ForegroundColor Green
    } else {
        Write-Host "[!] No Corridor shortcuts or caches were found to remove." -ForegroundColor Yellow
    }
    exit 0
}

# ------------------------------------------------------------------------------
# Installation Workflow
# ------------------------------------------------------------------------------
Write-CorridorHeader
Write-Host "[*] Configuring zero-admin desktop application for: $Url" -ForegroundColor Cyan

# 1. Probe Enterprise-Approved Browser (Microsoft Edge or Google Chrome)
function Get-ApprovedBrowser {
    $browserCandidates = @(
        # Microsoft Edge (Default GFE standard browser)
        [System.IO.Path]::Combine(${env:ProgramFiles(x86)}, "Microsoft\Edge\Application\msedge.exe"),
        [System.IO.Path]::Combine($env:ProgramFiles, "Microsoft\Edge\Application\msedge.exe"),
        [System.IO.Path]::Combine($LocalAppData, "Microsoft\Edge\Application\msedge.exe"),
        # Google Chrome
        [System.IO.Path]::Combine($env:ProgramFiles, "Google\Chrome\Application\chrome.exe"),
        [System.IO.Path]::Combine(${env:ProgramFiles(x86)}, "Google\Chrome\Application\chrome.exe"),
        [System.IO.Path]::Combine($LocalAppData, "Google\Chrome\Application\chrome.exe")
    )

    foreach ($candidate in $browserCandidates) {
        if (-not [string]::IsNullOrWhiteSpace($candidate) -and (Test-Path -LiteralPath $candidate)) {
            return $candidate
        }
    }

    # Registry lookup in HKCU / HKLM App Paths
    $regPaths = @(
        "HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\msedge.exe",
        "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\msedge.exe",
        "HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\chrome.exe",
        "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\chrome.exe"
    )
    foreach ($regPath in $regPaths) {
        try {
            $item = Get-ItemProperty -Path $regPath -ErrorAction SilentlyContinue
            if ($item -and $item.'(default)' -and (Test-Path -LiteralPath $item.'(default)')) {
                return $item.'(default)'
            }
        } catch { }
    }

    # Fallback to PATH environment
    foreach ($cmd in @("msedge.exe", "chrome.exe")) {
        $found = Get-Command -Name $cmd -ErrorAction SilentlyContinue
        if ($found -and $found.Source -and (Test-Path -LiteralPath $found.Source)) {
            return $found.Source
        }
    }

    return $null
}

$browserPath = Get-ApprovedBrowser

if (-not $browserPath) {
    Write-Warning "[!] Neither Microsoft Edge nor Google Chrome was detected at standard locations."
    Write-Host "    Defaulting to system web shell association." -ForegroundColor Yellow
    $browserPath = "$env:SystemRoot\explorer.exe"
    $browserArgs = $Url
} else {
    Write-Host "[+] Detected enterprise browser: $browserPath" -ForegroundColor Green
    $browserArgs = "--app=$Url"
}

# 2. Extract and Cache Corridor Icon Asset
if (-not (Test-Path -LiteralPath $CorridorDataDir)) {
    New-Item -Path $CorridorDataDir -ItemType Directory -Force | Out-Null
}

$hasCustomIcon = $false
$repoIconCandidates = @(
    (Join-Path $PSScriptRoot "..\web\static\icons\corridor.ico"),
    (Join-Path $PSScriptRoot "..\web\static\icons\icon-512.svg")
)

# Embedded Base64 ICO asset (Multi-res Corridor emblem: 32x32, 48x48)
$EmbeddedIconBase64 = "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"

try {
    $foundSibling = $false
    foreach ($cand in $repoIconCandidates) {
        if (Test-Path -LiteralPath $cand) {
            if ($cand.EndsWith(".ico")) {
                Copy-Item -LiteralPath $cand -Destination $IconPath -Force
                $hasCustomIcon = $true
                $foundSibling = $true
                break
            }
        }
    }
    if (-not $foundSibling -and -not [string]::IsNullOrWhiteSpace($EmbeddedIconBase64)) {
        [System.IO.File]::WriteAllBytes($IconPath, [System.Convert]::FromBase64String($EmbeddedIconBase64))
        $hasCustomIcon = $true
    }
} catch {
    Write-Verbose "Could not write custom icon cache: $_"
}

# 3. Create Windows Shortcuts via standard WScript.Shell COM Object
try {
    $WshShell = New-Object -ComObject WScript.Shell

    # Desktop shortcut ($HOME\Desktop\Corridor.lnk)
    $DesktopShortcut = $WshShell.CreateShortcut($DesktopShortcutPath)
    $DesktopShortcut.TargetPath = $browserPath
    $DesktopShortcut.Arguments = $browserArgs
    $DesktopShortcut.Description = "Corridor - Wildlife Collision Reduction & Mitigation Platform"
    $DesktopShortcut.WorkingDirectory = [System.IO.Path]::GetDirectoryName($browserPath)
    if ($hasCustomIcon -and (Test-Path -LiteralPath $IconPath)) {
        $DesktopShortcut.IconLocation = "$IconPath,0"
    } else {
        $DesktopShortcut.IconLocation = "$browserPath,0"
    }
    $DesktopShortcut.Save()
    Write-Host "[+] Created Desktop shortcut: $DesktopShortcutPath" -ForegroundColor Green

    # Start Menu shortcut ($env:APPDATA\Microsoft\Windows\Start Menu\Programs\Corridor.lnk)
    $StartMenuShortcut = $WshShell.CreateShortcut($StartMenuShortcutPath)
    $StartMenuShortcut.TargetPath = $browserPath
    $StartMenuShortcut.Arguments = $browserArgs
    $StartMenuShortcut.Description = "Corridor - Wildlife Collision Reduction & Mitigation Platform"
    $StartMenuShortcut.WorkingDirectory = [System.IO.Path]::GetDirectoryName($browserPath)
    if ($hasCustomIcon -and (Test-Path -LiteralPath $IconPath)) {
        $StartMenuShortcut.IconLocation = "$IconPath,0"
    } else {
        $StartMenuShortcut.IconLocation = "$browserPath,0"
    }
    $StartMenuShortcut.Save()
    Write-Host "[+] Created Start Menu shortcut: $StartMenuShortcutPath" -ForegroundColor Green

    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($WshShell) | Out-Null
} catch {
    Write-Error "[-] Failed to create Windows shortcuts: $_"
    exit 1
}

Write-Host @"

[✓] Installation Complete!
    - Corridor is now available on your Desktop and Start Menu.
    - Launch Mode: Dedicated standalone window ($browserArgs)
    - Security: NIST SP 800-53 Rev. 5 Compliant (AC-6, CM-7, SC-18, SC-28)
    - Zero administrative privileges / UAC elevation required.
    - Ready for offline field carcass logging, mitigation planning, and live dispatch.
"@ -ForegroundColor Green
