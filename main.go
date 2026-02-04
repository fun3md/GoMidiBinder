package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/fun3md/GoMidiBinder/internal/midi"
)

var configPath string

func init() {
    flag.StringVar(&configPath, "config", "config.yaml", "path to config file")
}

func main() {
    flag.Parse()
    log.Println("GoMidiBinder starting")

    cfg, err := LoadConfig(configPath)
    if err != nil {
        log.Printf("could not load config (%s): %v", configPath, err)
        fmt.Println("See config.example.yaml for an example.")
    }

    if err := InitDriver(); err != nil {
        log.Fatalf("failed to init midi driver: %v", err)
    }

    grandmaIn, grandmaOut, err := FindVirtualPorts(cfg)
    if err != nil {
        log.Fatalf("could not find virtual GrandMA ports: %v", err)
    }

    dm := NewDeviceManager(cfg, grandmaIn, grandmaOut)

    // start watcher
    stopCh := make(chan struct{})
    go dm.Watch(2*time.Second, stopCh)

    // listen to GrandMA feedback
    if err := ListenGrandMAFeedback(dm); err != nil {
        log.Printf("warning: feedback listener not started: %v", err)
    }

    // graceful shutdown on SIGINT/SIGTERM
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
    close(stopCh)
    midi.CloseDriver()
    log.Println("GoMidiBinder stopped")
}
