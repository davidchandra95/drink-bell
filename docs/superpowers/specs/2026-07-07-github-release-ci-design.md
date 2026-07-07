# Drink Bell GitHub Release CI Design

Date: 2026-07-07

## Summary

Publish Drink Bell to a new public personal GitHub repository and add CI/release automation for unsigned desktop packages.

Target repository:

```text
github.com/davidchandra95/drink-bell
```

The first release system should be intentionally conservative:

- Public GitHub repository under `davidchandra95`.
- MIT license.
- Normal push/PR CI for tests and a Linux build check.
- Tag-triggered release workflow for desktop artifacts.
- Unsigned release packages first.
- Simple downloadable packages, not notarized/signed installers.
- Four initial targets:
  - `darwin-arm64`
  - `darwin-amd64`
  - `windows-amd64`
  - `linux-amd64`

## Current Project State

The local project exists at:

```text
/Users/slowtyper/code/drink-bell
```

Current state:

- The app source exists and builds locally.
- `go.mod` uses module path `drink-bell`.
- Fyne dependency is present.
- The folder is not currently a git repository.
- A local build output exists at `/Users/slowtyper/code/drink-bell/drink-bell`.

The local build output must not be committed. Release artifacts should be produced by CI from source.

Before the first public commit, the Go module path should be updated to:

```text
github.com/davidchandra95/drink-bell
```

The internal import in `cmd/drink-bell/main.go` should be updated at the same time. Keeping the local-only module path would work for a private local app, but it is a weak default for a public GitHub Go repository.

## Goals

- Initialize the local project as a git repository.
- Add GitHub-ready project metadata:
  - `.gitignore`
  - `LICENSE`
  - release packaging metadata
  - GitHub Actions workflows
- Update the Go module path to `github.com/davidchandra95/drink-bell`.
- Create a public GitHub repository named `drink-bell` under `davidchandra95`.
- Push the initial source to `main`.
- Provide a tag-triggered release pipeline.
- Produce unsigned downloadable packages for macOS, Windows, and Linux.
- Avoid signing/notarization complexity in v1.
- Keep release artifacts reproducible from CI.

## Non-Goals

- No macOS notarization in v1.
- No Apple Developer ID signing in v1.
- No Windows Authenticode signing in v1.
- No `.msi`, `.deb`, `.rpm`, AppImage, Homebrew formula, winget manifest, or package-manager publishing in v1.
- No automatic semantic version bumping in v1.
- No release creation from ordinary branch pushes.

## Repository Setup

The repository setup should be explicit and auditable:

```text
drink-bell/
  -> git init
  -> add .gitignore
  -> add MIT LICENSE
  -> update go module path and imports
  -> add packaging metadata
  -> add CI/release workflows
  -> run validation
  -> stage only intended source/config/docs files
  -> commit
  -> create public GitHub repo davidchandra95/drink-bell
  -> add origin
  -> push main
```

The first commit should include source, tests, docs, workflow files, license, icon, and packaging metadata. It should exclude local build outputs and machine-specific files.

Suggested commit message:

```text
chore: publish initial drink bell app
```

## Git Ignore Policy

Add a `.gitignore` that excludes:

```text
/drink-bell
/dist/
/release/
/*.app
/*.exe
/*.tar.gz
/*.tar.xz
/*.zip
.DS_Store
```

The goal is to keep CI artifacts and local build outputs out of source control while preserving application source, docs, workflow files, and static packaging metadata.

## License

Add MIT license for a small public personal utility.

The copyright holder should be:

```text
David Chandra
```

## Packaging Metadata

Add:

```text
FyneApp.toml
assets/Icon.png
```

`FyneApp.toml` should carry the app name, app ID, icon path, and basic metadata used by the Fyne packaging command. Fyne's packaging docs state that `fyne package` can generate platform packages using the target OS and metadata such as an icon, and that a `.png` icon can be converted for platform formats.

The icon can be simple for v1. It only needs to be a valid PNG that makes packaging deterministic. It can be replaced later with a designed app icon.

Proposed metadata:

```toml
Website = "https://github.com/davidchandra95/drink-bell"

[Details]
  Icon = "assets/Icon.png"
  Name = "Drink Bell"
  ID = "dev.slowtyper.drinkbell"
  Version = "0.1.0"
  Build = 1
```

## CI Workflow

Add:

```text
.github/workflows/ci.yml
```

Trigger:

```yaml
on:
  push:
    branches: [main]
  pull_request:
```

CI job:

```text
ubuntu-latest
  -> checkout
  -> setup Go
  -> install Linux Fyne build dependencies
  -> go mod download
  -> go test -count=1 ./...
  -> go build ./cmd/drink-bell
```

Rationale:

- Ubuntu CI is enough for normal push/PR feedback.
- It catches Go test failures and Linux Fyne dependency regressions.
- Cross-platform package checks belong in the release workflow because they need OS-specific runners.

## Release Workflow

Add:

```text
.github/workflows/release.yml
```

Trigger only on tags:

```yaml
on:
  push:
    tags:
      - "v*"
```

The release workflow should have two stages:

```text
build matrix
  -> build and package per OS/arch
  -> upload one artifact per target

publish
  -> wait for every build matrix target
  -> download artifacts
  -> create GitHub Release for the pushed tag
  -> attach artifacts
```

Use `permissions: contents: write` only for the publish job. GitHub's workflow permissions documentation states that `contents: write` allows release creation, so the build jobs should not need broad write permission.

### Release Targets

```text
darwin-arm64   -> DrinkBell-darwin-arm64.zip
darwin-amd64   -> DrinkBell-darwin-amd64.zip
windows-amd64  -> DrinkBell-windows-amd64.zip
linux-amd64    -> DrinkBell-linux-amd64.tar.xz
```

### macOS Packaging

Runner:

```text
macos-latest
```

Targets:

```text
GOOS=darwin GOARCH=arm64
GOOS=darwin GOARCH=amd64
```

Package shape:

```text
Drink Bell.app
DrinkBell-darwin-arm64.zip
DrinkBell-darwin-amd64.zip
```

Expected limitation:

- The `.app` bundle is unsigned and not notarized.
- Users may need to bypass Gatekeeper warnings.

### Windows Packaging

Runner:

```text
windows-latest
```

Target:

```text
GOOS=windows GOARCH=amd64
```

Package shape:

```text
DrinkBell.exe
DrinkBell-windows-amd64.zip
```

Expected limitation:

- The `.exe` is unsigned.
- Users may see SmartScreen warnings.

### Linux Packaging

Runner:

```text
ubuntu-latest
```

Target:

```text
GOOS=linux GOARCH=amd64
```

Package shape:

```text
DrinkBell-linux-amd64.tar.xz
```

The Linux runner must install Fyne native build dependencies before building/package steps. At minimum this should include common X11/OpenGL development packages used by Fyne builds.

Expected limitation:

- Linux packaging is a generic tarball for v1.
- Desktop integration may still vary by distro and desktop environment.

## Release Data Flow

```text
Developer pushes main
  |
  +-- ci.yml
      -> test/build only

Developer creates annotated tag v0.1.0
  |
  +-- push tag
      |
      +-- release.yml
          |
          +-- build darwin-arm64
          +-- build darwin-amd64
          +-- build windows-amd64
          +-- build linux-amd64
              |
              +-- upload artifacts
                  |
                  +-- publish GitHub Release
                      |
                      +-- attach all 4 artifacts
```

## Failure Handling

The release should fail closed.

- If `ci.yml` fails on `main`, do not tag a release.
- If any release matrix target fails, the publish job must not create a partial release.
- If artifact upload fails, the publish job must not run.
- If GitHub release creation fails, the tag remains but the release is absent; rerun or fix the workflow rather than uploading local artifacts manually.
- If local GitHub CLI auth is broken, stop before repo creation/push unless another authenticated path is available.

## Verification Plan

Before first commit/push:

```bash
go test -count=1 ./...
go build ./cmd/drink-bell
git status --short
git add <intended files only>
git diff --cached --stat
```

Before first release tag:

```bash
git status --short
gh auth status
gh repo view davidchandra95/drink-bell
```

First release tag:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

After tag push:

```bash
gh run list --workflow release.yml --limit 5
gh run watch <run-id>
gh release view v0.1.0 --repo davidchandra95/drink-bell
```

The release is considered complete only after the GitHub Release exists and contains all four expected artifacts.

## Security And Permissions

- Use the default `GITHUB_TOKEN` in GitHub Actions.
- Keep `contents: read` as the default workflow permission.
- Grant `contents: write` only in the release publish job.
- Do not add long-lived personal access tokens to the repository for v1.
- Do not store signing keys or certificates because v1 is unsigned.

## Sources Checked

- Fyne packaging documentation: `https://docs.fyne.io/started/packaging/`
- Fyne app metadata documentation: `https://docs.fyne.io/started/metadata/`
- GitHub Actions workflow syntax and permissions: `https://docs.github.com/actions/using-workflows/workflow-syntax-for-github-actions`
- GitHub Actions artifact documentation: `https://docs.github.com/en/actions/tutorials/store-and-share-data`
- GitHub CLI release creation manual: `https://cli.github.com/manual/gh_release_create`
