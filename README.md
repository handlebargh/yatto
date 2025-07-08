# 📝 yatto — Yet Another Terrific Task Organizer

**yatto** is a lightweight, terminal-based to-do application built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). It stores each task as
a separate JSON file on your filesystem and optionally manages the
task directory as a Git repository for versioning and collaboration.

---

## 🚀 Features

- ⚡ **TUI-based** interface powered by the Bubble Tea framework
- 🗂 **Local file storage**: Each task is stored as an individual JSON file for easy inspection and portability
- 🏷 **Basic task fields**: Title, description, and priority (High, Medium, Low)
- 🌱 **Git integration**: Optionally initializes a Git repository in your task directory for:
  - Full version history of all tasks
  - Safe collaboration and backup
  - Sync across machines with `git pull/push`

---

## 📦 Installation

```bash
go install github.com/handlebargh/yatto@latest
```

---

## 🖥️ Usage

Run from the terminal:

```bash
yatto
```

Inside the app:

- Use arrow keys or `j/k` to navigate
- Press `a` to create a new task
- Press `d` to delete a task (confirmation prompt included)
- Press `q` to quit

---

## 📁 Task Storage

Tasks are saved in a directory like:

```
~/.local/share/yatto/tasks/
```

Each task is a simple JSON file, making the data:

- Human-readable
- Easy to back up
- Suitable for automation/scripts

---

## 🔀 Git-Enabled Mode

If you enable Git support, yatto will:

- Automatically create a Git repo in the task directory
- Commit every add/delete/update
- Allow you to sync across devices (via Git remote)
- Make accidental deletions recoverable via history

> This is ideal for users who want **traceability, backups, or multi-device sync**.

---

## 🛣️ Roadmap

Planned features:

- 📅 Due dates
- ✅ Task completion status (done/undone)
- 🏷️ Tags or labels
- 📦 Export to Markdown or plain text

---

## 💡 Why JSON and Git?

- 🧱 JSON is lightweight and easy to manipulate
- 🧾 Git tracks every change without requiring a separate database
- 🔐 You're always in control of your data

---

## 🧑‍💻 Built With

- [Go](https://golang.org)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — a fun, functional TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — for terminal styling

---

## 📃 License

MIT — see [LICENSE](LICENSE)

---

## 🤝 Contributing

Contributions, feedback, and ideas are welcome! Feel free to open issues or pull requests.

---

## 🌟 Why yatto?

Because productivity tools shouldn’t be bloated. yatto gives you:

- A clean, focused terminal interface
- Local-first data ownership
- Git-enhanced history and safety

All wrapped up in a tiny Go binary.
