# ⚡ VoidFlasher PRIME

> 🌌 **Cyberpunk Edition v1.2**

A high-performance, professional Android Flasher TUI ported to Pure Go. Built with the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework and a vibrant **Synthwave/Cyberpunk** design system.

---

## 🚀 Key Features

- **🎨 Cyberpunk TUI Aesthetic**: A stunning high-contrast theme featuring Neon Pink, Cyan, and Indigo accents for a premium terminal experience.
- **🌈 Colorful Log Engine**: Advanced logstraming with keyword-based syntax highlighting:
  - `adb` commands in **Cyan**
  - `fastboot` commands in **Purple**
  - Actions (`flash`, `wipe`, `sideload`) in **Warning Gold**
  - Real-time status markers (`[ DONE ]`, `[ FAILED ]`)
- **📡 Live Logstream**: Minimalist, real-time command output without the clutter—focusing on raw execution speed and visibility.
- **🛰️ Premium HUD Interface**: Real-time device tracking (Model, Battery, Slot, Security) via a high-density Head-Up Display.
- **⌨️ Command Console**: A professional modal-based console for direct `adb` and `fastboot` execution with real-time log history.
- **📂 Intelligent File Picker**: Built-in explorer with automatic filtering (e.g., `.img` for flashing, `.zip` for sideloading).
- **🛡️ Confirmation Guards**: Destructive actions (wipe, flash) require explicit user confirmation to prevent accidental data loss.

---

## 📂 Project Structure

```text
flashtool/
├── cmd/flashtool/main.go     # Application Bootloader
├── internal/
│   ├── model.go              # Shared Application State
│   ├── update.go             # Central Event Dispatcher
│   ├── handlers.go           # Interaction & Logic Handlers
│   ├── view.go               # Top-level Layout Rendering
│   ├── view_modals.go        # Modal & Overlay Components
│   ├── view_panels.go        # Device HUD & Status Panels
│   ├── device.go             # Hardware Polling & Recognition
│   ├── commands.go           # Subprocess Engine & Log Parser
│   └── ui/styles.go          # Neon Design Tokens & Theme
└── config.json               # External Configuration
```

---

## 🛠️ Build & Development

### 📋 Prerequisites

- **Go 1.19+**
- `adb` and `fastboot` must be in your system `PATH`.

### 🔨 Build Commands

1. **Optimize Dependencies**:
   ```bash
   go mod tidy
   ```
2. **Compile Global Binary**:
   ```bash
   go build -o VoidFlasher.exe ./cmd/flashtool
   ```
3. **Internal Development Run**:
   ```bash
   go run ./cmd/flashtool
   ```

---

## 🧠 Architecture Evolution

Originally ported from a monolithic PowerShell script, this version introduces **Crescent Logic**:

| Feature          | PowerShell Legacy          | Go Core Optimization        |
| :--------------- | :------------------------- | :-------------------------- |
| **Logic**        | Procedural / Imperative    | Reactive (Elm Architecture) |
| **Threading**    | Single-threaded (Blocking) | Multi-core (Concurrent)     |
| **UI Rendering** | Selective Repainting       | Functional Layout Flow      |
| **Theme**        | Standard Console Colors    | **Cyberpunk / Synthwave**   |
| **Performance**  | High Overhead              | Near-Native (Static Binary) |

---

## 🛡️ Safety & Stability

- **Active Path Validation**: Tool checks for binary existence before execution.
- **Connection Heartbeat**: Pulsating visual indicator (`○` → `●`) in the status bar ensures the UI thread is responsive.
- **Auto-Refresh**: Background hardware polling keeps device information updated without manual intervention.

---

_Built with ❤️ for the Android Community by Antigravity AI._
