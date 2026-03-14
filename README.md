# cocotola-1.26

## Development Setup

### Conventional Commits

This project enforces [Conventional Commits](https://www.conventionalcommits.org/) format for commit messages.

#### Setup pre-commit hooks

1. Install pre-commit:

   ```bash
   pip install pre-commit
   ```

2. Install the git hooks:

   ```bash
   pre-commit install --hook-type commit-msg
   ```

3. (Optional) Run against all files to test:

   ```bash
   pre-commit run --all-files
   ```

#### Commit message format

Commit messages must follow this format:

```text
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Other changes that don't modify src or test files
- `revert`: Revert a previous commit

**Examples:**

```bash
git commit -m "feat: add user authentication"
git commit -m "fix: resolve login redirect issue"
git commit -m "docs: update README with setup instructions"
```
