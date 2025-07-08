# yatto - Yet Another Terrific Task Organizer

**yatto** is a terminal-based to-do application built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). It stores each task as
a separate JSON file on your filesystem and optionally manages the
task directory as a Git repository for versioning and synchronization.

## Features

- **TUI-based** interface powered by the Bubble Tea framework
- **Local file storage**: Each task is stored as an individual JSON file for easy inspection and portability
- **Basic task fields**: Title, description, and priority (High, Medium, Low)
- **Git integration**: Optionally initializes a Git repository in your task directory for:
  - Full version history of all tasks
  - Safe collaboration and backup
  - Sync across machines

## Installation

```bash
go install github.com/handlebargh/yatto@latest
```

## Task Storage

Tasks are saved in a directory like:

```
~/.local/share/yatto/tasks/
```

Each task is a simple JSON file.

## Git-Enabled Mode

If you enable Git support, yatto will:

- Automatically create a Git repo in the task directory
- Commit every add/delete/update
- Allow you to sync across devices (via Git remote)
- Make accidental deletions recoverable via history

> This is ideal for users who want **traceability, backups, or multi-device sync**.

## Roadmap

Planned features:

- 📅 Due dates
- ✅ Task completion status (done/undone)
- 🏷️ Tags or labels
- 📦 Export to Markdown or plain text

## 🧑‍💻 Built With

- [Go](https://go.dev)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — a fun, functional TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — for terminal styling

## 📃 License

MIT — see [LICENSE](LICENSE)

## 🤝 Contributing

Contributions, feedback, and ideas are welcome! Feel free to open issues or pull requests.
