# Flux + Momentum + Claude Code = ❤️ 
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/sirsjg/momentum)](https://github.com/stevegrehan/momentum/releases)
![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![macOS](https://img.shields.io/badge/macOS-000000?style=flat&logo=apple&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?style=flat&logo=linux&logoColor=black)

> [!WARNING]
> This tool is experimental and not ready for production use. 

The perfect companion to Flux. Because once the board starts moving, it shouldn’t stop.

## Features

> [!NOTE]
> Currently only Claude Code is supported. Future releases will add support for other agents such as Codex.

### Agent Orchestration
- **Automatic task execution** - Watches for tasks and spawns Claude Code agents automatically
- **Async & sync modes** - Run multiple agents in parallel or sequentially (`--execution-mode`)
- **Graceful cancellation** - Stop agents cleanly with SIGINT handling

### Terminal UI
- **Multi-panel dashboard** - Monitor multiple running agents simultaneously
- **Real-time output streaming** - Watch agent progress with parsed JSON output
- **Keyboard navigation** - Tab between panels, scroll with j/k, stop/close agents
- **Auto-update notifications** - Get notified when new versions are available

### Flux Integration
- **Smart task selection** - Automatically picks unblocked todo tasks from auto-enabled epics
- **Flexible filtering** - Filter by `--project`, `--epic`, or `--task`
- **Real-time sync** - Server-Sent Events (SSE) for instant task updates
- **Workflow automation** - Automatic status transitions (todo → in_progress → done)

## Install

### Homebrew (macOS & Linux)

Requires [Homebrew](https://brew.sh) to be installed.

```bash
brew tap sirsjg/momentum
brew install momentum
```