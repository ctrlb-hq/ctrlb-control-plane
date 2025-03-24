# Contribution Guidelines
Welcome! 🚀 We're excited that you're interested in contributing to this project. Please follow these guidelines to ensure a smooth and consistent development experience.

## ✅ How to Contribute

### 1. Fork the repo and clone it

```bash
git clone https://github.com/your-username/project-name.git
```

### 2. Create a new branch

```bash
git checkout -b feat/add-agent-labels
```

### Branch Naming Convention

Please use the following format for branch names:

```
[type]/short-description
```

- `feat/` – New feature
- `fix/` – Bug fix
- `refactor/` – Code refactoring
- `sec/` – Security-related work
- `test/` – Testing improvements
- `docs/` – Documentation changes
- `chore/` – Maintenance or build-related changes

**Examples:**
- `feat/agent-label-endpoint`
- `fix/null-agent-id`
- `sec/add-auth-validation`
  

### 3. Follow Commit Message Guidelines

Use the following format:

```text
[TYPE]: Short description of change
```

#### Allowed Types:

| Tag         | Purpose                                      |
|-------------|----------------------------------------------|
| `[FEAT]`     | New feature or functionality                |
| `[FIX]`      | Bug fix                                     |
| `[REFACTOR]` | Code refactor (no behavior change)          |
| `[SEC]`      | Security-related changes                    |
| `[TEST]`     | Test-related updates                        |
| `[DOCS]`     | Documentation only                          |
| `[CHORE]`    | Build, CI, or dependency-related changes    |
| `[PERF]`     | Performance improvements                    |
| `[STYLE]`    | Code formatting, comments, no logic change  |
| `[WIP]`      | Work in progress (not ready to merge)       |

**Examples:**

```text
[FEAT]: Add endpoint to label agents
[FIX]: Prevent crash on nil agent ID
[SEC]: Add input sanitization to auth flow
```

### 4. Run tests and lint

Make sure everything passes before pushing.

```bash
go test ./...
golangci-lint run
```

### 5. Push and open a PR

```bash
git push origin feat/add-agent-labels
```

Then open a Pull Request with a clear explanation of the changes.

---

## 🥪 Testing Tips

- All core logic should have unit tests.
- Avoid tight coupling between handlers and services to enable easier mocking.

---

## 🙌 Thank You!

We appreciate all kinds of contributions — bug reports, feature suggestions, code, documentation, or even refactoring!

