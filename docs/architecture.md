# Architecture Guide

NexForge uses a layered architecture strictly isolating the User Interface from the Business Logic and external command executions. 

## Core Principles
1. **Separation of Concerns:** UI code (`tui`), business orchestration (`engine`), and CLI interfaces (`cli`) must never tightly couple.
2. **Context-Driven Execution:** All background operations must use `context.Context` to allow clean timeouts and user cancellations.
3. **Platform Independence:** Binary paths (ADB/Fastboot) are dynamically resolved via `sysutil` checking embedded payloads, absolute paths, and the system `PATH`.

## Package Layout
- `cmd/nexforge`: The application entry point and root binary path.
- `internal/app`: Application container to perform dependency injection (DI) across all modules.
- `internal/cli`: Cobra-based command-line interface handling (`root`, `flash`, `devices`, `version`).
- `internal/config`: Application settings and JSON configuration persistence.
- `internal/domain`: Core domain models, logging abstractions, and entity definitions (`DeviceState`, `LogEntry`).
- `internal/engine`: The business orchestration layer. Executes processes, tracks long-running commands, and hosts `services` like `FlashService` and `DeviceService`.
- `internal/logger`: Centralized structured logging powered by Uber's Zap.
- `internal/platform`: Low-level integrations. Contains `adb` (device scanning & parsers) and `sysutil` (binary resolution & embedded payloads).
- `internal/tui`: The Bubble Tea framework rendering the Cyberpunk terminal interface. Managed in modular files (`model.go`, `view.go`, `update.go`, `handlers.go`).

## Concurrency & Safety
Concurrency is heavily restricted to prevent device corruption. The `engine.Engine` encapsulates a dedicated `sync.Mutex` alongside the cancellation context (`activeCmdCancel`). This ensures that only one flash/ADB operation is actively executing and transmitting logs to the UI at any given time.

TUI rendering performance is protected by debouncing mechanisms (`LogTickMsg`), guaranteeing steady 10 FPS log updates during aggressive outputs instead of blocking the main Bubble Tea loop.
