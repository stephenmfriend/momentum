# Flux + Momentum = ❤️ 
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE) ![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)

> [!WARNING]
> This repository is under construction and not ready for production use.

The perfect companion to Flux. Because once the board starts moving, it shouldn’t stop.

## Features

### Interactive TUI Mode
- **Three-pane layout** - Browse Projects, Epics, and Tasks in a unified view
- **Vim-style navigation** - Use `j/k` or arrow keys, `Tab`/`Shift+Tab` to switch panes
- **Task filtering** - Toggle between All, Todo, In Progress, and Done with `f`
- **Multi-select** - Select multiple tasks with `Space` and batch update them
- **Visual status indicators** - Color-coded task states (blocked, in-progress, done)
- **Real-time updates** - SSE subscription with automatic polling fallback
- **Search** - Filter items within any pane with `/`

### Headless Mode
- **Smart task selection** - Automatically picks the newest unblocked todo task
- **Flexible filtering** - Filter by `--project`, `--epic`, or `--task`
- **CI/CD friendly** - Perfect for automation pipelines and scripting

### Workflow Operations
- **Batch status transitions** - Start, complete, or reset multiple tasks at once
- **Dependency awareness** - Blocked tasks are visually distinguished

### Flux Integration
- Full REST client for Projects, Epics, and Tasks
- Real-time sync via Server-Sent Events (SSE)
