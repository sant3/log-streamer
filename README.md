## Overview

**Log Streamer** is a web-based tool for streaming and viewing log files from a server in real-time. It consists of two main components:

- **A Go backend:** A lightweight agent that monitors log files in a specified directory and serves them over an HTTP/S WebSocket connection.
- **A React frontend:** A user-friendly web interface that connects to the backend, lists available log files, and displays their content as a live stream.

For detailed information on each component, please refer to their respective READMEs:

- [Frontend README](./frontend/README.md)
- [Backend README](./backend/README.md)

## How to build for distribution

This project uses `make` to simplify the build and distribution process for both the frontend and backend components.

### Requirements

Before you begin, ensure you have the following tools installed:

*   **For all platforms:**
    *   `make`
    *   `tar` (or a compatible version)
    *   `openssl` (for certificate generation)
*   **For the frontend:**
    *   `pnpm` (tested on v10.24.0)
*   **For the backend:**
    *   Go (latest version recommended)

### Distribution Commands

-   `make dist-fe`: Builds the React frontend and packages it into a `streamer.tar.gz` archive located in `frontend/dist`.
-   `make dist-be`: Builds the Go backend for Linux, macOS, and Windows (64-bit). The resulting archives (`.tar.gz` or `.zip`) will be placed in `backend/dist`.

You can also build for a specific platform:
-   `make dist-be-linux`
-   `make dist-be-darwin`
-   `make dist-be-windows`

### TODO
* add BE for x32 ARCH
* finish create automatically self-signed certs for agent
