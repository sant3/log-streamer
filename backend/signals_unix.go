//go:build !windows

package main

import (
    "os"
    "os/signal"
    "syscall"
)

// setupSignalHandler registers interrupt and SIGTERM on Unix-like systems.
func setupSignalHandler(c chan os.Signal) {
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
}


