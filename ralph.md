# golang-youtube-downloader Development Prompt

## Project Context
You are developing **golang-youtube-downloader** - a Go port of [YoutubeDownloader](https://github.com/Tyrrrz/YoutubeDownloader).

This is a CLI application for downloading YouTube videos, playlists, and channel content.

## Current State
- Repository initialized with README.md
- Beads task tracking configured
- No code written yet - greenfield project

## Development Approach: CLI First
Build the command-line interface structure first, then implement YouTube functionality.

## Technical Requirements
- Go 1.21+
- Use cobra for CLI framework
- Follow standard Go project layout
- Write tests for all packages
- Use golangci-lint for code quality

## Task Workflow

### 1. Check Available Tasks
```bash
bd ready
```

### 2. Claim and Work on Tasks
```bash
bd update <id> --status in_progress
```

### 3. After Completing Work
```bash
bd close <id> --reason "Completed"
```

### 4. Create New Tasks as Discovered
```bash
bd create "Task title" -p 1
```

## Quality Gates (Run Before Completion)
1. `go build ./...` - Code compiles
2. `go test ./...` - Tests pass
3. `golangci-lint run ./...` - No lint errors

## Current Phase: CLI Foundation

### Phase 1 Completion Criteria
- [ ] Go module initialized (go.mod)
- [ ] cobra CLI framework integrated
- [ ] Basic command structure: `download`, `info`, `version`
- [ ] Help text for all commands
- [ ] Tests for CLI argument parsing

### Phase 1 Deliverables
1. `cmd/` - CLI entry points
2. `internal/cli/` - Command implementations
3. `go.mod`, `go.sum` - Dependencies
4. Basic test coverage

## Completion Signal
When ALL Phase 1 criteria are met and quality gates pass, output:
<promise>PHASE_COMPLETE</promise>

## Landing the Plane
Before signaling completion:
1. Run all quality gates
2. Update/close Beads tasks
3. Commit all changes
4. Run `bd sync`
5. Push to remote
