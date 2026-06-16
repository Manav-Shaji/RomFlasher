package internal

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	reProduct = regexp.MustCompile(`product:\s*(\S+)`)
	reBattery = regexp.MustCompile(`battery-voltage:\s*(\S+)`)
	reSlot    = regexp.MustCompile(`current-slot:\s*(\S+)`)
	reAdbBatt = regexp.MustCompile(`level:\s*(\d+)`)
)

type DeviceUpdateMsg DeviceState
type PollMsg time.Time
type HeartbeatMsg time.Time

func HeartbeatCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return HeartbeatMsg(t)
	})
}

func PollDeviceCmd() tea.Cmd {
	return tea.Batch(
		tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
			return PollMsg(t)
		}),
		HeartbeatCmd(),
	)
}

// Device Database logic moved to device_db.go

func CheckDeviceState() tea.Msg {
	state := DeviceState{Mode: ModeDisconnected, Serial: "-", Model: "-", Battery: "-", Slot: "-", Secure: "-"}

	if out, err := exec.Command("fastboot", "devices").Output(); err == nil && len(out) > 0 {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			state.Mode, state.Serial = ModeFastboot, parts[0]
			state = fetchFastbootDetails(state)
			return DeviceUpdateMsg(state)
		}
	}

	if out, err := exec.Command("adb", "devices").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, l := range lines {
			p := strings.Fields(l)
			if len(p) >= 2 {
				serial, status := p[0], p[1]
				switch status {
				case "device":
					state.Mode, state.Serial = ModeDevice, serial
					state = fetchAdbDetails(state)
					return DeviceUpdateMsg(state)
				case "sideload":
					state.Mode, state.Serial = ModeSideload, serial
					state.Model = "SIDELOAD DEVICE"
					return DeviceUpdateMsg(state)
				case "recovery":
					state.Mode, state.Serial = ModeRecovery, serial
					state = fetchAdbDetails(state)
					return DeviceUpdateMsg(state)
				case "unauthorized":
					state.Mode, state.Serial = ModeUnauthorized, serial
					state.Model = "ACTION REQUIRED"
					return DeviceUpdateMsg(state)
				case "offline":
					state.Mode, state.Serial = ModeOffline, serial
					state.Model = "OFFLINE"
					return DeviceUpdateMsg(state)
				}
			}
		}
	}

	return DeviceUpdateMsg(state)
}

func fetchFastbootDetails(s DeviceState) DeviceState {
	// Model
	if out, err := exec.Command("fastboot", "-s", s.Serial, "getvar", "product").CombinedOutput(); err == nil {
		if m := reProduct.FindStringSubmatch(string(out)); len(m) > 1 {
			codename := strings.ToUpper(strings.TrimSpace(m[1]))
			pretty := prettyDeviceName(codename)
			if pretty != codename {
				s.Model = fmt.Sprintf("%s (%s)", pretty, codename)
			} else {
				s.Model = codename
			}
		}
	}

	// Battery
	if out, err := exec.Command("fastboot", "-s", s.Serial, "getvar", "battery-voltage").CombinedOutput(); err == nil {
		if m := reBattery.FindStringSubmatch(string(out)); len(m) > 1 {
			val := strings.TrimSpace(m[1])
			if val != "" && val != "-" {
				s.Battery = val + " mV"
			} else {
				s.Battery = val
			}
		}
	}

	// Secure / Lock Status
	// We check multiple variables as vendors use different names
	isUnlocked := false
	
	// 1. Check "unlocked" variable
	if out, err := exec.Command("fastboot", "-s", s.Serial, "getvar", "unlocked").CombinedOutput(); err == nil {
		if strings.Contains(strings.ToLower(string(out)), "yes") {
			isUnlocked = true
		}
	}

	// 2. Check "device-state" variable
	if !isUnlocked {
		if out, err := exec.Command("fastboot", "-s", s.Serial, "getvar", "device-state").CombinedOutput(); err == nil {
			if strings.Contains(strings.ToLower(string(out)), "unlocked") {
				isUnlocked = true
			}
		}
	}

	// 3. Fallback to "secure" variable (meaning Secure Boot / Locked)
	if isUnlocked {
		s.Secure = "NO"
	} else {
		if out, err := exec.Command("fastboot", "-s", s.Serial, "getvar", "secure").CombinedOutput(); err == nil {
			if strings.Contains(strings.ToLower(string(out)), "yes") {
				s.Secure = "YES"
			} else {
				s.Secure = "NO"
			}
		} else {
			s.Secure = "YES" // Default to safe assumption
		}
	}

	return s
}

func fetchAdbDetails(s DeviceState) DeviceState {
	// ADB Details (Shell Required)
	
	// 1. Get Properties
	getProp := func(prop string) string {
		out, _ := exec.Command("adb", "-s", s.Serial, "shell", "getprop", prop).Output()
		return strings.TrimSpace(string(out))
	}

	marketProp := getProp("ro.product.marketname")
	modelProp := getProp("ro.product.model")
	brandProp := getProp("ro.product.brand")
	codenameProp := strings.ToUpper(getProp("ro.product.device"))

	// 2. Try 'adb devices -l' for friendly model tag
	marketingLabel := ""
	if out, err := exec.Command("adb", "devices", "-l").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, l := range lines {
			fields := strings.Fields(l)
			if len(fields) > 0 && fields[0] == s.Serial {
				for _, f := range fields {
					if strings.HasPrefix(f, "model:") {
						marketingLabel = strings.TrimPrefix(f, "model:")
						marketingLabel = strings.ReplaceAll(marketingLabel, "_", " ")
						break
					}
				}
				break
			}
		}
	}

	// 3. Selection Strategy (ADB Mode = CLEAN)
	// We check internal database first for the most accurate mapping
	name := ""
	
	// Try internal mapping with everything we have
	candidates := []string{marketProp, modelProp, marketingLabel, codenameProp}
	for _, c := range candidates {
		if c == "" { continue }
		pretty := prettyDeviceName(c)
		if pretty != strings.ToUpper(c) {
			name = pretty
			break
		}
	}

	// Fallback to property chain if no mapping found
	if name == "" {
		if marketProp != "" {
			name = marketProp
		} else if marketingLabel != "" && len(marketingLabel) > 3 {
			name = marketingLabel
		} else if brandProp != "" && modelProp != "" {
			name = brandProp + " " + modelProp
		} else {
			name = modelProp
		}
	}

	s.Model = strings.ToUpper(name)

	// Battery
	if out, err := exec.Command("adb", "-s", s.Serial, "shell", "dumpsys", "battery").Output(); err == nil {
		if m := reAdbBatt.FindStringSubmatch(string(out)); len(m) > 1 {
			s.Battery = m[1] + "%"
		}
	}

	// Secure
	s.Secure = "YES"
	return s
}
