//go:build windows

package main

import (
    "os"
)

// terminateSelf ends the current process after spawning the replacement.
func terminateSelf() {
    os.Exit(0)
}


