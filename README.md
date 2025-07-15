# YATTO - Yet Another Terrific Task Organizer

**YATTO** is a terminal-based to-do application built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). It stores each task as
a separate JSON file on your filesystem and manages the
task directory as a Git repository for versioning, synchronization and collaboration.

## Features

- **TUI-based** interface powered by the Bubble Tea framework
- **Local file storage**: Each task is stored as an individual JSON file for easy inspection and portability
- **Git integration**: Initializes a Git repository in your task directory for:
  - Full version history of all tasks
  - Safe collaboration and backup
  - Sync across machines

## Requirements

Git is required.

## Installation

### Go

```bash
go install github.com/handlebargh/yatto@latest
```

### Binary

Take a look at the releases.

## Configuration

A configuration file is automatically created at `${HOME}/.config/yatto/config.toml`

By default the following settings are written to the file and may be edited.

```toml
[git]
default_branch = 'main'

[git.remote]
enable = false
name = 'origin'
push_on_commit = false

[storage]
path = '${HOME}/.yatto'
```

## Task Storage

Tasks are saved in a directory like:

```bash
${HOME}/.yatto
```

Each task is a simple JSON file.

You can change the task storage directory in the config file.

## Git-Enabled workflow

- Automatically create a Git repo in the task directory
- Commit every add/delete/update
- Allow you to sync across devices (via Git remote)
- Make accidental deletions recoverable via history

### Git remotes

If you set a Git remote URL in the config file **after**
you have already created tasks, you will have to navigate
to the task storage directory and push the repository manually
in order to keep the application running.

To do so, inside the storage directory run this command:

```bash
git push -u origin main
```

## Built With

- [Go](https://go.dev)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - a fun, functional TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - for terminal styling

## License

MIT - see [LICENSE](LICENSE)

## Contributing

Contributions, feedback, and ideas are welcome! Feel free to open issues or pull requests.
