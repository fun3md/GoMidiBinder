# Manual Test Plan — GoMidiBinder

Ziel: Verifizieren, dass Eingabegeräte korrekt erkannt werden, MIDI-Events an GrandMA weitergeleitet werden, Feedback vom GrandMA an alle physikalischen Geräte gesendet wird und Hot-plugging stabil funktioniert.

Voraussetzungen
- Physische MIDI-Controller (z. B. Launchpad, APC Mini)
- loopMIDI (Windows) oder IAC/Audio-MIDI (macOS) für virtuelle Ports
- Auf Linux: `libasound2-dev` + Toolchain für native Builds (falls Integrationstests mit nativer RtMidi gewünscht)

Dateien
- `config.example.yaml` → kopieren nach `config.yaml` und anpassen

1) Unit / CI Checks
- `go test ./...` sollte ohne native Header erfolgreich laufen (stub-Modus).

2) Smoke Test (Stub)
- Schritte:
  1. `go build -o GoMidiBinder ./...`
  2. `./GoMidiBinder -config config.yaml`
  3. Prüfen: Logs zeigen Watcher-Aktivität; keine nativen Verbindungen erwartet.

3) Integrationstest (mit loopMIDI + native Treiber)
- Setup (Windows):
  - Installiere loopMIDI und erstelle zwei Ports: `GrandMA_Input`, `GrandMA_Output`.
  - Schließe physische Controller an.
  - Installiere MinGW-w64 oder entsprechenden GCC für CGO.

- Build:
```bash
sudo apt install -y build-essential pkg-config libasound2-dev   # Linux
go build -tags native -o GoMidiBinder ./...
```

- Lauf:
  1. Starte `GoMidiBinder -config config.yaml`.
  2. Verbinde GrandMA (oder ein Monitor-Tool) zu `GrandMA_Input` und `GrandMA_Output`.
  3. Sende MIDI-Events von einem physischen Gerät — verifiziere, dass GrandMA die Events empfängt.
  4. Sende Feedback (LEDs) aus GrandMA — verifiziere, dass alle physischen Geräte die Rückmeldung bekommen.

4) Hot-plug Test
- Schritte:
  1. Starte das Programm mit angeschlossenen Geräten.
  2. Ziehe ein Gerät ab — beobachte Logs, ob ein Disconnect erkannt wird und der Input aus `activeInputs` entfernt wird.
  3. Stecke das Gerät wieder ein — prüfe, dass es innerhalb ~3 Sekunden wieder verbunden wird.

5) Fehler- und Stabilitätstests
- Simuliere fehlerhafte/fragmentierte SysEx-Nachrichten, lange MIDI-Streams und abrupte Disconnects.
- Überprüfe, dass das Programm nicht crasht und nur fehlerhafte Verbindungen bereinigt.

6) Logging & Diagnostik
- Überprüfe Log-Level und Einträge: Connect, Disconnect, Send-Fehler.
- Optional: erweitere `main.go` um ein `--log-level` Flag für `debug`/`info`.

7) Abschluss-Checks
- Alle Tests dokumentieren (Datum, Hardware, Ergebnis).
- Bei Problemen: relevante Logs, `go env`, `uname -a`, und `apport`/`dmesg` Output sammeln.
