# yatto

**yatto** is a terminal-based to-do application built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). It stores each task as
a separate JSON file on your filesystem and manages the
task directory as a Git repository for versioning, synchronization and collaboration.

<img alt="yatto demo" src="docs/demo.gif" />

## Features

- **TUI-based** interface powered by the Bubble Tea framework
- **Local file storage**: Each task is stored as an individual JSON file for easy inspection and portability
- **Git integration**: Initializes a Git repository in your task directory for:
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

- Git

## Installation

### Go

```bash
go install github.com/handlebargh/yatto@latest
```

### [Homebrew](https://brew.sh/)

```bash
brew tap handlebargh/yatto
brew install yatto
```

### [Eget](https://github.com/zyedidia/eget)

```bash
eget handlebargh/yatto
```

### Binary

Take a look at the [releases](https://github.com/handlebargh/yatto/releases/latest).

## Configuration

A configuration file is automatically created at `${HOME}/.config/yatto/config.toml`

By default, the following settings are written to the file and may be edited.

```toml
[colors]
badge_text_dark = '#000000'
badge_text_light = '#000000'
blue_dark = '#1e90ff'
blue_light = '#1e90ff'
green_dark = '#02BF87'
green_light = '#02BA84'
indigo_dark = '#7571F9'
indigo_light = '#5A56E0'
orange_dark = '#FFA336'
orange_light = '#FFB733'
red_dark = '#FE5F86'
red_light = '#FE5F86'
vividred_dark = '#FE134D'
vividred_light = '#FE134D'
yellow_dark = '#CCCC00'
yellow_light = '#CCCC00'

[colors.form]
theme = 'Base16'

[vcs]
# allowed values: "git", "jj"
backend = "git"

[git]
default_branch = 'main'

[git.remote]
enable = false
name = 'origin'

[jj]
default_branch = "main"

[jj.remote]
enable = false
name = "origin"

[storage]
# replace with your actual home directory
# or any other path you'd like to use
path = '/home/foo/.yatto'
```

A config file may also be supplied by adding the `-config` flag:

```bash
yatto -config $PATH_TO_CONFIG_FILE
```

### Colors and themes

Don't like the default colors? Just change them.
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

Tasks are saved in a directory like:

```bash
${HOME}/.yatto
```

Each task is a simple JSON file, while projects are directories holding their associated tasks.

You can change the task storage directory in the config file.

### Git remotes

To set up a remote

1. Create a new repository on the Git host of your choice. The repository must be empty, meaning that nothing must be committed at creation (uncheck README, .gitignore and license files).

2. Run yatto at least once to create the task storage directory.

3. Add the remote and push the local repository.

    ```bash
    cd ${HOME}/.yatto
    git remote add origin $GIT_REMOTE_URL
    git push -u origin main
    ```

4. Enable `git.remote` in the config.

    ```toml
    [git.remote]
    enable = true
    ```

## Non-interactive mode

You can print a static list of your tasks to standard output:

```bash
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

```bash
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

Huge thanks to the [Charm](https://charm.land/) team for their incredible open-source libraries,
which power much of this project.
