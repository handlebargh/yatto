# YATTO

## Yet Another Terrific Task Organizer

[![Go Report Card](https://goreportcard.com/badge/github.com/handlebargh/yatto)](https://goreportcard.com/report/github.com/handlebargh/yatto)
![GitHub License](https://img.shields.io/github/license/handlebargh/yatto?color=blue)
![GitHub Release](https://img.shields.io/github/v/release/handlebargh/yatto?color=blue)

**YATTO** is a terminal-based to-do application built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). It stores each task as
a separate JSON file on your filesystem and manages the
task directory as a Git repository for versioning, synchronization and collaboration.

<img alt="YATTO demo" src="docs/demo.gif" />

## Features

- **TUI-based** interface powered by the Bubble Tea framework
- **Local file storage**: Each task is stored as an individual JSON file for easy inspection and portability
- **Git integration**: Initializes a Git repository in your task directory for:
  - Full version history of all tasks
  - Safe collaboration and backup
  - Sync across machines
- Every change is immediately committed and - if a remote is configured - pushed.
- **Tasks attributes** include due dates and searchable labels.
- Tasks can be **sorted** by different attributes.

## Roadmap

- **In progress state**: Mark tasks that are being worked on.
- **Sub-tasks**: Create tasks associated with a parent task.

## Requirements

- Git

## Installation

### Go

```bash
go install github.com/handlebargh/yatto@latest
```

### Binary

Take a look at the [releases](https://github.com/handlebargh/yatto/releases/latest).

## Configuration

A configuration file is automatically created at `${HOME}/.config/yatto/config.toml`

By default the following settings are written to the file and may be edited.

```toml
[git]
default_branch = 'main'

[git.remote]
enable = false
name = 'origin'

[storage]
path = '${HOME}/.yatto'
```

A config file may also be supplied by adding the `-config` flag:

```bash
yatto -config $PATH_TO_CONFIG_FILE
```

## Task Storage

Tasks are saved in a directory like:

```bash
${HOME}/.yatto
```

Each task is a simple JSON file, while projects are directories holding their associated tasks.

You can change the task storage directory in the config file.

### Git remotes

To set up a remote

1. Create a new repository on the Git host of your choice.

The repository must be empty, meaning that nothing must be committed at creation
(uncheck README, .gitignore and license files).

2. Enable remotes in the config

```toml
[git.remote]
enable = true
```

3. Run yatto at least once to create the task storage directory.

4. Add the remote and push the local repository

```bash
cd ${HOME}/.yatto
git remote add $GIT_REMOTE_URL
git push -u origin main
```

## License

MIT - see [LICENSE](LICENSE)

## Contributing

Contributions, feedback, and ideas are welcome! Feel free to open issues or pull requests.
