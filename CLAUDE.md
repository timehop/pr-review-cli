# Claude Development Guide

## Project Overview

This repository provides a CLI tool for fetching GitHub PR review comments and integrating them with Claude Code for automated feedback analysis and resolution.

## Scope

The scope of this repository is focused and intentionally limited to:
- **Primary Purpose**: Give Claude a tool to locally fetch PR review comments
- **Secondary Purpose**: Enable Claude to address PR feedback efficiently
- **Integration**: Seamless Claude Code integration via custom slash commands

## Key Components

### 1. CLI Tool (`pr-review-cli`)
- **Purpose**: Fetch GitHub PR review comments via GitHub API
- **Output Formats**: 
  - `claude`: Optimized for Claude analysis (default)
  - `human`: Terminal-friendly format
  - `json`: Raw structured data
- **Authentication**: GitHub personal access token (env var or flag)

### 2. Claude Command Integration
- **File**: `.claude/commands/fetch-pr-comments.md`
- **Usage**: `/fetch-pr-comments url:https://github.com/owner/repo/pull/123`
- **Workflow**: Parse URL → Extract details → Fetch comments → Format for analysis

### 3. Core Files
- `main.go`: CLI entry point and command handling
- `github.go`: GitHub API client and PR comment fetching
- `models.go`: Data structures for comments and responses
- `formatter.go`: Output formatting for different formats
- `go.mod`: Go module definition

## Usage Patterns

### For Claude Development
1. Use `/fetch-pr-comments url:...` to get PR feedback
2. Analyze comments for patterns and required changes
3. Address feedback systematically
4. Re-run after changes to verify resolution

### For Manual Development
```bash
# Development workflow
pr-review-cli fetch --owner owner --repo repo --pr 123 --format claude
# Review output and address feedback
# Iterate as needed
```

## Installation Requirements
- Go 1.23.4+
- GitHub personal access token
- Access to target repositories for PR review fetching

## Future Considerations
- Keep dependencies minimal
- Maintain backwards compatibility
- Focus on Claude integration improvements
- Optimize for common PR review workflows