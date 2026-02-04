//go:build native
// +build native

package main

import (
    "log"

    "gitlab.com/gomidi/midi/v2"
    _ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

// InitDriver ensures the native rtmidi driver is available. The driver registers on import.
func InitDriver() error {
    return nil
}

func FindVirtualPorts(cfg Config) (midi.Sender, midi.Receiver, error) {
    outToGrandMA, err := midi.FindOutPort(cfg.TargetAppIn)
    if err != nil {
        return nil, nil, err
    }
    inFromGrandMA, err := midi.FindInPort(cfg.TargetAppOut)
    if err != nil {
        return nil, nil, err
    }
    return outToGrandMA.(midi.Sender), inFromGrandMA.(midi.Receiver), nil
}

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
