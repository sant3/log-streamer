//go:build windows

package main

import (
    "os"
    "os/signal"
)

// setupSignalHandler registers signals on Windows (no SIGTERM available).
func setupSignalHandler(c chan os.Signal) {
    signal.Notify(c, os.Interrupt)
}


