# Architecture Guide

VoidFlasher PRIME uses a layered architecture strictly isolating the User Interface from the Business Logic and external command executions.

## Package Layout
- `cmd/flashtool`: The application entry point and dependency injector.
- `internal/ui`: The Bubble Tea model handling the terminal rendering (Cyberpunk aesthetic).
- `internal/flasher`: The orchestrator containing the **Safety Engine**, the **Flashing State Machine**, and the **Single Operation Lock**.
- `internal/android`: Managed wrappers around ADB and Fastboot executions handling timeouts, retries, and disconnect recovery.
- `internal/device`: Hardware metadata extraction (Codename, A/B Slots, Dynamic Partitions).
- `internal/updater`: GitHub API poller comparing semantic versions.
- `internal/logger`: Structured, persistent JSON logger for comprehensive Audit Logs.
- `internal/version`: Build-time metadata tracker.

## Concurrency & Safety
Concurrency is heavily restricted to prevent device corruption. The `flasher` package utilizes a global `sync.Mutex` ensuring only one command pipeline is actively transmitting to the device. Any concurrent dispatch is immediately rejected with a controlled error.
