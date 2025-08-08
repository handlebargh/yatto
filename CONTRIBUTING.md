# Contributing to This Project

Thanks for your interest in contributing!
Pull requests, bug reports, and new ideas are welcome.

## Commit Messages

Commit messages **must** follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

I also recommend (and use) [Commitizen](https://commitizen-tools.github.io/commitizen/) to make this easy.

### Commit Message Format

- Start with the commit type (see below).
- Write a short and imperative summary of the code changes.
- Start in **lower case** and finish **without** period.

- If you can, add more detail in the commit body - this helps reviewers and future maintainers understand why a change was made, not just what was changed.
  Use the body to include:
  - The reasoning or motivation behind the change.
  - Any limitations, trade-offs, or side effects.
  - References to related issues and PRs.

---

**Example:**

fix: prevent crash when config file is missing

Previously, the application would panic during startup if the
configuration file was not found. This happened because the
file-reading function assumed the file always existed and did not
check for os.ErrNotExist.

This change adds a check for missing configuration files and falls
back to default settings, allowing the application to start normally
even when no config file is present.

Closes #73

---

**Allowed `<type>` values** include:

- `feat` – New feature
- `fix` – Bug fix
- `docs` – Documentation only changes
- `style` – Formatting, whitespaces, etc; no code changes
- `refactor` – Code change that neither fixes a bug nor adds a feature
- `perf` – Performance improvements
- `test` – Adding or correcting tests

#### About scopes

This project is monolithic, so you don’t usually need to use a scope in your commit messages.
However, if you feel a scope adds clarity, you may include one in parentheses right after the commit type:

## Development Workflow

1. Fork the repository.

2. Clone your fork.

3. Create a new branch:

```bash
git checkout -b feat/my-new-feature
```

4. Make your changes.

5. Lint before committing - all code must pass golangci-lint:

```bash
golangci-lint run ./...
```

If it fails, fix all reported issues before committing.

7. Commit your changes.

8. Push your branch and open a Pull Request.
