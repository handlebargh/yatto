# yatto

**yatto** is a terminal-based to-do application built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). It stores each task as
a separate JSON file on your filesystem and manages the
task directory as a Git or Jujutsu repository for versioning, synchronization and collaboration.

<img alt="yatto demo" src="docs/demo.gif" />

## Features

- **TUI-based** interface powered by the Bubble Tea framework
- **Local file storage**: Each task is stored as an individual JSON file for easy inspection and portability
- **VCS integration**: Initializes a Git or Jujutsu repository in your task directory for:
  - Full version history of all tasks
  - Safe collaboration and backup
  - Sync across machines
- Every change is immediately committed and - if a remote is configured - pushed.
- Tasks are organized in **projects**
- **Task attributes** include due dates and searchable labels.
- Tasks can be:
  - written in **markdown**
  - **sorted** by due date or priority
  - **marked** as in progress
- **Non-interactive output**: Print all open tasks from any project
- Supports **simple theme customization**

## Roadmap

- **Sub-tasks**: Create tasks associated with a parent task.

## Requirements

You need to have at least one of the supported version control systems installed:

- Git
- Jujutsu (jj)

## Installation

### Go

```shell
go install github.com/handlebargh/yatto@latest
```

### [Homebrew](https://brew.sh/)

```shell
brew tap handlebargh/yatto
brew install yatto
```

### [Eget](https://github.com/zyedidia/eget)

```shell
eget handlebargh/yatto
```

### Binary

Take a look at the [releases](https://github.com/handlebargh/yatto/releases/latest).

### Verifying Release Binaries

All release binaries are published with a [SLSA provenance attestation](https://slsa.dev/) (*.intoto.jsonl files).
This allows you to cryptographically verify that the binaries were built from this source code,
using GitHub Actions, and not tampered with.

#### How to verify

1. Install the [verifier](https://github.com/slsa-framework/slsa-verifier).
2. Download the yatto binary and provenance file.
3. Verify the binary, e.g.

    ```shell
    slsa-verifier verify-artifact \
      --provenance-path ./yatto-linux-amd64.intoto.jsonl \
      --source-uri github.com/handlebargh/yatto \
      --source-versioned-tag v0.17.0 \
      yatto-linux-amd64
    ```

## Configuration

When you start the application for the first time,
it will ask you to set up a configuration file located at: `${HOME}/.config/yatto/config.toml`

See [examples/config.toml](examples/config.toml) as a reference with all available configuration values.

> [!TIP]
> Alternatively, a config file may also be supplied by adding the `-config` flag:
>
> ```bash
> yatto -config $PATH_TO_CONFIG_FILE
> ```

### Colors and themes

User interface colors are customizable.
Any color value supported by [lipgloss](https://github.com/charmbracelet/lipgloss?tab=readme-ov-file#colors) will be accepted.

Every color accepts a light and a dark value for either light or dark terminal themes.

If you feel like sharing your theme, just post it in an issue,
and I'll be happy to add it to the repository.

You can also choose from one of the [predefined form themes](https://github.com/charmbracelet/huh?tab=readme-ov-file#themes). The following config values are supported:

- Charm
- Dracula
- Catppuccin
- Base16
- Base

Set a theme like this:

```toml
[colors.form]
theme = 'Catppuccin'

```

## Task Storage

At first startup, the application will also ask whether to create a task storage directory.
By default, tasks are stored in:

```bash
${HOME}/.yatto
```

Each task is represented as a simple JSON file, and projects are stored as directories
containing their related tasks.

The task storage location can be customized in the config file.

### VCS remotes

To set up a remote

1. Create a new repository on the Git host of your choice.
   The repository must be empty, meaning that nothing must be committed at creation
   (uncheck README, .gitignore and license files).

2. Enable remote in the config.

   #### Git
    ```toml
    [git.remote]
    enable = true
    url = <GIT_REMOTE_URL>
    ```

   #### Jujutsu
     ```toml
    [jj.remote]
    enable = true
    url = <GIT_REMOTE_URL>
    ```

## Non-interactive mode

You can print a static list of your tasks to standard output:

```shell
yatto -print

# Limit to any project you want
# Get the IDs from the directory names in your storage directory
# Run this command to print all project's metadata files:
# find ${HOME}/.yatto -type f -name "project.json" -exec cat {} +
yatto -print -projects "2023255a-1749-4f6c-9877-0c73ab42e5ab b5811d17-dbc7-4556-886b-92047a27e0f6"

# Filter labels with regular expression
# The next command will only show tasks that have a label "frontend"
yatto -print -regex frontend
```

If you want to print this list whenever you run an interactive shell,
open your `~/.bashrc` (or `~/.zshrc`) and add the following snippet:

```shell
# Print yatto task list only in interactive shells
case $- in
    *i*)
        if command -v yatto >/dev/null 2>&1; then
            yatto -print
        fi
        ;;
esac
```

> [!TIP]
> Add the -pull flag to pull from a configured remote before printing.

## License

MIT - see [LICENSE](LICENSE)

## Contributing

Contributions, feedback, and ideas are welcome! See [how to contribute](CONTRIBUTING.md) to this repository.

## Acknowledgements

Huge thanks to the [Charm](https://github.com/charmbracelet) team and their contributors
for their incredible open-source libraries, which power much of this project.
