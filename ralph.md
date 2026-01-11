# golang-youtube-downloader Development Prompt

## Project Context
You are developing **golang-youtube-downloader** - a Go port of [YoutubeDownloader](https://github.com/Tyrrrz/YoutubeDownloader).

This is a CLI application for downloading YouTube videos, playlists, and channel content.

## Current State
- Repository initialized with README.md
- Beads task tracking configured
- No code written yet - greenfield project

## Development Approach: CLI First + TDD
Build the command-line interface structure first, then implement YouTube functionality.

**MANDATORY: All development MUST follow TDD (Test-Driven Development).**

## Technical Requirements
- Go 1.21+
- Use cobra for CLI framework
- Follow standard Go project layout
- Use golangci-lint for code quality

## TDD Workflow (MANDATORY)

For EVERY piece of functionality, follow this cycle strictly:

### 1. RED - Write Failing Test First
```bash
# Write test that defines expected behavior
# Run test - it MUST fail
go test ./...
```
**Do NOT write implementation code until you have a failing test.**

### 2. GREEN - Write Minimal Code to Pass
```bash
# Write ONLY enough code to make the test pass
go test ./...
```
**The test MUST pass before proceeding.**

### 3. REFACTOR - Improve Code Quality
```bash
# Clean up code while keeping tests green
go test ./...
golangci-lint run ./...
```

**Repeat this cycle for every feature, function, and component.**

## Task Workflow

### 1. Check Available Tasks
```bash
bd ready
```

### 2. Claim Task
```bash
bd update <id> --status in_progress
```

### 3. Work on Task (TDD Cycle)
- Write failing test
- Implement minimal code
- Refactor
- Repeat until task complete

### 4. Before Closing Task - Run ALL Quality Gates
```bash
go build ./...
go test ./...
golangci-lint run ./...
```

### 5. Commit and Push (ONLY IF ALL TESTS PASS)
**CRITICAL: You may ONLY commit and push if ALL tests pass.**
```bash
# Verify ALL tests pass
go test ./...

# If and ONLY if all tests pass:
git add .
git commit -m "descriptive message (bd-xxx)"
git push
```

**If ANY test fails - DO NOT commit. Fix the issue first.**

### 6. Close Task
```bash
bd close <id> --reason "Completed"
bd sync
```

### 7. Create New Tasks as Discovered
```bash
bd create "Task title" -p 1
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
- Include Beads task ID in commit message

## Current Phase: CLI Foundation

### Phase 1 Completion Criteria
- [ ] Go module initialized (go.mod)
- [ ] cobra CLI framework integrated
- [ ] Basic command structure: `download`, `info`, `version`
- [ ] Help text for all commands
- [ ] Tests for CLI argument parsing (written FIRST per TDD)

### Phase 1 Deliverables
1. `cmd/` - CLI entry points
2. `internal/cli/` - Command implementations
3. `go.mod`, `go.sum` - Dependencies
4. Comprehensive test coverage (TDD)

## Completion Signal
When ALL Phase 1 criteria are met and ALL quality gates pass, output:
<promise>PHASE_COMPLETE</promise>

## Landing the Plane
Before signaling completion:
1. Verify ALL tests pass: `go test ./...`
2. Run all quality gates
3. Update/close all Beads tasks
4. Final commit and push (only if tests pass)
5. Run `bd sync`
