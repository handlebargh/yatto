# Contributing to This Project

Thanks for your interest in contributing!
Pull requests, bug reports, and new ideas are welcome.

## Commit Messages

Commit messages **must** follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

## Development Workflow

1. Fork the repository.

2. Clone your fork.

3. Create a new branch:

    ```shell
    git checkout -b feat/my-new-feature
    ```

4. Make your changes.

5. Lint and format before committing - all code must pass golangci-lint:

    ```shell
    just lint
    just fmt
    ```

    If it fails, fix all reported issues before committing.

6. Run the test suite before committing:

    ```shell
    just test
    ```

> [!TIP]
> If you don't want to install just, take a look a the justfile and run the commands manually.

> [!TIP]
> If you use cocogitto, run `cog install-hook --all` before commiting.

7. Commit your changes.

8. Push your branch and open a Pull Request.
