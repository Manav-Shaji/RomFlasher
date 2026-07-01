package platform

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

type DeviceState struct {
	Mode    DeviceMode
	Serial  string
	Model   string
	Battery string
	Slot    string
	Secure  string
}
