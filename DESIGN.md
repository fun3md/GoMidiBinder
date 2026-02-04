# GoMidiBinder — Design & Architektur

Dieses Dokument enthält die ausführliche Projektbeschreibung, Architektur, Anforderungen und Implementierungspläne für `GoMidiBinder`.

## Ziel
Replace MIDI-OX with a stable, compiled, lightweight CLI tool that aggregates multiple physical MIDI controllers into a single stream for GrandMA, handles feedback (LEDs), and auto-reconnects devices upon hot-plugging.

---

## 1. Architecture Overview

The application acts as a "Man-in-the-Middle" router.

The Loop (Data Flow):

1. Ingress (Controllers → GrandMA):
   - Detects multiple physical devices (e.g., Launchpad, APC Mini).
   - Merges all incoming signals.
   - Forwards them to a specific Target Virtual Port (connected to GrandMA).

2. Egress (GrandMA → Feedback):
   - Listens to a Source Virtual Port (output from GrandMA).
   - Broadcasts that signal to all connected physical devices (to trigger LED feedback/motor faders).

3. The Watcher (Hot-plug Logic):
   - Runs in a background goroutine.
   - Scans for devices matching specific names every X seconds.
   - Automatically initializes connections for new devices.
   - Cleanly tears down connections for removed devices.

---

## 2. Requirements & Specifications

### Functional Requirements

- OS Support: Primary: Windows (GrandMA onPC). Secondary: macOS/Linux.
- Device Aggregation: Must support N input devices.
- Device filtering: Identify devices by Name Substring rather than Port ID.
- Auto-Reconnect: Re-detect within ~3 seconds after unplug/replug.
- Stability: Must not crash on malformed data or abrupt disconnects.
- Latency: Near-zero overhead.

### The Virtual Cable Issue (Windows)

Go cannot create virtual MIDI devices on Windows without kernel drivers — rely on loopMIDI. Create two ports (e.g. `GrandMA_Input`, `GrandMA_Output`) and connect the app to them.

---

## 3. Implementation Plan

### Tech Stack

- Language: Go (1.20+)
- Library: `gitlab.com/gomidi/midi/v2`
- Driver: `gitlab.com/gomidi/midi/v2/drivers/rtmididrv` (native RtMidi bindings)

### Configuration (`config.yaml`)

Example:

```yaml
target_app_in: "GrandMA_Input"
target_app_out: "GrandMA_Output"
devices:
  - "Launchpad"
  - "APC Mini"
  - "X-Touch"
```

### Core Components

- DeviceManager: verwaltet aktive Verbindungen (Inputs/Outputs), GrandMA Sender/Receiver und Config.
- Watcher: periodisch verfügbare Ports prüfen, neu verbinden, trennen.
- Router: Input-Callbacks leiten Nachrichten an GrandMA weiter; GrandMA-Feedback wird an alle physischen Outputs gesendet.

### Hot-plug & Robustness

- Nutze `midi.ListenTo` Callback-Fehler, um abgebrochene Verbindungen zu erkennen und aufzuräumen.
- Führe Connect/Disconnect-Operationen threadsicher via Mutex aus.
- Filtere Ports per Name-Substring (case-insensitive).

---

## 4. Build & Deployment

- Stub/CI: Standardmäßig verwendet das Repo einen Stub, damit `go test` ohne native Abhängigkeiten läuft.
- Native: Build mit `-tags native` aktiviert `rtmididrv`.

Linux Beispiel (native):

```bash
sudo apt install -y build-essential pkg-config libasound2-dev
go build -tags native -o GoMidiBinder ./...
```

Windows: installiere loopMIDI und MinGW-w64 für CGO, baue auf Windows oder cross-compile mit geeigneten Toolchains/CI.

---

## 5. Testing & Manual Plan

- Unit-tests: Config loader, DeviceManager helpers.
- Integration: Build with native driver, connect loopMIDI ports, attach real devices and verify message forwarding and feedback.

---

## 6. Example Snippet

Siehe `main.go` im Repository für ein kurzes Startbeispiel.

---

Dieses Dokument kann als Referenz für die Entwicklung und spätere Ergänzung von Details wie CLI-Flags, Logging-Levels und CI-Release-Konfiguration dienen.
