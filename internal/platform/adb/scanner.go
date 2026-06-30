package adb

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"sync"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"flashtool/internal/domain"
	"flashtool/internal/platform/sysutil"
)

var (
	reProduct = regexp.MustCompile(`product:\s*(\S+)`)
	reBattery = regexp.MustCompile(`battery-voltage:\s*(\S+)`)
	reSlot    = regexp.MustCompile(`current-slot:\s*(\S+)`)
	reSecure  = regexp.MustCompile(`secure:\s*(\S+)`)
	reAdbBatt = regexp.MustCompile(`level:\s*(\d+)`)

	isScanning  atomic.Bool
	deviceCache = make(map[string]domain.DeviceState)
	cacheMu     sync.RWMutex
)

type DeviceUpdateMsg domain.DeviceState
type PollMsg time.Time
type HeartbeatMsg time.Time
type SkipUpdateMsg struct{}

func HeartbeatCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return HeartbeatMsg(t)
	})
}

func PollDeviceCmd() tea.Cmd {
	return tea.Batch(
		tea.Tick(3000*time.Millisecond, func(t time.Time) tea.Msg {
			return PollMsg(t)
		}),
		HeartbeatCmd(),
	)
}

func runFastbootCmd(args ...string) ([]byte, error) {
	cmdPath, err := sysutil.ResolveCommandPath("fastboot")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, cmdPath, args...).CombinedOutput()
}

func runAdbCmd(args ...string) ([]byte, error) {
	cmdPath, err := sysutil.ResolveCommandPath("adb")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, cmdPath, args...).CombinedOutput()
}

func CheckDeviceState() tea.Msg {
	if !isScanning.CompareAndSwap(false, true) {
		return SkipUpdateMsg{}
	}
	defer isScanning.Store(false)

	state := domain.DeviceState{Mode: domain.ModeDisconnected, Serial: "-", Model: "-", Battery: "-", Slot: "-", Secure: "-"}

	if out, err := runFastbootCmd("devices"); err == nil && len(out) > 0 {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			state.Mode, state.Serial = domain.ModeFastboot, parts[0]
			state = fetchFastbootDetails(state)
			return DeviceUpdateMsg(state)
		}
	}

	if out, err := runAdbCmd("devices"); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, l := range lines {
			p := strings.Fields(l)
			if len(p) >= 2 {
				serial, status := p[0], p[1]
				switch status {
				case "device":
					state.Mode, state.Serial = domain.ModeDevice, serial
					state = fetchAdbDetails(state)
					return DeviceUpdateMsg(state)
				case "sideload":
					state.Mode, state.Serial = domain.ModeSideload, serial
					state.Model = "SIDELOAD DEVICE"
					return DeviceUpdateMsg(state)
				case "recovery":
					state.Mode, state.Serial = domain.ModeRecovery, serial
					state = fetchAdbDetails(state)
					return DeviceUpdateMsg(state)
				case "unauthorized":
					state.Mode, state.Serial = domain.ModeUnauthorized, serial
					state.Model = "ACTION REQUIRED"
					return DeviceUpdateMsg(state)
				case "offline":
					state.Mode, state.Serial = domain.ModeOffline, serial
					state.Model = "OFFLINE"
					return DeviceUpdateMsg(state)
				}
			}
		}
	}

	return DeviceUpdateMsg(state)
}

func fetchFastbootDetails(s domain.DeviceState) domain.DeviceState {
	out, err := runFastbootCmd("-s", s.Serial, "getvar", "all")
	if err != nil {
		return s
	}
	
	outputStr := string(out)
	
	cacheMu.RLock()
	cached, ok := deviceCache[s.Serial]
	cacheMu.RUnlock()

	if ok && cached.Model != "" && cached.Model != "-" {
		s.Model = cached.Model
		s.Secure = cached.Secure
	} else {
		if m := reProduct.FindStringSubmatch(outputStr); len(m) > 1 {
			codename := strings.ToUpper(strings.TrimSpace(m[1]))
			pretty := prettyDeviceName(codename)
			if pretty != codename {
				s.Model = fmt.Sprintf("%s (%s)", pretty, codename)
			} else {
				s.Model = codename
			}
		}

		isUnlocked := false
		if strings.Contains(strings.ToLower(outputStr), "unlocked: yes") || strings.Contains(strings.ToLower(outputStr), "device-state: unlocked") {
			isUnlocked = true
		}

		if isUnlocked {
			s.Secure = "NO"
		} else {
			if m := reSecure.FindStringSubmatch(outputStr); len(m) > 1 {
				if strings.ToLower(m[1]) == "yes" {
					s.Secure = "YES"
				} else {
					s.Secure = "NO"
				}
			} else {
				s.Secure = "YES"
			}
		}

		cacheMu.Lock()
		deviceCache[s.Serial] = domain.DeviceState{Model: s.Model, Secure: s.Secure}
		cacheMu.Unlock()
	}

	if m := reBattery.FindStringSubmatch(outputStr); len(m) > 1 {
		val := strings.TrimSpace(m[1])
		if val != "" && val != "-" {
			s.Battery = val + " mV"
		} else {
			s.Battery = val
		}
	}

	return s
}

func fetchAdbDetails(s domain.DeviceState) domain.DeviceState {
	cacheMu.RLock()
	cached, ok := deviceCache[s.Serial]
	cacheMu.RUnlock()

	if ok && cached.Model != "" && cached.Model != "-" {
		s.Model = cached.Model
		s.Secure = cached.Secure
	} else {
		out, _ := runAdbCmd("-s", s.Serial, "shell", "getprop ro.product.marketname; getprop ro.product.model; getprop ro.product.brand; getprop ro.product.device")
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for i := range lines { lines[i] = strings.TrimSpace(lines[i]) }
		for len(lines) < 4 { lines = append(lines, "") }
		marketProp, modelProp, brandProp, codenameProp := lines[0], lines[1], lines[2], strings.ToUpper(lines[3])

		marketingLabel := ""
		if out, err := runAdbCmd("devices", "-l"); err == nil {
			devLines := strings.Split(string(out), "\n")
			for _, l := range devLines {
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

		name := ""
		candidates := []string{marketProp, modelProp, marketingLabel, codenameProp}
		for _, c := range candidates {
			if c == "" { continue }
			pretty := prettyDeviceName(c)
			if pretty != strings.ToUpper(c) {
				name = pretty
				break
			}
		}

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
		s.Secure = "YES"

		cacheMu.Lock()
		deviceCache[s.Serial] = domain.DeviceState{Model: s.Model, Secure: s.Secure}
		cacheMu.Unlock()
	}

	if out, err := runAdbCmd("-s", s.Serial, "shell", "dumpsys", "battery"); err == nil {
		if m := reAdbBatt.FindStringSubmatch(string(out)); len(m) > 1 {
			s.Battery = m[1] + "%"
		}
	}

	return s
}
