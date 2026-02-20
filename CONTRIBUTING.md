# Contributing to This Project

Thanks for your interest in contributing!
Pull requests, bug reports, and new ideas are welcome.

## Development Workflow

1. Fork the repository.

2. Clone your fork.

3. Create a new branch:

    ```
    git checkout -b my-branch-name
    ```

4. Make your changes.

5. Format, lint and test before committing - all code must pass the checks:

    ```
    just fmt
    just lint
    just test
    ```

> [!IMPORTANT]
> If it fails, fix all reported issues before committing.

> [!TIP]
> If you don't want to install just, take a look a the justfile and run the commands manually.

6. Commit your changes.

7. Push your branch and open a Pull Request.
