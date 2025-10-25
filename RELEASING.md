# Release Process

This project uses automated releases via GitHub Actions.

## Automatic Releases

Releases are automatically created when code is merged to the `master` branch.

### Version Bumping

The version is automatically incremented based on the commit message:

- **Patch version** (v1.0.0 → v1.0.1): Default for all commits
  ```
  fix: correct context handling
  chore: update dependencies
  docs: improve README
  ```

- **Minor version** (v1.0.0 → v1.1.0): Commits starting with `feat:`
  ```
  feat: add HTTP/3 support
  feat: implement new middleware
  ```

- **Major version** (v1.0.0 → v2.0.0): Breaking changes
  ```
  BREAKING: change Context API
  feat!: redesign middleware system
  fix!: breaking bug fix
  ```

### Skip Auto-Release

To merge to master without creating a release, include `[no release]` in your commit message:

```bash
git commit -m "chore: update CI config [no release]"
```

## Manual Releases

You can also create releases manually by pushing a tag:

```bash
# Create and push a tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

This will trigger the release workflow and create a GitHub release.

## Release Checklist

Before merging to master:

1. ✅ All tests pass
2. ✅ Code is reviewed
3. ✅ Documentation is updated
4. ✅ Examples work correctly
5. ✅ Commit message follows convention

## First Release

To create the first release (v0.1.0):

```bash
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

## Workflow Files

- `.github/workflows/ci.yml` - Runs tests and linting on PRs
- `.github/workflows/auto-release.yml` - Auto-creates releases on master merge
- `.github/workflows/release.yml` - Creates releases from manual tags

## Changelog

Changelogs are automatically generated from commit messages. Follow conventional commits:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Test updates
- `chore:` - Maintenance tasks

Example:
```
feat: add Server Push support for HTTP/2

Implements HTTP/2 Server Push API with c.Push() method.
Closes #123
```
