# Contributing to FastSchema

We would love for you to contribute to FastSchema and help make it even better than it is today!
As a contributor, here are the guidelines we would like you to follow:

- [Commit Message Guidelines](#commit)

## <a name="commit"></a> Commit Message Conventions

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification and [Angular 
Commit Message Format](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#-commit-message-format) 
for commit messages, but with our **scope**.

#### Type

- **feat**: A new feature
- **fix**: Fixes for bugs
- **docs**: Modifications that only affect documentation
- **refactor**: Code changes that neither fix a bug nor introduce a feature
- **test**: Adding or correcting tests
- **perf**: Changes that enhance performance
- **ci**: Updates to CI configuration files and scripts
- **chore**: Routine tasks or maintenance that don't change the application's behavior (e.g., updating build tools, cleaning up code v.v...)
- **revert**: Reverts a previous commit

#### Scope
*Updating*
- `common`
- `db`
- `log`
- `schema`
- `content`
- `auth`
- `user`
- `media/file`
- `permission`
- `git`
- `config`
- `dash`

For example:
- `feat: add login functionality`
- `feat(auth): add login functionality`
- `fix(auth): cannot login`
- `docs(readme): update setup instructions`
- `chore: add .gitignore file`

This approach helps maintain a clear and organized commit history. For more information, refer to the [Conventional Commits documentation](https://www.conventionalcommits.org) and the [Angular commit message format](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit).

