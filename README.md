# VoidFlasher PRIME

VoidFlasher PRIME is a production-grade, highly reliable Android flashing terminal utility. Built with Go, Bubble Tea, and strict safety validation at its core, it brings OEM-level safety checks into a modern cyberpunk-themed terminal UI.

## Features
- **Un-bypassable Safety Engine**: Performs rigorous pre-flight checks (connection, device mode, battery level, SHA256 validation) before executing *any* command.
- **Flashing State Machine**: Enforces strict operational transitions ensuring no concurrent flash operations can corrupt your device.
- **Cyberpunk UI**: Built on Charm's Bubble Tea, offering a split-pane dashboard with Device Info, Live Logs, and an actionable Queue.
- **Reliable Executions**: All ADB and Fastboot commands run via managed interfaces with context timeouts, automatic retries, and disconnect recovery handling.
- **Audit Logging**: Generates persistent structured JSON logs for every operation to guarantee traceability for troubleshooting.

## Installation
Download the latest pre-compiled binaries for Windows, Linux, or macOS from the [Releases page](https://github.com/Manav-Shaji/RomFlasher/releases).

## Build Instructions
1. Ensure you have Go 1.25.6+ installed.
2. Clone the repository: `git clone https://github.com/Manav-Shaji/RomFlasher.git`
3. Build the binary:
   ```sh
   go build -o VoidFlasher ./cmd/flashtool
   ```

## Development
Please see [CONTRIBUTING.md](CONTRIBUTING.md) and the [Architecture Guide](docs/architecture.md) for detailed workflows.
