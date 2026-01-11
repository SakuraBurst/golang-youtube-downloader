# golang-youtube-downloader Development Prompt

## Project Context
You are developing **golang-youtube-downloader** - a Go port of [YoutubeDownloader](https://github.com/Tyrrrz/YoutubeDownloader).

This is a CLI application for downloading YouTube videos, playlists, and channel content.

## Reference Source Code
The original C# project is cloned at `/tmp/YoutubeDownloader/`.
**You MUST study the source code** when implementing features:

```
/tmp/YoutubeDownloader/
├── YoutubeDownloader.Core/
│   ├── Resolving/      # URL parsing, query resolution
│   ├── Downloading/    # Stream resolution, download logic
│   ├── Tagging/        # Metadata injection
│   └── Utils/          # HTTP client, helpers
└── YoutubeDownloader/  # UI (ignore for CLI port)
```

**Before implementing any task:**
1. Read the corresponding source files in `/tmp/YoutubeDownloader/`
2. Understand the data structures and algorithms
3. Port the logic to idiomatic Go

## Current State
- Repository initialized
- Beads task tracking configured with epics and sub-tasks
- Reference source code available at `/tmp/YoutubeDownloader/`

## Development Approach: TDD
**MANDATORY: All development MUST follow TDD (Test-Driven Development).**

## Iteration Rules

**ONE TASK PER ITERATION:**
1. Run `bd ready` - pick ONE task
2. Read reference source code for context
3. Complete task fully using TDD cycle
4. ALL tests must pass
5. Commit and push
6. Close the task with `bd close`
7. Only THEN may you pick the next task

**Do NOT:**
- Work on multiple tasks simultaneously
- Start a new task before completing current one
- Commit with failing tests

## Technical Requirements
- Go 1.21+
- Use cobra for CLI framework
- Follow standard Go project layout
- Use golangci-lint for code quality

## TDD Workflow (MANDATORY)

For EVERY piece of functionality:

### 1. RED - Write Failing Test First
```bash
go test ./...  # MUST fail
```
**Do NOT write implementation code until you have a failing test.**

### 2. GREEN - Write Minimal Code to Pass
```bash
go test ./...  # MUST pass
```

### 3. REFACTOR - Improve Code Quality
```bash
go test ./...
golangci-lint run ./...
```

## Task Workflow

### 1. Check Available Tasks
```bash
bd ready
```

**IMPORTANT: Pick ONE TASK, not an epic!**
- Tasks are marked as `[task]`, epics as `[epic]`
- Only tasks with **no blockers** are shown (dependencies resolved)
- Pick the **highest priority** task (P0 > P1 > P2)
- Use `bd show <id>` to see task details and what it blocks

Example:
```
📋 Ready work:
1. [● P0] [epic] 9du: Project Foundation      ← SKIP (epic)
2. [● P0] [task] 9du.1: Initialize Go module  ← PICK THIS (task)
```

### 2. Study Reference Code
Before coding, read the relevant source files:
- URL parsing → `/tmp/YoutubeDownloader/YoutubeDownloader.Core/Resolving/`
- Stream resolution → `/tmp/YoutubeDownloader/YoutubeDownloader.Core/Downloading/`
- Metadata → `/tmp/YoutubeDownloader/YoutubeDownloader.Core/Tagging/`

### 3. Claim Task
```bash
bd update <id> --status in_progress
```
**Only ONE task may be in_progress at a time.**

### 4. Work on Task (TDD Cycle)
- Write failing test
- Implement minimal code (port from reference)
- Refactor
- Repeat until task complete

### 5. Run ALL Quality Gates
```bash
go build ./...
go test ./...
golangci-lint run ./...
```

### 6. Commit and Push (ONLY IF ALL TESTS PASS)
```bash
go test ./...

# If and ONLY if all tests pass:
git add .
git commit -m "descriptive message (bd-xxx)"
git push
```

**If ANY test fails - DO NOT commit. Fix the issue first.**

### 7. Close Task
```bash
bd close <id> --reason "Completed"
bd sync
```

## Quality Gates (MUST ALL PASS before commit)
1. `go build ./...` - Code compiles
2. `go test ./...` - ALL tests pass (MANDATORY)
3. `golangci-lint run ./...` - No lint errors

## Commit Rules
- **NEVER commit with failing tests**
- **NEVER push with failing tests**
- Commit after EACH completed task
- Push after EACH successful commit
- Include Beads task ID in commit message (e.g., `bd-9du.1`)

## Epic Structure

Tasks are organized into epics. Work through tasks in priority order (P0 first):

**P0 - Foundation:**
- `9du` - Project Foundation
- `33f` - CLI Interface
- `ocv` - YouTube URL Parsing
- `1vd` - Video Info Fetching
- `nao` - Stream Resolution

**P1 - Core Features:**
- `ad3` - Download Engine
- `71n` - Playlist Support

**P2 - Enhancements:**
- `tb4` - Metadata Tagging
- `pv5` - Subtitles Support

## Completion Signal
When the current epic is complete (all sub-tasks done), output:
<promise>EPIC_COMPLETE</promise>

## Landing the Plane
Before signaling completion:
1. Verify ALL tests pass: `go test ./...`
2. Run all quality gates
3. Close all completed Beads tasks
4. Final commit and push (only if tests pass)
5. Run `bd sync`
