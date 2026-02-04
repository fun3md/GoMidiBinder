package main

import (
    "io/ioutil"
    "os"
    "testing"
)

func TestLoadConfig(t *testing.T) {
    content := []byte("target_app_in: GM_IN\ntarget_app_out: GM_OUT\ndevices:\n  - Launchpad\n")
    if err := ioutil.WriteFile("tmp_config.yaml", content, 0644); err != nil {
        t.Fatalf("write tmp config: %v", err)
    }
    defer os.Remove("tmp_config.yaml")

    cfg, err := LoadConfig("tmp_config.yaml")
    if err != nil {
        t.Fatalf("load config failed: %v", err)
    }
    if cfg.TargetAppIn != "GM_IN" {
        t.Fatalf("expected GM_IN, got %s", cfg.TargetAppIn)
    }
    if len(cfg.Devices) != 1 || cfg.Devices[0] != "Launchpad" {
        t.Fatalf("unexpected devices: %v", cfg.Devices)
    }
}
