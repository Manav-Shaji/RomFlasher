package domain

import "time"

type DeviceMode string

const (
	ModeDisconnected DeviceMode = "DISCONNECTED"
	ModeFastboot     DeviceMode = "FASTBOOT"
	ModeDevice       DeviceMode = "DEVICE"
	ModeRecovery     DeviceMode = "RECOVERY"
	ModeSideload     DeviceMode = "SIDELOAD"
	ModeUnauthorized DeviceMode = "UNAUTHORIZED"
	ModeOffline      DeviceMode = "OFFLINE"
)

type LogLevel string

const (
	LogInfo    LogLevel = "INFO"
	LogError   LogLevel = "ERROR"
	LogSuccess LogLevel = "SUCCESS"
)

type LogEntry struct {
	Level     LogLevel
	Text      string
	Timestamp time.Time
}

type DeviceState struct {
	Mode    DeviceMode
	Serial  string
	Model   string
	Battery string
	Slot    string
	Secure  string
}
