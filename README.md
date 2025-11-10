# PR Review CLI Tool

The PR review CLI tool fetches GitHub PR review comments using the GitHub GraphQL API, organizing them into threaded conversations with resolved/outdated status. By default, it shows only unresolved, non-outdated threads to focus on actionable feedback.

## Installation

### Install from source
```bash
go install github.com/AObuchow/pr-review-cli@latest
```

### Build locally
```bash
git clone https://github.com/AObuchow/pr-review-cli
cd pr-review-cli
go build -o pr-review-cli
```

The binary will be installed to your `$GOPATH/bin` directory (typically `~/go/bin`), which should be in your PATH.

## Authentication

You need a GitHub personal access token to use this tool. You can provide it in two ways:

1. **Environment variable** (recommended):
   ```bash
   export GITHUB_TOKEN=your_github_token_here
   ```

2. **Command line flag**:
   ```bash
   pr-review-cli fetch --token your_github_token_here [other options]
   ```

## Usage

### Basic Usage

Fetch PR review threads with Claude-optimized format (default):
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2
```

**Default Behavior:** Shows only unresolved, non-outdated review threads. This focuses on actionable feedback that needs to be addressed.

### Filtering Options

Control which threads are included in the output:

#### Include Resolved Threads
By default, resolved threads are hidden. Include them with:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --include-resolved
```

#### Include Outdated Threads
Threads on outdated code are hidden by default. Show them with:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --include-outdated
```

#### Include General PR Comments
Only inline code review threads are shown by default. Include general PR discussion comments with:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --include-general
```

#### Combined Filtering
Combine multiple flags to see all feedback:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --include-resolved --include-outdated --include-general
```

### Output Formats

The tool supports three output formats:

#### Claude Format (default)
Optimized for Claude AI analysis:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --format claude
```

#### Human-readable Format
Pretty-printed format for terminal viewing:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --format human
```

#### JSON Format
Raw JSON output for programmatic use:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --format json
```

### Complete Examples

Using environment variable for authentication:
```bash
export GITHUB_TOKEN=ghp_your_token_here

# Get unresolved threads only (default)
pr-review-cli fetch --owner microsoft --repo vscode --pr 123

# See all feedback including resolved and outdated
pr-review-cli fetch --owner facebook --repo react --pr 456 --include-resolved --include-outdated

# Human-friendly format with visual indicators
pr-review-cli fetch --owner facebook --repo react --pr 456 --format human

# Include general PR discussion comments
pr-review-cli fetch --owner microsoft --repo vscode --pr 123 --include-general
```

Using token flag:
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --token ghp_your_token_here --format json
```

## Help

Get general help:
```bash
pr-review-cli help
```

Get help for specific commands:
```bash
pr-review-cli fetch --help
```

## Output

The tool organizes PR feedback into threaded conversations with status indicators:

### Thread Structure
- **Review threads**: Grouped conversations on specific code lines
- **Thread status**: `[RESOLVED] ✓` or `[UNRESOLVED]` markers
- **Outdated indicator**: `⚠️ [OUTDATED]` for threads on changed code
- **Reply tracking**: Shows conversation flow with reply indicators
- **General comments**: Optional non-code PR discussion (use `--include-general`)

### Information Included
- Comment content and conversation threads
- File locations and line numbers in the diff
- Reviewer information
- Timestamps
- Enhanced summary statistics:
  - Unresolved thread count
  - Resolved thread count
  - Outdated thread count
  - Files affected
  - Authors

### Format Differences
The output format varies by option:
- **Claude format**: Optimized for AI analysis with clear action items
- **Human format**: Terminal-friendly with emojis and visual indicators (✅/❌/⚠️)
- **JSON format**: Complete structured data for programmatic use

## Claude Integration

### Using the Claude Command

This repository includes a custom Claude command that makes it easy to fetch PR comments directly from a GitHub PR URL. To use it in your own repository:

1. Copy the `.claude/commands/fetch-pr-comments.md` file to your repository's `.claude/commands/` directory
2. Make sure you have the `pr-review-cli` tool installed (see Installation section above)
3. Set your `GITHUB_TOKEN` environment variable

Then you can use the command in Claude Code:

```
/fetch-pr-comments url:https://github.com/AObuchow/Sample-Commander/pull/1
```

This will:
- Parse the PR URL to extract owner, repo, and PR number
- Fetch unresolved, non-outdated review threads (actionable feedback only)
- Format the output for Claude to analyze and help address the feedback

**Why Default Filtering?** By showing only unresolved threads, Claude focuses on feedback that actually needs your attention, making the workflow more efficient.

### Example Workflow

1. Get actionable PR feedback:
   ```
   /fetch-pr-comments url:https://github.com/your-org/your-repo/pull/123
   ```

2. Claude will fetch and display unresolved threads, then you can ask Claude to help:
   - "Help me address the performance concerns raised in the review"
   - "Fix the code style issues mentioned in the comments"
   - "Explain how to resolve the security vulnerability pointed out"

3. To see all feedback including resolved/outdated:
   ```bash
   pr-review-cli fetch --owner your-org --repo your-repo --pr 123 --include-resolved --include-outdated
   ```