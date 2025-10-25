# Contributing to Aqylly

Thank you for your interest in contributing to Aqylly!

## Development Workflow

1. **Fork the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/aqylly.git
   cd aqylly
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feat/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

3. **Make your changes**
   - Write code
   - Add tests
   - Update documentation

4. **Run tests locally**
   ```bash
   go test -v ./...
   ```

5. **Run linter**
   ```bash
   golangci-lint run
   ```

6. **Commit your changes**

   Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

   ```bash
   # For new features
   git commit -m "feat: add new middleware for authentication"

   # For bug fixes
   git commit -m "fix: correct context cancellation handling"

   # For documentation
   git commit -m "docs: update README with new examples"

   # For breaking changes
   git commit -m "feat!: redesign Context API"
   # or
   git commit -m "BREAKING: change middleware signature"
   ```

7. **Push and create Pull Request**
   ```bash
   git push origin feat/your-feature-name
   ```

## Commit Message Convention

Format: `<type>: <description>`

### Types:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks
- `perf:` - Performance improvements

### Breaking Changes:
Add `!` after type or start with `BREAKING:`
```bash
feat!: change Handler signature
BREAKING: remove deprecated methods
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` and `goimports`
- Keep functions small and focused
- Add comments for exported functions
- Write meaningful variable names

## Testing

- Write tests for new features
- Maintain or improve code coverage
- Test edge cases
- Use table-driven tests when appropriate

Example:
```go
func TestContext_Param(t *testing.T) {
    tests := []struct {
        name     string
        params   map[string]string
        key      string
        expected string
    }{
        {"existing param", map[string]string{"id": "123"}, "id", "123"},
        {"missing param", map[string]string{}, "id", ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := &Context{Params: tt.params}
            if got := c.Param(tt.key); got != tt.expected {
                t.Errorf("Param() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

## Pull Request Process

1. **Ensure CI passes**
   - All tests must pass
   - Linter must pass
   - Coverage should not decrease

2. **Update documentation**
   - Update README if needed
   - Add/update code comments
   - Update examples if relevant

3. **Request review**
   - Describe your changes
   - Link related issues
   - Add screenshots if UI changes

4. **Address review comments**
   - Make requested changes
   - Push updates to your branch

## Release Process

Releases are automated! See [RELEASING.md](RELEASING.md) for details.

- Merging to `master` automatically creates a release
- Version is determined by commit message
- Changelog is auto-generated

## Questions?

Feel free to:
- Open an issue for discussion
- Ask questions in pull requests
- Reach out to maintainers

## Code of Conduct

Be respectful and constructive. We're all here to build something great together!
