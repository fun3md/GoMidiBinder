//go:build !native
// +build !native

package main

import (
    "log"

    "github.com/fun3md/GoMidiBinder/internal/midi"
)

// InitDriver is a no-op in stub mode. To enable the real driver, build with `-tags native`.
func InitDriver() error {
    return nil
}

// FindVirtualPorts locates the GrandMA virtual ports (stubbed — uses the midi API without forcing a native driver).
func FindVirtualPorts(cfg Config) (midi.Sender, midi.Receiver, error) {
    outToGrandMA, err := midi.FindOutPort(cfg.TargetAppIn)
    if err != nil {
        return nil, nil, err
    }
    inFromGrandMA, err := midi.FindInPort(cfg.TargetAppOut)
    if err != nil {
        return nil, nil, err
    }
    return outToGrandMA, inFromGrandMA, nil
}

// ListenGrandMAFeedback sets up a listener on the GrandMA output and broadcasts to devices.
func ListenGrandMAFeedback(dm *DeviceManager) error {
    if dm.grandmaOut == nil {
        return nil
    }
    _, err := midi.Listen(dm.grandmaOut, func(msg midi.Message, t int32) {
        dm.BroadcastToOutputs(msg)
    }, midi.UseSysEx())
    if err != nil {
        log.Printf("failed to listen to GrandMA out: %v", err)
        return err
    }
    return nil
}
