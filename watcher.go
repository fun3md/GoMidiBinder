package main

import (
    "log"
    "time"

    "github.com/fun3md/GoMidiBinder/internal/midi"
)

func (m *DeviceManager) Watch(interval time.Duration, stopCh <-chan struct{}) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            inPorts := midi.GetInPorts()
            for _, p := range inPorts {
                name := p.Name()
                if name == m.cfg.TargetAppOut || name == m.cfg.TargetAppIn {
                    // skip virtual cables
                    continue
                }
                if m.matchesConfig(name) && !m.isInputConnected(name) {
                    m.ConnectInput(p)
                }
            }

            // Ensure outputs for feedback: connect any matching Out ports
            outPorts := midi.GetOutPorts()
            for _, p := range outPorts {
                name := p.Name()
                if name == m.cfg.TargetAppOut || name == m.cfg.TargetAppIn {
                    continue
                }
                if m.matchesConfig(name) {
                    // connect if not present
                    m.mu.Lock()
                    _, ok := m.activeOutputs[name]
                    m.mu.Unlock()
                    if !ok {
                        m.ConnectOutputByName(name)
                    }
                }
            }

        case <-stopCh:
            log.Println("Watcher stopping")
            return
        }
    }
}
