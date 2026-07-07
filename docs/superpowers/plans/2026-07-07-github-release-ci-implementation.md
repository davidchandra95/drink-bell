# Drink Bell GitHub Release CI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Inline execution is recommended because this plan uses local GitHub authentication, repository creation, a first commit, push, and release verification.

**Goal:** Publish Drink Bell to `github.com/davidchandra95/drink-bell` and add unsigned, tag-triggered release CI artifacts for macOS, Windows, and Linux.

**Architecture:** Keep source release mechanics in the repository: GitHub Actions workflows own CI and release packaging, Fyne metadata owns app package metadata, and GitHub Releases own the published artifacts. Local execution initializes git, validates the source, commits only intended files, creates the public GitHub repository, pushes `main`, then pushes an annotated release tag.

**Tech Stack:** Go 1.26.3, Fyne v2.7.4, Fyne tools v1.7.2, GitHub Actions, GitHub CLI 2.93.0, Git 2.53.0.

## Global Constraints

- Target repository is `github.com/davidchandra95/drink-bell`.
- Repository visibility is public.
- Add MIT license.
- Release artifacts are unsigned for v1.
- Release artifacts are simple downloadable packages, not notarized/signed installers.
- Release workflow triggers only on tags matching `v*`.
- Initial release targets are exactly `darwin-arm64`, `darwin-amd64`, `windows-amd64`, and `linux-amd64`.
- Normal CI runs on pushes to `main` and pull requests.
- Local build output `/Users/slowtyper/code/drink-bell/drink-bell` must not be committed.
- Update Go module path from `drink-bell` to `github.com/davidchandra95/drink-bell` before the first public commit.
- Use `contents: write` only for the release publish job.
- Do not add long-lived personal access tokens, signing certificates, or release secrets.
- Commit/push/tag operations are in scope because David explicitly requested publishing this project to GitHub.

---

## File Structure

```text
/Users/slowtyper/code/drink-bell/
  .github/
    workflows/
      ci.yml
      release.yml
  assets/
    Icon.png
  tools/
    genicon/
      main.go
  .gitignore
  FyneApp.toml
  LICENSE
  README.md
  cmd/
    drink-bell/
      main.go
  docs/
    superpowers/
      plans/
        2026-07-07-drink-bell-implementation.md
        2026-07-07-github-release-ci-implementation.md
      specs/
        2026-07-07-github-release-ci-design.md
        2026-07-07-water-drink-reminder-design.md
  go.mod
  go.sum
  internal/
    reminder/
    ui/
```

Responsibilities:

```text
.gitignore
  -> excludes local build/release outputs and machine files

LICENSE
  -> MIT license for the public repo

FyneApp.toml
  -> app metadata consumed by fyne package

assets/Icon.png
  -> deterministic packaging icon

tools/genicon/main.go
  -> tiny reproducible icon generator using Go standard library only

.github/workflows/ci.yml
  -> normal push/PR tests and Linux build check

.github/workflows/release.yml
  -> tag-triggered unsigned desktop package release

README.md
  -> documents local run/test/build plus release behavior

go.mod and imports
  -> public GitHub module path
```

---

### Task 1: Repository Metadata And Module Path

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/.gitignore`
- Create: `/Users/slowtyper/code/drink-bell/LICENSE`
- Modify: `/Users/slowtyper/code/drink-bell/go.mod`
- Modify: `/Users/slowtyper/code/drink-bell/cmd/drink-bell/main.go`
- Modify: `/Users/slowtyper/code/drink-bell/internal/ui/app.go`
- Remove from working tree: `/Users/slowtyper/code/drink-bell/drink-bell`

**Interfaces:**
- Produces: module path `github.com/davidchandra95/drink-bell`
- Produces: import path `github.com/davidchandra95/drink-bell/internal/ui`
- Produces: import path `github.com/davidchandra95/drink-bell/internal/reminder`
- Produces: `.gitignore` exclusion for local binary and release artifacts
- Produces: MIT license text

- [ ] **Step 1: Remove local build artifact**

Run:

```bash
rm -f /Users/slowtyper/code/drink-bell/drink-bell
```

Expected:

```text
```

No output is expected.

- [ ] **Step 2: Add `.gitignore`**

Create `/Users/slowtyper/code/drink-bell/.gitignore`:

```gitignore
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

- [ ] **Step 3: Add MIT license**

Create `/Users/slowtyper/code/drink-bell/LICENSE`:

```text
MIT License

Copyright (c) 2026 David Chandra

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 4: Update Go module path**

Run:

```bash
go mod edit -module github.com/davidchandra95/drink-bell
```

Expected:

```text
```

No output is expected.

- [ ] **Step 5: Update command import**

Modify `/Users/slowtyper/code/drink-bell/cmd/drink-bell/main.go` to:

```go
package main

import (
	"github.com/davidchandra95/drink-bell/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	fyneApp := app.NewWithID("dev.slowtyper.drinkbell")
	ui.New(fyneApp).Run()
}
```

- [ ] **Step 6: Update UI import**

In `/Users/slowtyper/code/drink-bell/internal/ui/app.go`, replace:

```go
"drink-bell/internal/reminder"
```

with:

```go
"github.com/davidchandra95/drink-bell/internal/reminder"
```

- [ ] **Step 7: Tidy and format**

Run:

```bash
gofmt -w /Users/slowtyper/code/drink-bell/cmd /Users/slowtyper/code/drink-bell/internal
go mod tidy
```

Expected:

```text
```

No output is expected on success.

- [ ] **Step 8: Validate module path change**

Run:

```bash
sed -n '1,12p' go.mod
go test -count=1 ./...
go build ./cmd/drink-bell
```

Expected `go.mod` starts with:

```text
module github.com/davidchandra95/drink-bell
```

Expected tests:

```text
?   	github.com/davidchandra95/drink-bell/cmd/drink-bell	[no test files]
ok  	github.com/davidchandra95/drink-bell/internal/reminder
?   	github.com/davidchandra95/drink-bell/internal/ui	[no test files]
```

Expected build:

```text
```

No build output is expected on success. On macOS, this build may emit `ld: warning: ignoring duplicate libraries: '-lobjc'`; that warning is acceptable if the command exits 0.

---

### Task 2: Packaging Metadata And Deterministic Icon

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/FyneApp.toml`
- Create: `/Users/slowtyper/code/drink-bell/tools/genicon/main.go`
- Generate: `/Users/slowtyper/code/drink-bell/assets/Icon.png`

**Interfaces:**
- Produces: Fyne app ID `dev.slowtyper.drinkbell`
- Produces: Fyne app name `Drink Bell`
- Produces: Fyne app version `0.1.0`
- Produces: PNG icon at `assets/Icon.png`

- [ ] **Step 1: Add Fyne metadata**

Create `/Users/slowtyper/code/drink-bell/FyneApp.toml`:

```toml
Website = "https://github.com/davidchandra95/drink-bell"

[Details]
Icon = "assets/Icon.png"
Name = "Drink Bell"
ID = "dev.slowtyper.drinkbell"
Version = "0.1.0"
Build = 1

[LinuxAndBSD]
GenericName = "Water Reminder"
Categories = ["Utility"]
Comment = "A small water reminder tray app"
Keywords = ["water", "hydration", "reminder"]
```

- [ ] **Step 2: Add icon generator**

Create `/Users/slowtyper/code/drink-bell/tools/genicon/main.go`:

```go
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

const size = 512

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	drawCircle(img, size/2, size/2, 230, color.RGBA{R: 32, G: 129, B: 246, A: 255})
	drawCircle(img, size/2-80, size/2-90, 68, color.RGBA{R: 106, G: 173, B: 255, A: 210})

	drawCup(img)

	if err := os.MkdirAll("assets", 0o755); err != nil {
		log.Fatalf("create assets directory: %v", err)
	}

	path := filepath.Join("assets", "Icon.png")
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("create icon: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Fatalf("encode icon: %v", err)
	}
}

func drawCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || y < 0 || x >= size || y >= size {
				continue
			}
			dx := float64(x - cx)
			dy := float64(y - cy)
			if math.Sqrt(dx*dx+dy*dy) <= float64(radius) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func drawCup(img *image.RGBA) {
	outline := color.RGBA{R: 235, G: 246, B: 255, A: 255}
	glass := color.RGBA{R: 224, G: 242, B: 255, A: 235}
	water := color.RGBA{R: 78, G: 183, B: 255, A: 245}
	shadow := color.RGBA{R: 17, G: 80, B: 160, A: 80}

	for y := 156; y <= 388; y++ {
		progress := float64(y-156) / float64(388-156)
		left := int(172 + progress*34)
		right := int(340 - progress*34)
		for x := left; x <= right; x++ {
			switch {
			case x-left < 8 || right-x < 8 || y < 166 || y > 378:
				img.SetRGBA(x, y, outline)
			case y > 250:
				img.SetRGBA(x, y, water)
			default:
				img.SetRGBA(x, y, glass)
			}
		}
	}

	for y := 392; y <= 410; y++ {
		for x := 206; x <= 306; x++ {
			img.SetRGBA(x, y, shadow)
		}
	}

	for y := 136; y <= 158; y++ {
		for x := 164; x <= 348; x++ {
			img.SetRGBA(x, y, outline)
		}
	}
}
```

- [ ] **Step 3: Generate icon**

Run:

```bash
go run ./tools/genicon
```

Expected:

```text
```

No output is expected on success.

- [ ] **Step 4: Validate icon exists and compiles**

Run:

```bash
test -s assets/Icon.png
go test -count=1 ./...
```

Expected:

```text
?   	github.com/davidchandra95/drink-bell/cmd/drink-bell	[no test files]
ok  	github.com/davidchandra95/drink-bell/internal/reminder
?   	github.com/davidchandra95/drink-bell/internal/ui	[no test files]
?   	github.com/davidchandra95/drink-bell/tools/genicon	[no test files]
```

---

### Task 3: Normal CI Workflow

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/.github/workflows/ci.yml`

**Interfaces:**
- Produces: push-to-main and pull-request CI
- Produces: Ubuntu dependency setup for Fyne
- Produces: `go test -count=1 ./...`
- Produces: `go build ./cmd/drink-bell`

- [ ] **Step 1: Add CI workflow**

Create `/Users/slowtyper/code/drink-bell/.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches:
      - main
  pull_request:

permissions:
  contents: read

jobs:
  test:
    name: Test and build
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Install Fyne Linux dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev

      - name: Download modules
        run: go mod download

      - name: Run tests
        run: go test -count=1 ./...

      - name: Build app
        run: go build ./cmd/drink-bell
```

- [ ] **Step 2: Validate workflow file is present**

Run:

```bash
test -s .github/workflows/ci.yml
go test -count=1 ./...
```

Expected:

```text
?   	github.com/davidchandra95/drink-bell/cmd/drink-bell	[no test files]
ok  	github.com/davidchandra95/drink-bell/internal/reminder
?   	github.com/davidchandra95/drink-bell/internal/ui	[no test files]
?   	github.com/davidchandra95/drink-bell/tools/genicon	[no test files]
```

---

### Task 4: Tag-Triggered Release Workflow

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/.github/workflows/release.yml`

**Interfaces:**
- Consumes: `FyneApp.toml`
- Consumes: `assets/Icon.png`
- Produces: `DrinkBell-darwin-arm64.zip`
- Produces: `DrinkBell-darwin-amd64.zip`
- Produces: `DrinkBell-windows-amd64.zip`
- Produces: `DrinkBell-linux-amd64.tar.xz`
- Produces: GitHub Release on tag `v*`

- [ ] **Step 1: Add release workflow**

Create `/Users/slowtyper/code/drink-bell/.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: read

env:
  FYNE_TOOLS_VERSION: v1.7.2
  APP_ID: dev.slowtyper.drinkbell
  APP_NAME: Drink Bell

jobs:
  build-macos:
    name: Build macOS ${{ matrix.arch }}
    runs-on: macos-latest
    strategy:
      fail-fast: false
      matrix:
        arch:
          - arm64
          - amd64

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Package app
        shell: bash
        run: |
          set -euo pipefail
          version="${GITHUB_REF_NAME#v}"
          mkdir -p dist
          GOOS=darwin GOARCH="${{ matrix.arch }}" go build -o dist/DrinkBell ./cmd/drink-bell
          go run "fyne.io/tools/cmd/fyne@${FYNE_TOOLS_VERSION}" package \
            --executable dist/DrinkBell \
            --target darwin \
            --release \
            --icon assets/Icon.png \
            --name "${APP_NAME}" \
            --app-id "${APP_ID}" \
            --app-version "${version}"
          ditto -c -k --keepParent "${APP_NAME}.app" "dist/DrinkBell-darwin-${{ matrix.arch }}.zip"
          test -s "dist/DrinkBell-darwin-${{ matrix.arch }}.zip"

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: DrinkBell-darwin-${{ matrix.arch }}
          path: dist/DrinkBell-darwin-${{ matrix.arch }}.zip
          if-no-files-found: error

  build-windows:
    name: Build Windows amd64
    runs-on: windows-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Set up MinGW
        uses: msys2/setup-msys2@v2
        with:
          msystem: MINGW64
          update: true
          install: mingw-w64-x86_64-gcc

      - name: Add MinGW to PATH
        shell: pwsh
        run: |
          "C:\msys64\mingw64\bin" | Out-File -FilePath $env:GITHUB_PATH -Encoding utf8 -Append

      - name: Package app
        shell: pwsh
        run: |
          $ErrorActionPreference = "Stop"
          $version = $env:GITHUB_REF_NAME -replace '^v', ''
          $exe = "$($env:APP_NAME).exe"
          $env:GOOS = "windows"
          $env:GOARCH = "amd64"
          go build -o "$exe" ./cmd/drink-bell
          go run "fyne.io/tools/cmd/fyne@$env:FYNE_TOOLS_VERSION" package `
            --executable "$exe" `
            --target windows `
            --release `
            --icon assets/Icon.png `
            --name "$env:APP_NAME" `
            --app-id "$env:APP_ID" `
            --app-version "$version"
          New-Item -ItemType Directory -Force dist | Out-Null
          Compress-Archive -Path "$exe" -DestinationPath "dist/DrinkBell-windows-amd64.zip" -Force
          if (!(Test-Path "dist/DrinkBell-windows-amd64.zip")) {
            throw "missing windows artifact"
          }

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: DrinkBell-windows-amd64
          path: dist/DrinkBell-windows-amd64.zip
          if-no-files-found: error

  build-linux:
    name: Build Linux amd64
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Install Fyne Linux dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev

      - name: Package app
        shell: bash
        run: |
          set -euo pipefail
          version="${GITHUB_REF_NAME#v}"
          mkdir -p dist
          GOOS=linux GOARCH=amd64 go build -o dist/DrinkBell ./cmd/drink-bell
          go run "fyne.io/tools/cmd/fyne@${FYNE_TOOLS_VERSION}" package \
            --executable dist/DrinkBell \
            --target linux \
            --release \
            --icon assets/Icon.png \
            --name "${APP_NAME}" \
            --app-id "${APP_ID}" \
            --app-version "${version}"
          mv "${APP_NAME}.tar.xz" dist/DrinkBell-linux-amd64.tar.xz
          test -s dist/DrinkBell-linux-amd64.tar.xz

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: DrinkBell-linux-amd64
          path: dist/DrinkBell-linux-amd64.tar.xz
          if-no-files-found: error

  publish:
    name: Publish GitHub Release
    runs-on: ubuntu-latest
    needs:
      - build-macos
      - build-windows
      - build-linux
    permissions:
      contents: write

    steps:
      - name: Download artifacts
        uses: actions/download-artifact@v4
        with:
          path: release-assets
          merge-multiple: true

      - name: Verify artifacts
        shell: bash
        run: |
          set -euo pipefail
          ls -la release-assets
          test -s release-assets/DrinkBell-darwin-arm64.zip
          test -s release-assets/DrinkBell-darwin-amd64.zip
          test -s release-assets/DrinkBell-windows-amd64.zip
          test -s release-assets/DrinkBell-linux-amd64.tar.xz

      - name: Create release
        env:
          GH_TOKEN: ${{ github.token }}
        shell: bash
        run: |
          set -euo pipefail
          gh release create "${GITHUB_REF_NAME}" release-assets/* \
            --repo "${GITHUB_REPOSITORY}" \
            --title "Drink Bell ${GITHUB_REF_NAME}" \
            --notes "Unsigned desktop packages for Drink Bell ${GITHUB_REF_NAME}."
```

- [ ] **Step 2: Validate workflow file is present**

Run:

```bash
test -s .github/workflows/release.yml
go test -count=1 ./...
```

Expected:

```text
?   	github.com/davidchandra95/drink-bell/cmd/drink-bell	[no test files]
ok  	github.com/davidchandra95/drink-bell/internal/reminder
?   	github.com/davidchandra95/drink-bell/internal/ui	[no test files]
?   	github.com/davidchandra95/drink-bell/tools/genicon	[no test files]
```

---

### Task 5: README Release Documentation

**Files:**
- Modify: `/Users/slowtyper/code/drink-bell/README.md`

**Interfaces:**
- Consumes: `.github/workflows/ci.yml`
- Consumes: `.github/workflows/release.yml`
- Produces: release documentation for tag-triggered unsigned packages

- [ ] **Step 1: Add release section**

Modify `/Users/slowtyper/code/drink-bell/README.md` by inserting this section after the `Build` section:

````markdown
## Releases

Release artifacts are built by GitHub Actions when a version tag is pushed.

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The release workflow publishes unsigned desktop packages:

- `DrinkBell-darwin-arm64.zip`
- `DrinkBell-darwin-amd64.zip`
- `DrinkBell-windows-amd64.zip`
- `DrinkBell-linux-amd64.tar.xz`

These artifacts are unsigned for the first release setup. macOS may show Gatekeeper warnings and Windows may show SmartScreen warnings.
````

- [ ] **Step 2: Run docs-adjacent validation**

Run:

```bash
go test -count=1 ./...
```

Expected:

```text
?   	github.com/davidchandra95/drink-bell/cmd/drink-bell	[no test files]
ok  	github.com/davidchandra95/drink-bell/internal/reminder
?   	github.com/davidchandra95/drink-bell/internal/ui	[no test files]
?   	github.com/davidchandra95/drink-bell/tools/genicon	[no test files]
```

---

### Task 6: Local Validation, Git Initialization, Commit, And Push

**Files:**
- Create: `/Users/slowtyper/code/drink-bell/.git/`
- Publish: `github.com/davidchandra95/drink-bell`

**Interfaces:**
- Consumes: all files created or modified in Tasks 1-5
- Produces: local git branch `main`
- Produces: first commit `chore: publish initial drink bell app`
- Produces: GitHub remote `origin`
- Produces: public GitHub repo `https://github.com/davidchandra95/drink-bell`

- [ ] **Step 1: Final local validation before git init**

Run:

```bash
go test -count=1 ./...
go build ./cmd/drink-bell
rm -f drink-bell
```

Expected tests:

```text
?   	github.com/davidchandra95/drink-bell/cmd/drink-bell	[no test files]
ok  	github.com/davidchandra95/drink-bell/internal/reminder
?   	github.com/davidchandra95/drink-bell/internal/ui	[no test files]
?   	github.com/davidchandra95/drink-bell/tools/genicon	[no test files]
```

Expected build:

```text
```

No build output is expected on success. The macOS duplicate `-lobjc` linker warning is acceptable if the command exits 0.

- [ ] **Step 2: Confirm GitHub auth and target repo absence**

Run:

```bash
gh auth status
gh repo view davidchandra95/drink-bell --json name,visibility,url,defaultBranchRef
```

Expected auth includes:

```text
Logged in to github.com account davidchandra95
Git operations protocol: ssh
```

Expected repo view currently fails with:

```text
Could not resolve to a Repository with the name 'davidchandra95/drink-bell'
```

If the repo exists by the time this task runs, stop and inspect whether it is empty before pushing.

- [ ] **Step 3: Initialize git repository**

Run:

```bash
git init -b main
```

Expected:

```text
Initialized empty Git repository in /Users/slowtyper/code/drink-bell/.git/
```

- [ ] **Step 4: Inspect untracked scope before staging**

Run:

```bash
git status --short
find . -maxdepth 4 -type f | sort
```

Expected `git status --short` lists source/config/docs files as untracked and does not include `/drink-bell`.

- [ ] **Step 5: Stage only intended files**

Run:

```bash
git add \
  .gitignore \
  .github/workflows/ci.yml \
  .github/workflows/release.yml \
  FyneApp.toml \
  LICENSE \
  README.md \
  assets/Icon.png \
  cmd/drink-bell/main.go \
  docs/superpowers/plans/2026-07-07-drink-bell-implementation.md \
  docs/superpowers/plans/2026-07-07-github-release-ci-implementation.md \
  docs/superpowers/specs/2026-07-07-github-release-ci-design.md \
  docs/superpowers/specs/2026-07-07-water-drink-reminder-design.md \
  go.mod \
  go.sum \
  internal/reminder/controller.go \
  internal/reminder/controller_test.go \
  internal/reminder/options.go \
  internal/reminder/options_test.go \
  internal/reminder/state.go \
  internal/reminder/state_test.go \
  internal/ui/app.go \
  tools/genicon/main.go
```

Expected:

```text
```

No output is expected.

- [ ] **Step 6: Review staged diff**

Run:

```bash
git diff --cached --stat
git diff --cached --check
git status --short
```

Expected:

```text
```

`git diff --cached --check` should produce no output and exit 0. `git status --short` should show only staged additions and no local binary.

- [ ] **Step 7: Commit**

Run:

```bash
git commit -m "chore: publish initial drink bell app"
```

Expected:

```text
[main (root-commit) <sha>] chore: publish initial drink bell app
```

- [ ] **Step 8: Create public GitHub repository and remote**

Run:

```bash
gh repo create davidchandra95/drink-bell \
  --public \
  --description "A small cross-platform water reminder tray app built with Go and Fyne." \
  --source=. \
  --remote=origin
```

Expected:

```text
https://github.com/davidchandra95/drink-bell
```

- [ ] **Step 9: Push main**

Run:

```bash
git push -u origin main
```

Expected:

```text
branch 'main' set up to track 'origin/main'
```

- [ ] **Step 10: Verify remote repo**

Run:

```bash
gh repo view davidchandra95/drink-bell --json name,visibility,url,defaultBranchRef
git status --short --branch
```

Expected repo fields:

```json
{"name":"drink-bell","visibility":"PUBLIC","url":"https://github.com/davidchandra95/drink-bell",...}
```

Expected git status:

```text
## main...origin/main
```

---

### Task 7: Initial Tag Release And Remote Verification

**Files:**
- No source file changes.

**Interfaces:**
- Consumes: pushed `main`
- Produces: annotated git tag `v0.1.0`
- Produces: GitHub Actions release workflow run
- Produces: GitHub Release `v0.1.0`
- Produces: four unsigned artifacts attached to the release

- [ ] **Step 1: Confirm clean local state**

Run:

```bash
git status --short --branch
```

Expected:

```text
## main...origin/main
```

- [ ] **Step 2: Create annotated tag**

Run:

```bash
git tag -a v0.1.0 -m "v0.1.0"
```

Expected:

```text
```

No output is expected.

- [ ] **Step 3: Push tag**

Run:

```bash
git push origin v0.1.0
```

Expected:

```text
* [new tag]         v0.1.0 -> v0.1.0
```

- [ ] **Step 4: Find release workflow run**

Run:

```bash
gh run list --workflow release.yml --repo davidchandra95/drink-bell --limit 5
```

Expected:

```text
Release
```

The listed run should be associated with tag `v0.1.0`.

- [ ] **Step 5: Watch release workflow**

Run:

```bash
run_id="$(gh run list --workflow release.yml --repo davidchandra95/drink-bell --limit 1 --json databaseId --jq '.[0].databaseId')"
gh run watch "$run_id" --repo davidchandra95/drink-bell --exit-status
```

Expected:

```text
✓
```

If the workflow is still running and David does not want to wait, report the current run URL and stop polling. If any matrix job fails, inspect logs with `gh run view --log-failed` before changing code.

- [ ] **Step 6: Verify GitHub Release artifacts**

Run:

```bash
gh release view v0.1.0 --repo davidchandra95/drink-bell --json tagName,name,url,assets
```

Expected JSON includes:

```text
"tagName":"v0.1.0"
"DrinkBell-darwin-arm64.zip"
"DrinkBell-darwin-amd64.zip"
"DrinkBell-windows-amd64.zip"
"DrinkBell-linux-amd64.tar.xz"
```

- [ ] **Step 7: Final local audit**

Run:

```bash
git status --short --branch
git log --oneline --decorate -3
```

Expected status:

```text
## main...origin/main
```

Expected log includes:

```text
v0.1.0
chore: publish initial drink bell app
```

---

## Self-Review Checklist

- Public GitHub repo target: Task 6.
- MIT license: Task 1.
- Local binary excluded: Task 1 and Task 6.
- Module path changed to `github.com/davidchandra95/drink-bell`: Task 1.
- Packaging metadata and icon: Task 2.
- Normal CI on main/pull request: Task 3.
- Tag-only release workflow: Task 4.
- Unsigned macOS, Windows, and Linux artifacts: Task 4.
- Four target artifacts: Task 4 and Task 7.
- GitHub Actions `contents: write` only for publish job: Task 4.
- Initial commit and push: Task 6.
- Initial `v0.1.0` tag and release verification: Task 7.
- No signing secrets, personal access tokens, package manager publishing, notarization, or MSI/DEB/RPM/AppImage scope included.
