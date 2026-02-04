### Project Name: `GoMidiBinder`
**Goal:** Replace MIDI-OX with a stable, compiled, lightweight CLI tool that aggregates multiple physical MIDI controllers into a single stream for GrandMA, handles feedback (LEDs), and auto-reconnects devices upon hot-plugging.

---

### 1. Architecture Overview

The application will act as a "Man-in-the-Middle" router.

**The Loop (Data Flow):**
1.  **Ingress (Controllers $\to$ GrandMA):**
    *   Detects multiple physical devices (e.g., Launchpad, APC Mini).
    *   Merges all incoming signals.
    *   Forwards them to a specific **Target Virtual Port** (connected to GrandMA).
2.  **Egress (GrandMA $\to$ Feedback):**
    *   Listens to a **Source Virtual Port** (output from GrandMA).
    *   Broadcasts that signal to **all** connected physical devices (to trigger LED feedback/motor faders).
3.  **The Watcher (Hot-plug Logic):**
    *   Runs in a background goroutine.
    *   Scans for devices matching specific names every $X$ seconds.
    *   Automatically initializes connections for new devices.
    *   Cleanly tears down connections for removed devices.

---

### 2. Requirements & Specifications

#### 2.1 Functional Requirements
*   **OS Support:** Primary: Windows (GrandMA onPC). Secondary: macOS/Linux.
*   **Device Aggregation:** Must support $N$ input devices.
*   **Device filtering:** Identify devices by **Name Substring** (e.g., "Launchpad", "NanoKorg") rather than Port ID (which changes on reboot).
*   **Auto-Reconnect:** If a USB cable is unplugged and replugged, the software must pick it up within 3 seconds without restarting.
*   **Stability:** Must not crash if a device sends malformed data or disconnects abruptly.
*   **Latency:** Near-zero overhead (pass-through).

#### 2.2 The "Virtual Cable" Issue (Windows Specific)
*   *Constraint:* Go cannot natively create a virtual MIDI device on Windows without writing a Kernel driver.
*   *Solution:* You will still need **loopMIDI** (by Tobias Erichsen) installed.
*   *Workflow:*
    1.  Create a loopMIDI port named `GrandMA_In`.
    2.  Create a loopMIDI port named `GrandMA_Out`.
    3.  **GoMidiBinder** connects to these existing ports.
    4.  GrandMA connects to these existing ports.
*   *Note:* On macOS/Linux, the OS allows creating virtual ports dynamically, but to keep the code unified, relying on external virtual ports is safer.

---

### 3. Implementation Plan

#### 3.1 Tech Stack
*   **Language:** Go (1.20+)
*   **Library:** `gitlab.com/gomidi/midi/v2`
*   **Driver:** `gitlab.com/gomidi/midi/v2/drivers/rtmididrv` (Stable C-binding wrapper)

#### 3.2 Configuration Structure (`config.yaml`)
We avoid hardcoding. The app should load a config:

```yaml
# Virtual ports created by loopMIDI (Windows) or IAC (Mac)
target_app_in: "GrandMA_Input"   # We send data HERE
target_app_out: "GrandMA_Output" # We read data from HERE

# List of hardware devices to auto-connect
# Uses partial string matching (case-insensitive)
devices:
  - "Launchpad"
  - "APC Mini"
  - "X-Touch"
```

#### 3.3 Core Components (Go Code Structure)

**A. The Manager Struct**
Holds the state of active connections to prevent double-connecting.

```go
type DeviceManager struct {
    activeInputs  map[string]func() // Map name -> stop function
    activeOutputs map[string]func() 
    grandmaIn     midi.Sender
    grandmaOut    midi.Receiver
    config        Config
}
```

**B. The Watcher Loop**
This is the heart of the "Anti-Crash" logic.

```go
func (m *DeviceManager) Watch(interval time.Duration) {
    ticker := time.NewTicker(interval)
    for range ticker.C {
        // 1. Get current OS ports
        inPorts := midi.GetInPorts()
        outPorts := midi.GetOutPorts()

        // 2. Check for configured devices that are present but NOT connected
        for _, port := range inPorts {
            if m.matchesConfig(port.Name()) && !m.isConnected(port.Name()) {
                m.connectInput(port)
            }
        }
        
        // 3. (Optional) Validate existing connections 
        // Note: rtmidi often handles the 'disconnect' error in the Read callback
    }
}
```

**C. The Router (Merge & Broadcast)**

```go
// Connect a physical controller input
func (m *DeviceManager) connectInput(port midi.InPort) {
    // Define the callback
    stop, err := midi.ListenTo(port, func(msg midi.Message, timestamp int32) {
        // FORWARD to GrandMA
        if m.grandmaIn != nil {
            m.grandmaIn.Send(msg)
        }
    }, midi.UseSysEx())
    
    if err == nil {
        m.activeInputs[port.Name()] = stop
        log.Printf("Connected Input: %s", port.Name())
    }
}

// Handle feedback from GrandMA
func (m *DeviceManager) handleGrandMAFeedback(msg midi.Message) {
    // BROADCAST to all known physical outputs
    for name, outPort := range m.activeOutputs {
        outPort.Send(msg)
    }
}
```

---

### 4. Step-by-Step Development Checklist

1.  **Setup Environment:**
    *   Install Go.
    *   Install `gcc` (Required for `rtmididrv` as it uses CGO). On Windows, install **TDM-GCC** or **MinGW-w64**.
    *   `go get gitlab.com/gomidi/midi/v2`
    *   `go get gitlab.com/gomidi/midi/v2/drivers/rtmididrv`

2.  **Prototype Phase (Proof of Concept):**
    *   Write a script that just lists ports (`midi.GetInPorts()`).
    *   Write a script that passes data from *one* specific hardcoded device to another.

3.  **The "GrandMA" Mock:**
    *   Before testing with the real desk, install loopMIDI.
    *   Create ports `GM_IN` and `GM_OUT`.
    *   Use MIDI-OX (ironically) just to monitor these ports to verify your Go app is sending data correctly.

4.  **Implement Hot-Plug Watcher:**
    *   Implement the Ticker.
    *   Test by unplugging the USB cable while the app runs.
    *   **Crucial:** Ensure you capture the error when a device disconnects. `gomidi/v2` might return an error in the listen callback; use that to delete the device from the `activeInputs` map so it can be rediscovered.

5.  **Build & Deploy:**
    *   Compile for Windows: `go build -ldflags -H=windowsgui -o MidiBinder.exe main.go` (Use `-H=windowsgui` to hide the console window if desired, or keep it for logs).

### 5. Why this fixes the problem

1.  **Dynamic ID Mapping:** We ignore ID numbers (which confuse MIDI-OX). We bind by **Name**. If Windows reassigns "Launchpad" from ID 2 to ID 5, `FindInPort("Launchpad")` still finds it.
2.  **Stateless Logic:** The router doesn't care "who" sent the note. It just pipes `Input[] -> GrandMA`.
3.  **Feedback Loop:** By sending GrandMA output to *all* devices, you ensure that if you page swap on the desk, the Launchpad *and* the APC update their LEDs simultaneously without complex routing rules.

### 6. Code Snippet to Start (main.go)

```go
package main

import (
	"log"
	"strings"
	"time"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // Registers the driver
)

var (
	targetDeviceIn  = "loopMIDI Port 1" // To GrandMA
	targetDeviceOut = "loopMIDI Port 2" // From GrandMA
	inputKeywords   = []string{"Launchpad", "APC", "Nano"} // Devices to find
)

func main() {
	defer midi.CloseDriver()
    
    // 1. Find the virtual cable to GrandMA
	outToGrandMA, err := midi.FindOutPort(targetDeviceIn)
	if err != nil {
		log.Fatalf("Could not find Virtual Cable to GrandMA: %v", err)
	}

    // 2. Start the watcher
	go deviceWatcher(outToGrandMA)

    // 3. Listen to GrandMA (Feedback loop) and broadcast to all
    // Note: You'll need logic here to find active outputs and send to them
    // This is simplified for brevity.
	
	select {} // Block forever
}

func deviceWatcher(grandMA midi.Sender) {
	connected := make(map[string]func())

	for {
		ports := midi.GetInPorts()
		for _, port := range ports {
            name := port.Name()
            
            // Skip the virtual cable itself to avoid feedback loops!
            if strings.Contains(name, targetDeviceOut) { continue }

            // Check against keywords
            isTarget := false
            for _, k := range inputKeywords {
                if strings.Contains(name, k) {
                    isTarget = true
                    break
                }
            }

			if isTarget {
				if _, exists := connected[name]; !exists {
					log.Printf("Found new device: %s", name)
					
                    // Connect and Merge
					stop, err := midi.ListenTo(port, func(msg midi.Message, t int32) {
						// Merge logic: Forward everything to GrandMA
						grandMA.Send(msg)
					}, midi.UseSysEx())

					if err == nil {
						connected[name] = stop
					}
				}
			}
		}
        
        // TODO: Add logic to remove disconnected devices from the map
		time.Sleep(2 * time.Second)
	}
}
```