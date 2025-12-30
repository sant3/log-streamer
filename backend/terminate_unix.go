//go:build !windows

package main

import (
    "os"
    "syscall"
)

// terminateSelf sends SIGTERM to the current process on Unix-like systems.
func terminateSelf() {
    pid := os.Getpid()
    // ignore error; if it fails, process will continue until external stop
    _ = syscall.Kill(pid, syscall.SIGTERM)
}


