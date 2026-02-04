package main

import (
    "log"
    "strings"
    "sync"

    "github.com/fun3md/GoMidiBinder/internal/midi"
)

type DeviceManager struct {
    mu           sync.Mutex
    activeInputs map[string]func()
    activeOutputs map[string]midi.Sender
    grandmaIn    midi.Sender
    grandmaOut   midi.Receiver
    cfg          Config
}

func NewDeviceManager(cfg Config, grandmaIn midi.Sender, grandmaOut midi.Receiver) *DeviceManager {
    return &DeviceManager{
        activeInputs:  make(map[string]func()),
        activeOutputs: make(map[string]midi.Sender),
        grandmaIn:     grandmaIn,
        grandmaOut:    grandmaOut,
        cfg:           cfg,
    }
}

func (m *DeviceManager) matchesConfig(name string) bool {
    lname := strings.ToLower(name)
    for _, k := range m.cfg.Devices {
        if strings.Contains(lname, strings.ToLower(k)) {
            return true
        }
    }
    return false
}

func (m *DeviceManager) isInputConnected(name string) bool {
    m.mu.Lock()
    defer m.mu.Unlock()
    _, ok := m.activeInputs[name]
    return ok
}

func (m *DeviceManager) addInput(name string, stop func()) {
    m.mu.Lock()
    m.activeInputs[name] = stop
    m.mu.Unlock()
}

func (m *DeviceManager) removeInput(name string) {
    m.mu.Lock()
    if stop, ok := m.activeInputs[name]; ok {
        stop()
        delete(m.activeInputs, name)
    }
    m.mu.Unlock()
}

func (m *DeviceManager) BroadcastToOutputs(msg midi.Message) {
    m.mu.Lock()
    defer m.mu.Unlock()
    for n, out := range m.activeOutputs {
        if out == nil {
            log.Printf("output %s has nil sender", n)
            continue
        }
        if err := out.Send(msg); err != nil {
            log.Printf("failed to send to %s: %v", n, err)
        }
    }
}

func (m *DeviceManager) ConnectInput(port midi.InPort) {
    name := port.Name()
    stop, err := midi.ListenTo(port, func(msg midi.Message, t int32) {
        if m.grandmaIn != nil {
            if err := m.grandmaIn.Send(msg); err != nil {
                log.Printf("send to GrandMA failed: %v", err)
            }
        }
    }, midi.UseSysEx())
    if err != nil {
        log.Printf("listen failed for %s: %v", name, err)
        return
    }
    m.addInput(name, stop)
    log.Printf("Connected Input: %s", name)
}

func (m *DeviceManager) ConnectOutputByName(name string) {
    // Try to find an out port with this name and store it as a Sender
    p, err := midi.FindOutPort(name)
    if err != nil {
        log.Printf("could not find out port %s: %v", name, err)
        return
    }
    // p implements midi.Sender
    m.mu.Lock()
    m.activeOutputs[name] = p.(midi.Sender)
    m.mu.Unlock()
    log.Printf("Connected Output: %s", name)
}
