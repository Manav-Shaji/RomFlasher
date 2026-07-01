package core

import "time"

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
