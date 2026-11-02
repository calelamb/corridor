// tests/desktop_install_test.go
package tests

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// resolvePath searches for target file path across standard test execution directories
func resolvePath(t *testing.T, relPaths ...string) string {
	t.Helper()
	for _, p := range relPaths {
		if _, err := os.Stat(p); err == nil {
			abs, err := filepath.Abs(p)
			if err == nil {
				return abs
			}
			return p
		}
	}
	t.Fatalf("could not locate file among candidates: %v", relPaths)
	return ""
}

// TestDesktopInstallScriptSyntaxAndPermissions verifies the shell script syntax and executable bit
func TestDesktopInstallScriptSyntaxAndPermissions(t *testing.T) {
	scriptPath := resolvePath(t, "../scripts/install-desktop.sh", "scripts/install-desktop.sh")

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("failed to stat script: %v", err)
	}

	// Verify executable permission
	if info.Mode()&0111 == 0 {
		t.Errorf("expected script %s to be executable, got mode %v", scriptPath, info.Mode())
	}

	// Run syntax check via bash -n (or sh -n if bash is not available)
	shellCmd := "bash"
	if _, err := exec.LookPath("bash"); err != nil {
		shellCmd = "sh"
	}

	cmd := exec.Command(shellCmd, "-n", scriptPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("shell script syntax check failed (%s -n %s): %v\nOutput: %s", shellCmd, scriptPath, err, string(out))
	}

	// Verify help flag output
	cmdHelp := exec.Command(scriptPath, "--help")
	helpOut, err := cmdHelp.CombinedOutput()
	if err != nil {
		t.Fatalf("script --help invocation failed: %v\nOutput: %s", err, string(helpOut))
	}

	helpText := string(helpOut)
	requiredHelpKeywords := []string{
		"Corridor Desktop Installer",
		"--uninstall",
		"--url",
		"--name",
		"--desktop",
		"--os",
		"Zero-admin",
		"NIST SP 800-53",
	}

	for _, kw := range requiredHelpKeywords {
		if !strings.Contains(helpText, kw) {
			t.Errorf("missing keyword %q in --help output", kw)
		}
	}
}

// TestDesktopInstallMacOSFlow verifies macOS .app bundle creation and uninstallation in isolated environment
func TestDesktopInstallMacOSFlow(t *testing.T) {
	scriptPath := resolvePath(t, "../scripts/install-desktop.sh", "scripts/install-desktop.sh")
	tmpHome := t.TempDir()

	// 1. Run macOS installation with temporary HOME
	cmd := exec.Command(scriptPath, "--os", "macos", "--desktop", "--url", "https://corridorwildlife.org", "--name", "CorridorTest")
	cmd.Env = append(os.Environ(), "HOME="+tmpHome)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("macOS install failed: %v\nOutput: %s", err, string(out))
	}

	appDir := filepath.Join(tmpHome, "Applications", "CorridorTest.app")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		t.Fatalf("expected macOS app bundle at %s, but directory does not exist", appDir)
	}

	// Verify Info.plist
	plistPath := filepath.Join(appDir, "Contents", "Info.plist")
	plistData, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("failed to read Info.plist: %v", err)
	}
	plistContent := string(plistData)
	if !strings.Contains(plistContent, "<key>CFBundleExecutable</key>") ||
		!strings.Contains(plistContent, "<string>corridor</string>") {
		t.Errorf("Info.plist missing expected CFBundleExecutable corridor")
	}
	if !strings.Contains(plistContent, "org.corridorwildlife.desktop") {
		t.Errorf("Info.plist missing CFBundleIdentifier")
	}

	// Verify launcher script
	launcherPath := filepath.Join(appDir, "Contents", "MacOS", "corridor")
	launcherInfo, err := os.Stat(launcherPath)
	if err != nil {
		t.Fatalf("failed to stat launcher script: %v", err)
	}
	if launcherInfo.Mode()&0111 == 0 {
		t.Errorf("launcher script at %s is not executable", launcherPath)
	}

	launcherData, err := os.ReadFile(launcherPath)
	if err != nil {
		t.Fatalf("failed to read launcher script: %v", err)
	}
	launcherContent := string(launcherData)
	if !strings.Contains(launcherContent, "--app=") {
		t.Errorf("launcher script missing --app= standalone flag")
	}
	if !strings.Contains(launcherContent, "Google Chrome") ||
		!strings.Contains(launcherContent, "Microsoft Edge") ||
		!strings.Contains(launcherContent, "Safari") {
		t.Errorf("launcher script missing expected browser fallbacks")
	}

	// Verify icon
	iconPath := filepath.Join(appDir, "Contents", "Resources", "icon-512.svg")
	iconData, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatalf("failed to read icon file: %v", err)
	}
	if !strings.Contains(string(iconData), "<svg") {
		t.Errorf("expected valid SVG icon in %s", iconPath)
	}

	// Verify Desktop link
	desktopLink := filepath.Join(tmpHome, "Desktop", "CorridorTest.app")
	if _, err := os.Stat(desktopLink); os.IsNotExist(err) {
		t.Errorf("expected desktop link at %s", desktopLink)
	}

	// 2. Test clean uninstallation
	cmdUninstall := exec.Command(scriptPath, "--os", "macos", "--name", "CorridorTest", "--uninstall")
	cmdUninstall.Env = append(os.Environ(), "HOME="+tmpHome)
	uninstOut, err := cmdUninstall.CombinedOutput()
	if err != nil {
		t.Fatalf("macOS uninstall failed: %v\nOutput: %s", err, string(uninstOut))
	}

	if _, err := os.Stat(appDir); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed after uninstall", appDir)
	}
	if _, err := os.Stat(desktopLink); !os.IsNotExist(err) {
		t.Errorf("expected desktop link %s to be removed after uninstall", desktopLink)
	}
}

// TestDesktopInstallLinuxFlow verifies Linux XDG freedesktop specification shortcuts and uninstallation
func TestDesktopInstallLinuxFlow(t *testing.T) {
	scriptPath := resolvePath(t, "../scripts/install-desktop.sh", "scripts/install-desktop.sh")
	tmpHome := t.TempDir()

	// 1. Run Linux installation with temporary HOME
	cmd := exec.Command(scriptPath, "--os", "linux", "--desktop", "--url", "https://corridorwildlife.org", "--name", "Corridor")
	cmd.Env = append(os.Environ(), "HOME="+tmpHome)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Linux install failed: %v\nOutput: %s", err, string(out))
	}

	// Verify XDG desktop file
	desktopPath := filepath.Join(tmpHome, ".local", "share", "applications", "corridor.desktop")
	desktopData, err := os.ReadFile(desktopPath)
	if err != nil {
		t.Fatalf("failed to read .desktop file at %s: %v", desktopPath, err)
	}

	desktopContent := string(desktopData)
	requiredDesktopStrings := []string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=Corridor",
		"Categories=Utility;Science;Geography;",
		"Icon=corridor",
		"StartupWMClass=corridorwildlife.org",
		"Actions=Report;Mitigate;Hotspots;Rangers;",
		"[Desktop Action Report]",
		"[Desktop Action Mitigate]",
		"[Desktop Action Hotspots]",
		"[Desktop Action Rangers]",
	}

	for _, req := range requiredDesktopStrings {
		if !strings.Contains(desktopContent, req) {
			t.Errorf("corridor.desktop missing required field %q", req)
		}
	}

	// Verify Linux launcher binary
	launcherPath := filepath.Join(tmpHome, ".local", "bin", "corridor")
	launcherInfo, err := os.Stat(launcherPath)
	if err != nil {
		t.Fatalf("failed to stat launcher script at %s: %v", launcherPath, err)
	}
	if launcherInfo.Mode()&0111 == 0 {
		t.Errorf("launcher script %s is not executable", launcherPath)
	}

	launcherData, err := os.ReadFile(launcherPath)
	if err != nil {
		t.Fatalf("failed to read launcher script: %v", err)
	}
	launcherContent := string(launcherData)
	if !strings.Contains(launcherContent, "--app=") {
		t.Errorf("Linux launcher missing --app= flag")
	}
	if !strings.Contains(launcherContent, "google-chrome") ||
		!strings.Contains(launcherContent, "microsoft-edge") ||
		!strings.Contains(launcherContent, "xdg-open") {
		t.Errorf("Linux launcher missing expected browser search candidates")
	}

	// Verify scalable vector icon installation
	iconPath := filepath.Join(tmpHome, ".local", "share", "icons", "hicolor", "scalable", "apps", "corridor.svg")
	iconData, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatalf("failed to read installed scalable icon: %v", err)
	}
	if !strings.Contains(string(iconData), "<svg") {
		t.Errorf("expected valid SVG in %s", iconPath)
	}

	// Verify Desktop launcher file
	desktopShortcut := filepath.Join(tmpHome, "Desktop", "corridor.desktop")
	if _, err := os.Stat(desktopShortcut); os.IsNotExist(err) {
		t.Errorf("expected desktop shortcut at %s", desktopShortcut)
	}

	// 2. Test clean Linux uninstallation
	cmdUninstall := exec.Command(scriptPath, "--os", "linux", "--uninstall")
	cmdUninstall.Env = append(os.Environ(), "HOME="+tmpHome)
	uninstOut, err := cmdUninstall.CombinedOutput()
	if err != nil {
		t.Fatalf("Linux uninstall failed: %v\nOutput: %s", err, string(uninstOut))
	}

	if _, err := os.Stat(desktopPath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed after uninstall", desktopPath)
	}
	if _, err := os.Stat(launcherPath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed after uninstall", launcherPath)
	}
	if _, err := os.Stat(iconPath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed after uninstall", iconPath)
	}
	if _, err := os.Stat(desktopShortcut); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed after uninstall", desktopShortcut)
	}
}

// TestDesktopInstallPowerShellStructureAndSecurity parses and validates Install-Corridor.ps1
func TestDesktopInstallPowerShellStructureAndSecurity(t *testing.T) {
	psPath := resolvePath(t, "../scripts/Install-Corridor.ps1", "scripts/Install-Corridor.ps1")

	data, err := os.ReadFile(psPath)
	if err != nil {
		t.Fatalf("failed to read Install-Corridor.ps1 from %s: %v", psPath, err)
	}

	content := string(data)
	if len(content) < 3000 {
		t.Errorf("PowerShell script too short: got %d bytes, want >= 3000 bytes", len(content))
	}

	// Verify parameters
	requiredParams := []string{
		"[switch]$Uninstall",
		"[string]$Url",
		"[string]$AppName",
	}
	for _, param := range requiredParams {
		if !strings.Contains(content, param) {
			t.Errorf("missing parameter %q in Install-Corridor.ps1", param)
		}
	}

	// Verify pure zero-admin user-space paths
	requiredUserPaths := []string{
		"SpecialFolder]::Desktop",
		"SpecialFolder]::Programs",
		"LOCALAPPDATA",
	}
	for _, p := range requiredUserPaths {
		if !strings.Contains(content, p) {
			t.Errorf("missing user-space directory reference %q in Install-Corridor.ps1", p)
		}
	}

	// Verify no elevation or admin checks are mandated
	if strings.Contains(strings.ToLower(content), "runas") ||
		strings.Contains(strings.ToLower(content), "verb runas") {
		t.Errorf("found administrative elevation 'runas' in zero-admin PowerShell script")
	}

	// Verify browser targets and app mode
	requiredBrowserTerms := []string{
		"msedge.exe",
		"chrome.exe",
		"--app=$Url",
	}
	for _, term := range requiredBrowserTerms {
		if !strings.Contains(content, term) {
			t.Errorf("missing browser launcher term %q in Install-Corridor.ps1", term)
		}
	}

	// Verify COM object usage for .lnk shortcut creation
	requiredCOMTerms := []string{
		"WScript.Shell",
		"CreateShortcut",
		"IconLocation",
		"Save()",
	}
	for _, com := range requiredCOMTerms {
		if !strings.Contains(content, com) {
			t.Errorf("missing COM shortcut term %q in Install-Corridor.ps1", com)
		}
	}

	// Verify uninstallation logic
	if !strings.Contains(content, "Remove-Item") {
		t.Errorf("missing Remove-Item cleanup logic in Install-Corridor.ps1")
	}

	// Verify NIST SP 800-53 security controls documentation
	requiredSecurityControls := []string{
		"NIST SP 800-53",
		"AC-6",
		"CM-7",
		"SC-18",
		"SC-28",
	}
	for _, ctrl := range requiredSecurityControls {
		if !strings.Contains(content, ctrl) {
			t.Errorf("missing security compliance reference %q in Install-Corridor.ps1", ctrl)
		}
	}

	// Verify embedded ICO base64 is present and represents valid icon data
	re := regexp.MustCompile(`\$EmbeddedIconBase64\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		t.Fatalf("could not locate $EmbeddedIconBase64 declaration in Install-Corridor.ps1")
	}

	icoBytes, err := base64.StdEncoding.DecodeString(matches[1])
	if err != nil {
		t.Fatalf("failed to decode embedded base64 icon: %v", err)
	}

	if len(icoBytes) < 22 {
		t.Fatalf("decoded icon byte slice too small: got %d bytes, want >= 22", len(icoBytes))
	}

	// ICO header signature check: 00 00 01 00 (Reserved=0, Type=1 for ICO)
	if icoBytes[0] != 0 || icoBytes[1] != 0 || icoBytes[2] != 1 || icoBytes[3] != 0 {
		t.Errorf("invalid ICO header bytes: %v", icoBytes[:4])
	}
}

// TestDesktopDocumentation verifies completeness of docs/DESKTOP_INSTALL.md
func TestDesktopDocumentation(t *testing.T) {
	docPath := resolvePath(t, "../docs/DESKTOP_INSTALL.md", "docs/DESKTOP_INSTALL.md")

	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("failed to read DESKTOP_INSTALL.md from %s: %v", docPath, err)
	}

	content := string(data)
	if len(content) < 3000 {
		t.Errorf("DESKTOP_INSTALL.md too short: got %d bytes, want >= 3000 bytes", len(content))
	}

	requiredSections := []string{
		"1. Executive Overview",
		"2. 1-Minute Automated Desktop Installation",
		"2.1. Windows 10 / 11 GFE",
		"2.2. macOS Workstations",
		"2.3. Linux Workstations",
		"3. 1-Minute Manual Browser Installation",
		"4. Offline Usage & Field Persistence Architecture",
		"5. Federal Cybersecurity & FISMA Compliance Notes",
		"6. Dispatch Desk & Multi-Monitor Deployment",
		"7. Keyboard Shortcuts Reference Card",
	}

	for _, sec := range requiredSections {
		if !strings.Contains(content, sec) {
			t.Errorf("missing section header %q in DESKTOP_INSTALL.md", sec)
		}
	}

	requiredKeywords := []string{
		"Zero-Admin",
		"Microsoft Edge",
		"Google Chrome",
		"Safari",
		"Install-Corridor.ps1",
		"install-desktop.sh",
		"IndexedDB",
		"ServiceWorker",
		"NIST SP 800-53",
		"AC-6",
		"CM-7",
		"SC-18",
		"SC-28",
		"Section 508",
		"Poaching Protection Privacy",
	}

	for _, kw := range requiredKeywords {
		if !strings.Contains(content, kw) {
			t.Errorf("missing keyword %q in DESKTOP_INSTALL.md", kw)
		}
	}
}

// TestDesktopWebManifestIntegrity ensures web manifest declares standalone display and icons
func TestDesktopWebManifestIntegrity(t *testing.T) {
	manifestPath := resolvePath(t, "../web/static/manifest.webmanifest", "web/static/manifest.webmanifest")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read web manifest from %s: %v", manifestPath, err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}

	display, ok := manifest["display"].(string)
	if !ok || display != "standalone" {
		t.Errorf("expected display mode 'standalone', got %q", display)
	}

	name, ok := manifest["name"].(string)
	if !ok || !strings.Contains(name, "Corridor") {
		t.Errorf("expected manifest name to contain 'Corridor', got %q", name)
	}

	icons, ok := manifest["icons"].([]interface{})
	if !ok || len(icons) == 0 {
		t.Fatalf("manifest missing icons array")
	}

	found512 := false
	for _, icon := range icons {
		m, ok := icon.(map[string]interface{})
		if !ok {
			continue
		}
		src, _ := m["src"].(string)
		if strings.Contains(src, "512") {
			found512 = true
			break
		}
	}
	if !found512 {
		t.Errorf("expected manifest to include 512px icon reference")
	}
}
