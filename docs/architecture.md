# Architecture Guide

NexForge uses a layered architecture strictly isolating the User Interface from the Business Logic and external command executions.

## Package Layout
- `cmd/flashtool`: The application entry point and dependency injector.
- `internal/config`: Application settings, dynamic path generation, and JSON configuration persistence.
- `internal/core`: The business logic layer. Contains the `Engine` for executing and managing subprocesses (ADB/Fastboot) with context timeouts, log piping, and strict cancellation policies. Also handles device state polling and safety validation.
- `internal/tui`: The Bubble Tea framework rendering the Cyberpunk terminal interface. Sub-packages include layout logic, handlers, and the cohesive `theme` package for cohesive styling.

## Concurrency & Safety
Concurrency is heavily restricted to prevent device corruption. The `core.Engine` encapsulates a dedicated `sync.Mutex` alongside the cancellation context (`activeCmdCancel`). This ensures that only one flash/ADB operation is actively executing and transmitting logs to the UI at any given time. Any user attempt to initiate concurrent dispatch is structurally prevented or cleanly rejected.
