package core

type FlashState int

const (
	StateIdle FlashState = iota
	StateDetecting
	StateValidating
	StatePreparing
	StateFlashing
	StateVerifying
	StateRebooting
	StateCompleted
	StateFailed
)

func (s FlashState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateDetecting:
		return "Detecting Device"
	case StateValidating:
		return "Validating"
	case StatePreparing:
		return "Preparing"
	case StateFlashing:
		return "Flashing"
	case StateVerifying:
		return "Verifying"
	case StateRebooting:
		return "Rebooting"
	case StateCompleted:
		return "Completed"
	case StateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}
