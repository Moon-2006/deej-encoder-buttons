package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/jfreymuth/pulse/proto"
	"github.com/mitchellh/go-ps"
)

const (
	getCurrentWindowInternalCooldown = time.Millisecond * 350
)

var (
	lastGetCurrentWindowResult []string
	lastGetCurrentWindowCall   = time.Now()
)

// displayServerType represents the type of display server being used
type displayServerType int

const (
	displayServerUnknown displayServerType = iota
	displayServerX11
	displayServerWayland
)

// detectDisplayServer determines which display server is being used
func detectDisplayServer() displayServerType {
	// Check for X11 first (including XWayland)
	// Many Wayland compositors run XWayland, so try X11 first
	if os.Getenv("DISPLAY") != "" {
		return displayServerX11
	}

	// If no X11, check for pure Wayland
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return displayServerWayland
	}

	return displayServerUnknown
}

// getAudioProcessNamesForPID connects to PulseAudio and finds all audio session process names
// that match the given PID. This is useful for finding the actual audio process when
// the window process name differs from the audio process name (e.g., Wine/Proton games)
func getAudioProcessNamesForPID(pid int) ([]string, error) {
	// Connect to PulseAudio
	client, conn, err := proto.Connect("")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PulseAudio: %w", err)
	}
	defer conn.Close()

	// Get all sink inputs (audio sessions)
	request := proto.GetSinkInputInfoList{}
	reply := proto.GetSinkInputInfoListReply{}

	if err := client.Request(&request, &reply); err != nil {
		return nil, fmt.Errorf("failed to get sink input list: %w", err)
	}

	// Find audio sessions matching the PID
	var audioProcessNames []string
	for _, info := range reply {
		// Check if this audio session belongs to our PID
		if pidProp, ok := info.Properties["application.process.id"]; ok {
			sessionPID, err := strconv.Atoi(pidProp.String())
			if err != nil {
				continue
			}

			if sessionPID == pid {
				// Found a matching audio session, get the process binary name
				if nameProp, ok := info.Properties["application.process.binary"]; ok {
					audioProcessNames = append(audioProcessNames, nameProp.String())
				}
			}
		}
	}

	return audioProcessNames, nil
}

// getAudioProcessNamesForPIDTree finds audio sessions for a PID and all its child processes
// This is useful for Wine/Proton games where the audio might come from a child process
func getAudioProcessNamesForPIDTree(rootPID int) ([]string, error) {
	// Connect to PulseAudio
	client, conn, err := proto.Connect("")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PulseAudio: %w", err)
	}
	defer conn.Close()

	// Get all sink inputs (audio sessions)
	request := proto.GetSinkInputInfoList{}
	reply := proto.GetSinkInputInfoListReply{}

	if err := client.Request(&request, &reply); err != nil {
		return nil, fmt.Errorf("failed to get sink input list: %w", err)
	}

	// Get all PIDs in the process tree
	pidsInTree := make(map[int]bool)
	pidsInTree[rootPID] = true

	// Get all processes
	processes, err := ps.Processes()
	if err == nil {
		// Build a tree of child processes
		for {
			foundNew := false
			for _, proc := range processes {
				ppid := proc.PPid()
				pid := proc.Pid()

				// If parent is in our tree and child isn't, add child
				if pidsInTree[ppid] && !pidsInTree[pid] {
					pidsInTree[pid] = true
					foundNew = true
				}
			}

			// Stop when we don't find any new processes
			if !foundNew {
				break
			}
		}
	}

	// Find audio sessions matching any PID in the tree
	var audioProcessNames []string
	for _, info := range reply {
		if pidProp, ok := info.Properties["application.process.id"]; ok {
			sessionPID, err := strconv.Atoi(pidProp.String())
			if err != nil {
				continue
			}

			if pidsInTree[sessionPID] {
				// Found a matching audio session in the process tree
				if nameProp, ok := info.Properties["application.process.binary"]; ok {
					audioProcessNames = append(audioProcessNames, nameProp.String())
				}
			}
		}
	}

	return audioProcessNames, nil
}

func getCurrentWindowProcessNames() ([]string, error) {
	// Apply an internal cooldown to avoid calling system APIs too frequently
	now := time.Now()
	if lastGetCurrentWindowCall.Add(getCurrentWindowInternalCooldown).After(now) {
		fmt.Fprintf(os.Stderr, "[deej.current] Using cached result\n")
		return lastGetCurrentWindowResult, nil
	}

	lastGetCurrentWindowCall = now

	// Get the window title and process info
	windowTitle, windowPID, err := getCurrentWindowInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[deej.current] Error getting window info: %v\n", err)
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "[deej.current] Window detected: title='%s', PID=%d\n", windowTitle, windowPID)

	// Get process name for fallback checks
	process, _ := ps.FindProcess(windowPID)
	var processName string
	if process != nil {
		processName = strings.ToLower(process.Executable())
	}

	windowTitleLower := strings.ToLower(windowTitle)

	// SUPER AGGRESSIVE fallback for Wine/Proton games
	// Strategy: If there's ANY wine audio playing AND the current window isn't a known native app,
	// assume it's a Wine/Proton game and return wine64-preloader

	// List of known native applications that should NOT trigger wine fallback
	knownNativeApps := []string{
		"spotify", "chrome", "firefox", "discord", "slack", "teams",
		"vlc", "mpv", "code", "vscode", "vscodium", "telegram",
		"thunderbird", "evolution", "dolphin", "nautilus", "konsole",
		"gnome-terminal", "terminal", "kitty", "alacritty",
	}

	isKnownNativeApp := false
	for _, app := range knownNativeApps {
		if strings.Contains(processName, app) || strings.Contains(windowTitleLower, app) {
			isKnownNativeApp = true
			break
		}
	}

	// Check if there's ANY wine audio playing in the system
	hasWineAudioPlaying := false
	client, conn, err := proto.Connect("")
	if err == nil {
		defer conn.Close()
		request := proto.GetSinkInputInfoList{}
		reply := proto.GetSinkInputInfoListReply{}
		if client.Request(&request, &reply) == nil {
			for _, info := range reply {
				if binaryProp, ok := info.Properties["application.process.binary"]; ok {
					binary := strings.ToLower(binaryProp.String())
					if strings.Contains(binary, "wine") {
						hasWineAudioPlaying = true
						fmt.Fprintf(os.Stderr, "[deej.current] Found Wine audio session: %s\n", binary)
						break
					}
				}
			}
		}
	}

	// AGGRESSIVE FALLBACK: If wine is playing audio and this isn't a known native app, assume it's a game
	if hasWineAudioPlaying && !isKnownNativeApp {
		fmt.Fprintf(os.Stderr, "[deej.current] AGGRESSIVE FALLBACK: Wine audio detected + unknown app (title:'%s') → using wine-preloader\n", windowTitle)
		result := []string{"wine64-preloader", "wine-preloader"}
		lastGetCurrentWindowResult = result
		return result, nil
	}

	// Also check for obvious Wine/Proton indicators in process or window
	isObviousWineProcess := strings.Contains(processName, "wine") ||
		strings.Contains(processName, "proton") ||
		strings.Contains(windowTitleLower, ".exe")

	if isObviousWineProcess {
		fmt.Fprintf(os.Stderr, "[deej.current] Detected obvious Wine/Proton process, using wine fallback\n")
		result := []string{"wine64-preloader", "wine-preloader"}
		lastGetCurrentWindowResult = result
		return result, nil
	}

	// Try to find audio sessions for this PID first
	audioProcessNames, err := getAudioProcessNamesForPID(windowPID)
	if err == nil && len(audioProcessNames) > 0 {
		fmt.Fprintf(os.Stderr, "[deej.current] Found %d audio processes by PID: %v\n", len(audioProcessNames), audioProcessNames)
		lastGetCurrentWindowResult = audioProcessNames
		return audioProcessNames, nil
	}
	fmt.Fprintf(os.Stderr, "[deej.current] No audio processes found by PID\n")

	// Try to find audio sessions by matching window title with audio application names
	// This is more reliable for containerized apps (Flatpak/Steam) where PIDs don't match
	audioByName, err := getAudioProcessNamesByWindowTitle(windowTitle)
	if err == nil && len(audioByName) > 0 {
		fmt.Fprintf(os.Stderr, "[deej.current] Found %d audio processes by window title: %v\n", len(audioByName), audioByName)
		lastGetCurrentWindowResult = audioByName
		return audioByName, nil
	}
	fmt.Fprintf(os.Stderr, "[deej.current] No audio processes found by window title\n")

	// Try to find audio sessions for child processes
	childAudioProcessNames, err := getAudioProcessNamesForPIDTree(windowPID)
	if err == nil && len(childAudioProcessNames) > 0 {
		fmt.Fprintf(os.Stderr, "[deej.current] Found %d audio processes in process tree: %v\n", len(childAudioProcessNames), childAudioProcessNames)
		lastGetCurrentWindowResult = childAudioProcessNames
		return childAudioProcessNames, nil
	}
	fmt.Fprintf(os.Stderr, "[deej.current] No audio processes found in process tree\n")

	// Final fallback: return the process name
	if process == nil {
		return nil, fmt.Errorf("failed to find process for PID %d", windowPID)
	}

	result := []string{process.Executable()}
	fmt.Fprintf(os.Stderr, "[deej.current] Using fallback process name: %v\n", result)
	lastGetCurrentWindowResult = result
	return result, nil
}

// getCurrentWindowInfo gets both the window title and PID of the currently focused window
func getCurrentWindowInfo() (string, int, error) {
	displayServer := detectDisplayServer()

	switch displayServer {
	case displayServerX11:
		return getCurrentWindowInfoX11()
	case displayServerWayland:
		return getCurrentWindowInfoWayland()
	default:
		return "", 0, fmt.Errorf("unable to detect display server")
	}
}

// getCurrentWindowInfoX11 gets window title and PID using X11
func getCurrentWindowInfoX11() (string, int, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return "", 0, fmt.Errorf("failed to connect to X server: %w", err)
	}
	defer conn.Close()

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	// Get active window
	activeWindowAtom, err := getAtom(conn, "_NET_ACTIVE_WINDOW")
	if err != nil {
		return "", 0, fmt.Errorf("failed to get _NET_ACTIVE_WINDOW atom: %w", err)
	}

	reply, err := xproto.GetProperty(conn, false, root, activeWindowAtom, xproto.AtomWindow, 0, 1).Reply()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get active window property: %w", err)
	}

	if reply.ValueLen == 0 {
		return "", 0, fmt.Errorf("no active window")
	}

	windowID := xproto.Window(get32(reply.Value))

	// Get window title
	nameAtom, err := getAtom(conn, "_NET_WM_NAME")
	if err != nil {
		nameAtom, err = getAtom(conn, "WM_NAME")
		if err != nil {
			return "", 0, fmt.Errorf("failed to get window name atom: %w", err)
		}
	}

	nameReply, err := xproto.GetProperty(conn, false, windowID, nameAtom, xproto.GetPropertyTypeAny, 0, 1024).Reply()
	var windowTitle string
	if err == nil && nameReply.ValueLen > 0 {
		windowTitle = string(nameReply.Value)
	}

	// Get PID
	pidAtom, err := getAtom(conn, "_NET_WM_PID")
	if err != nil {
		return "", 0, fmt.Errorf("failed to get _NET_WM_PID atom: %w", err)
	}

	pidReply, err := xproto.GetProperty(conn, false, windowID, pidAtom, xproto.AtomCardinal, 0, 1).Reply()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get window PID property: %w", err)
	}

	if pidReply.ValueLen == 0 {
		return "", 0, fmt.Errorf("window has no PID property")
	}

	pid := int(get32(pidReply.Value))
	return windowTitle, pid, nil
}

// getCurrentWindowInfoWayland gets window title and PID using Wayland
func getCurrentWindowInfoWayland() (string, int, error) {
	// For now, we only support getting PID from Wayland
	// Window titles are harder to get in Wayland
	pid, err := getCurrentWindowPIDWayland()
	if err != nil {
		return "", 0, err
	}

	// Try to get process name as fallback for title
	process, err := ps.FindProcess(pid)
	if err == nil && process != nil {
		return process.Executable(), pid, nil
	}

	return "", pid, nil
}

// getAudioProcessNamesByWindowTitle finds audio sessions whose application name matches the window title
// This is useful for Flatpak/containerized apps where PIDs don't match between host and container
func getAudioProcessNamesByWindowTitle(windowTitle string) ([]string, error) {
	if windowTitle == "" {
		return nil, fmt.Errorf("empty window title")
	}

	client, conn, err := proto.Connect("")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PulseAudio: %w", err)
	}
	defer conn.Close()

	request := proto.GetSinkInputInfoList{}
	reply := proto.GetSinkInputInfoListReply{}

	if err := client.Request(&request, &reply); err != nil {
		return nil, fmt.Errorf("failed to get sink input list: %w", err)
	}

	var audioProcessNames []string
	windowTitleLower := strings.ToLower(windowTitle)

	for _, info := range reply {
		// Check if audio application name contains part of the window title
		if appName, ok := info.Properties["application.name"]; ok {
			appNameStr := strings.ToLower(appName.String())

			// Check for partial matches (e.g., "ARC Raiders" in window title matches "ARC Raiders" in audio)
			if strings.Contains(appNameStr, windowTitleLower) || strings.Contains(windowTitleLower, appNameStr) {
				if binaryProp, ok := info.Properties["application.process.binary"]; ok {
					audioProcessNames = append(audioProcessNames, binaryProp.String())
				}
			}
		}
	}

	return audioProcessNames, nil
}

// getCurrentWindowPIDX11 gets the PID of the focused window using X11
func getCurrentWindowPIDX11() (int, error) {
	// Connect to X server
	conn, err := xgb.NewConn()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to X server: %w", err)
	}
	defer conn.Close()

	// Get the setup information
	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	// Get the _NET_ACTIVE_WINDOW atom
	activeWindowAtom, err := getAtom(conn, "_NET_ACTIVE_WINDOW")
	if err != nil {
		return 0, fmt.Errorf("failed to get _NET_ACTIVE_WINDOW atom: %w", err)
	}

	// Get the active window ID
	reply, err := xproto.GetProperty(conn, false, root, activeWindowAtom, xproto.AtomWindow, 0, 1).Reply()
	if err != nil {
		return 0, fmt.Errorf("failed to get active window property: %w", err)
	}

	if reply.ValueLen == 0 {
		return 0, fmt.Errorf("no active window")
	}

	// Parse the window ID
	windowID := xproto.Window(get32(reply.Value))

	// Get the _NET_WM_PID atom
	pidAtom, err := getAtom(conn, "_NET_WM_PID")
	if err != nil {
		return 0, fmt.Errorf("failed to get _NET_WM_PID atom: %w", err)
	}

	// Get the PID of the window
	pidReply, err := xproto.GetProperty(conn, false, windowID, pidAtom, xproto.AtomCardinal, 0, 1).Reply()
	if err != nil {
		return 0, fmt.Errorf("failed to get window PID property: %w", err)
	}

	if pidReply.ValueLen == 0 {
		return 0, fmt.Errorf("window has no PID property")
	}

	pid := int(get32(pidReply.Value))
	return pid, nil
}

// getAtom is a helper function to get an X11 atom by name
func getAtom(conn *xgb.Conn, name string) (xproto.Atom, error) {
	reply, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		return 0, err
	}
	return reply.Atom, nil
}

// getCurrentWindowPIDWayland gets the PID of the focused window using Wayland
func getCurrentWindowPIDWayland() (int, error) {
	// Wayland doesn't expose focused window info directly for security reasons
	// We need to use compositor-specific methods

	// Try different compositor-specific commands
	compositorCommands := []struct {
		name    string
		command string
		args    []string
		parser  func(string) (int, error)
	}{
		{
			name:    "Sway",
			command: "swaymsg",
			args:    []string{"-t", "get_tree"},
			parser:  parseSwayTree,
		},
		{
			name:    "Hyprland",
			command: "hyprctl",
			args:    []string{"activewindow", "-j"},
			parser:  parseHyprlandWindow,
		},
		{
			name:    "KDE (kwin_wayland)",
			command: "qdbus",
			args:    []string{"org.kde.KWin", "/KWin", "org.kde.KWin.activeWindow"},
			parser:  parseKDEWindow,
		},
	}

	// Try each compositor command
	for _, cmd := range compositorCommands {
		if _, err := exec.LookPath(cmd.command); err != nil {
			// Command not found, try next one
			continue
		}

		output, err := exec.Command(cmd.command, cmd.args...).Output()
		if err != nil {
			// Command failed, try next one
			continue
		}

		pid, err := cmd.parser(string(output))
		if err != nil {
			// Parser failed, try next one
			continue
		}

		return pid, nil
	}

	// If no compositor-specific method worked, return error
	return 0, fmt.Errorf("wayland compositor not supported - please use X11 or install swaymsg/hyprctl for your compositor")
}

// parseSwayTree parses swaymsg output to find the focused window's PID
func parseSwayTree(output string) (int, error) {
	// This is a simplified parser - a full implementation would use JSON parsing
	// Look for "focused": true and then find the "pid" field
	lines := strings.Split(output, "\n")
	foundFocused := false

	for _, line := range lines {
		if strings.Contains(line, `"focused": true`) {
			foundFocused = true
		}

		if foundFocused && strings.Contains(line, `"pid"`) {
			// Extract PID from line like: "pid": 12345,
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}

			pidStr := strings.TrimSpace(parts[1])
			pidStr = strings.TrimSuffix(pidStr, ",")

			pid, err := strconv.Atoi(pidStr)
			if err == nil {
				return pid, nil
			}
		}

		// Reset if we find another node before finding the PID
		if foundFocused && strings.Contains(line, `"focused": false`) {
			foundFocused = false
		}
	}

	return 0, fmt.Errorf("could not find focused window PID in sway output")
}

// parseHyprlandWindow parses hyprctl output to find the active window's PID
func parseHyprlandWindow(output string) (int, error) {
	// Look for "pid": 12345 in JSON output
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, `"pid"`) {
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}

			pidStr := strings.TrimSpace(parts[1])
			pidStr = strings.TrimSuffix(pidStr, ",")

			pid, err := strconv.Atoi(pidStr)
			if err == nil {
				return pid, nil
			}
		}
	}

	return 0, fmt.Errorf("could not find window PID in hyprland output")
}

// parseKDEWindow parses KDE window info to get PID
func parseKDEWindow(output string) (int, error) {
	// KDE returns window ID, we need to get PID from it
	// This is a simplified implementation
	return 0, fmt.Errorf("KDE window parsing not fully implemented")
}

// readProcComm reads the process name from /proc/<pid>/comm
func readProcComm(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// readProcCmdline reads the full command line from /proc/<pid>/cmdline
func readProcCmdline(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", err
	}

	// cmdline uses null bytes as separators
	parts := bytes.Split(data, []byte{0})
	if len(parts) > 0 {
		// Get just the executable name (last part of the path)
		cmdPath := string(parts[0])
		cmdParts := strings.Split(cmdPath, "/")
		return cmdParts[len(cmdParts)-1], nil
	}

	return "", fmt.Errorf("empty cmdline")
}

// get32 reads a uint32 from a byte slice in little-endian format
func get32(b []byte) uint32 {
	if len(b) < 4 {
		return 0
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
