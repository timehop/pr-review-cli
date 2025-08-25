# PR Review CLI Tool

The PR review CLI tool parses GitHub PR review comments and their location in the diff, providing various output formats optimized for different use cases.

## Installation

### Install from source
```bash
go install github.com/your-username/pr-review-cli@latest
```

### Build locally
```bash
git clone https://github.com/your-username/pr-review-cli
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

Fetch PR comments with Claude-optimized format (default):
```bash
pr-review-cli fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2
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
pr-review-cli fetch --owner microsoft --repo vscode --pr 123
pr-review-cli fetch --owner facebook --repo react --pr 456 --format human
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

The tool provides structured information about PR review comments including:
- Comment content and metadata
- File locations and line numbers in the diff
- Reviewer information
- Timestamps
- Summary statistics

The output format varies depending on the selected format option, with the Claude format being optimized for AI analysis and the human format being optimized for terminal readability.

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
- Fetch the PR review comments using the `pr-review-cli` tool
- Format the output for Claude to analyze and help address the feedback

### Example Workflow

1. Get PR review comments:
   ```
   /fetch-pr-comments url:https://github.com/your-org/your-repo/pull/123
   ```

2. Claude will fetch and display the comments, then you can ask Claude to help address them:
   - "Help me address the performance concerns raised in the review"
   - "Fix the code style issues mentioned in the comments"
   - "Explain how to resolve the security vulnerability pointed out"