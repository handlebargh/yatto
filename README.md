# yatto

**A Git-backed terminal task manager with full history, sync, and collaboration built in.**

[![Go Report Card](https://goreportcard.com/badge/github.com/handlebargh/yatto)](https://goreportcard.com/report/github.com/handlebargh/yatto)
[![Go Version](https://img.shields.io/github/go-mod/go-version/handlebargh/yatto)](https://github.com/handlebargh/yatto/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/handlebargh/yatto)](https://github.com/handlebargh/yatto/releases)
[![CI](https://github.com/handlebargh/yatto/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/handlebargh/yatto/actions)
[![codecov](https://codecov.io/gh/handlebargh/yatto/branch/main/graph/badge.svg?threshold=80&style=flat-square)](https://codecov.io/gh/handlebargh/yatto)

**yatto** is a terminal-first to-do application that stores each task as a plain JSON file
and manages your task directory as a Git or Jujutsu repository.

Every change is versioned. Collaboration and sync come with your version control system.
Your tasks remain transparent, portable, and fully under your control.

<img alt="yatto demo" src="docs/demo.gif" />

## Why yatto?

- **Versioned by default** - every change is committed automatically
- **Uses standard VCS remotes** - GitHub, GitLab, self-hosted, etc.
- **Plain files, no lock-in** - inspect, edit, or script tasks with standard tools
- **Terminal-native** - fast, keyboard-driven navigation
- **Offline-first** - works anywhere your VCS does

## Features

- TUI interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Local file storage (one JSON file per task)
- Git or Jujutsu integration:
    - full task history
    - safe backup
    - multi-machine sync
- Automatic commit on every change (optional auto-push)
- Project-based task organization
- Task attributes:
    - due dates
    - labels
    - priority
    - author / assignee
- Markdown support for task descriptions
- Non-interactive output (`yatto print`) for simple dashboards
- Simple theme and color customization

## Requirements

You need to have at least one of the supported version control systems installed:

- Git
- Jujutsu (jj)

## Installation

<details>
  <summary>Go</summary>

To install, run the following [Go](https://go.dev/) command.

```shell
go install github.com/handlebargh/yatto@latest
```

</details>

<details>
  <summary>AUR</summary>

If using [yay](https://github.com/Jguer/yay), run the following command.

```shell
yay -S yatto
```

Or if using [paru](https://github.com/Morganamilo/paru), run this command:

```shell
paru -S yatto
```

</details>

<details>
  <summary>Homebrew</summary>

To install, run the following [brew](https://brew.sh/) commands.

```shell
brew tap handlebargh/yatto
brew install yatto
```

</details>

<details>
  <summary>Nix</summary>

To install, run the following [Nix](https://nixos.org/download/) commands.

A nix-shell can be used to try yatto, before installing it permanently.

```shell
nix-shell -p yatto
```

On NixOS, add the following Nix code to your NixOS configuration.

```
environment.systemPackages = [
    pkgs.yatto
];
```

</details>

<details>
  <summary>Eget</summary>

To install, run the following [eget](https://github.com/zyedidia/eget) commands.

```shell
eget handlebargh/yatto
```

</details>

<details>
  <summary>Scoop</summary>

To install, run the following [scoop](https://scoop.sh/) commands.

```powershell
scoop bucket add scoop-handlebargh https://github.com/handlebargh/scoop-handlebargh
scoop install scoop-handlebargh/yatto
```

</details>

<details>
  <summary>Binaries and Linux packages</summary>

Take a look at the [releases](https://github.com/handlebargh/yatto/releases/latest) for prebuilt binaries and packages.

**Verifying release binaries**

1. Install [cosign](https://github.com/sigstore/cosign)
2. Verify the signed checksum file

    ```shell
    cosign verify-blob \
      --bundle /path/to/checksums.txt.sig.bundle \
      --certificate-identity-regexp "https://github.com/handlebargh/yatto/.*" \
      --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
      /path/to/checksums.txt
    ```

3. Check that the binaries match the signed checksums

    ```shell
    sha256sum --check --ignore-missing path/to/checksums.txt
    ```

**Verifying release packages**

1. Download and import the public key:

    ```shell
    sudo rpm --import yatto_signing_pubkey.gpg      # For RPM
    gpg --import yatto_signing_pubkey.gpg           # For DEB
    ```

2. Verify the package:

    ```shell
    rpm --checksig /path/to/yatto.rpm       # RPM
    dpkg-sig --verify /path/to/yatto.deb    # DEB
    ```

3. Install only after the signature is valid:

    ```shell
    sudo rpm -i /path/to/yatto.rpm       # RPM

    sudo dpkg -i /path/to/yatto.deb      # DEB
    or
    sudo apt install /path/to/yatto.deb  # DEB
    ```

</details>

## Configuration

When you start the application for the first time,
it will ask you to set up a configuration file located at: `${HOME}/.config/yatto/config.toml`

See [examples/config.toml](examples/config.toml) as a reference with all available configuration values.

> [!TIP]
> Alternatively, a config file may also be supplied by adding the `--config` flag:
>
> ```bash
> yatto --config $PATH_TO_CONFIG_FILE
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

#### From scratch

2. Run yatto and enter your SSH URL in the config dialog.

#### With existing storage directory

2. Push the current state manually to the remote.

3. Enable remote in the config.

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
yatto print

# Limit to any project you want
# Get the IDs from the directory names in your storage directory
# Run this command to print all project's metadata files:
# find ${HOME}/.yatto -type f -name "project.json" -exec cat {} +
yatto print --projects "2023255a-1749-4f6c-9877-0c73ab42e5ab b5811d17-dbc7-4556-886b-92047a27e0f6"

# Filter labels with regular expression
# The next command will only show tasks that have a label "frontend"
yatto print --regex frontend
```

If you want to print this list whenever you run an interactive shell,
open your `~/.bashrc` (or `~/.zshrc`) and add the following snippet:

```shell
# Print yatto task list only in interactive shells
case $- in
    *i*)
        if command -v yatto >/dev/null 2>&1; then
            yatto print
        fi
        ;;
esac
```

> [!TIP]
> Add the --pull flag to pull from a configured remote before printing.

## License

MIT - see [LICENSE](LICENSE)

## Contributing

Contributions, feedback, and ideas are welcome! See [how to contribute](CONTRIBUTING.md) to this repository.

## Acknowledgements

Huge thanks to the [Charm](https://github.com/charmbracelet) team and their contributors
for their incredible open-source libraries, which power much of this project.
