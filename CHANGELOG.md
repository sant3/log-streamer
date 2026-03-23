# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CI/CD pipeline with GitHub Actions (build, test, release)
- Backend refactoring: separated code into `config.go`, `middleware.go`, `handlers.go`
- Frontend refactoring: extracted `ServerPanel`, `LogViewer`, `LogControls` components
- Custom hooks: `useLogStream`, `useServerStatus`
- Comprehensive test suites for backend (60%+ coverage) and frontend components
- Linting configuration (`.golangci.yml`, `.prettierrc`)
- Community files: issue templates, PR template, CONTRIBUTING guide
- Godoc comments on all exported functions

### Changed
- Backend `main.go` reduced to entry point with wiring only
- Frontend `App.js` simplified using extracted components and hooks
- Middleware functions now accept explicit parameters instead of using global state

## [1.0.0] - 2024-12-01

### Added
- Real-time log file streaming via SSE (Server-Sent Events)
- Multi-server support with server status polling
- Log file autocomplete with keyboard navigation
- Text highlighting with match count
- Dark/light theme switcher
- Font size controls and line number toggle
- Auto-scroll functionality
- CORS middleware with configurable origins
- JWT authentication middleware
- IP whitelist middleware
- Directory traversal prevention (`.log` files only)
- TLS/HTTPS support with self-signed certificate generation
- Cross-platform builds (Linux, macOS, Windows)
- Runtime configuration via CLI flags and environment variables (`STREAMER_*` prefix)
- Multi-server configuration via `servers.js` without rebuild
