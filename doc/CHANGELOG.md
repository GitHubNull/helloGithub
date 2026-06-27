# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/lang/zh-CN/spec/v2.0.0.html).

## [Unreleased]

### Added

- GitHub Actions automated release workflow for multi-platform builds.
- CHANGELOG documentation and README link.

## [0.2.0] - 2026-06-27

### Added

- Complete project README with feature overview and architecture design.
- Detailed project structure documentation.
- Quick start guides for Windows and Linux/macOS.
- Configuration reference for `config.yaml`.
- System DNS setup instructions.
- Technology stack and future roadmap sections.

## [0.1.0] - 2026-06-27

### Added

- Initial GitHub Fast DNS service implementation.
- DNS proxy server with GitHub domain interception.
- IP scanner with latency-based optimal IP selection.
- GitHub official IP range fetching via `https://api.github.com/meta`.
- Scheduler for periodic IP scanning and pool updates.
- ServiceManager for unified lifecycle management (GUI integration ready).
- YAML-based configuration support.
- Cross-platform build scripts (`build.bat`, `build.sh`, `run.bat`, `run.sh`).
- Structured logging with `log/slog`.
- Layered `.gitignore` configuration for Go project.

