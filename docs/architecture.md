# Architecture Guide

NexForge uses a minimal, idiomatic Go architecture strictly isolating the User Interface from the Business Logic and external command executions. 

## Core Principles
1. **Separation of Concerns:** UI code (`tui`), business orchestration (`core`), and infrastructure/configuration (`config`) must never tightly couple.
2. **Context-Driven Execution:** All background operations must use `context.Context` to allow clean timeouts and user cancellations.
3. **Platform Independence:** Binary paths (ADB/Fastboot) are dynamically resolved checking embedded payloads, absolute paths, and the system `PATH`.

## Package Layout
- `cmd/nexforge`: The application entry point and Cobra CLI commands (`root`, `flash`, `devices`, `version`).
- `internal/config`: Application configuration, structured logging initialization, and the central dependency injection `App` container. Provides strict environment variable and file validation upon load.
- `internal/core`: The business orchestration layer. Executes processes, tracks long-running commands, and hosts `FlashService` and `DeviceService`. Contains the robust `ValidationService` that enforces safety rules before any flash operation, and defines the explicit `FlashState` state machine.
- `internal/platform`: Low-level integrations. Contains device scanning logic, device state types, binary resolution, and embedded payloads. All system commands are wrapped with context propagation, timeouts, and intelligent retry logic.
- `internal/tui`: The Bubble Tea framework rendering the TUI. Managed in modular files (`model.go`, `view.go`, `update.go`, `handlers.go`). Views react natively to the `FlashState` engine instead of disparate boolean flags.

## Concurrency & Safety
Concurrency is heavily restricted to prevent device corruption. The `core.Engine` encapsulates a dedicated `sync.Mutex` alongside the cancellation context. This ensures that only one flash/ADB operation is actively executing and transmitting logs to the UI at any given time.

TUI rendering performance is protected by debouncing mechanisms, guaranteeing steady 10 FPS log updates during aggressive outputs instead of blocking the main Bubble Tea loop.
