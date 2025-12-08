---
allowed-tools: Bash(pr-review-cli:*), Bash(go install:*), Bash(command -v:*), Bash(test -n:*), Bash(source:*)
description: Fetch GitHub PR review comments in Claude-optimized format from a PR URL
---

I'll fetch the PR review comments from the provided GitHub PR URL and format them for analysis.

**Default Behavior:** By default, this fetches only **unresolved, non-outdated** review threads using GitHub's GraphQL API. This focuses on actionable feedback that still needs to be addressed.

**Requirements:**
- `GITHUB_TOKEN` environment variable must be set with a GitHub personal access token
- `pr-review-cli` must be installed (will attempt to install via `go install` if missing)

{{#if url}}
The PR URL is: {{url}}

Let me check the prerequisites and fetch the comments:

```bash
# Step 1: Check if pr-review-cli is available
if ! command -v pr-review-cli &> /dev/null; then
    echo "[!] pr-review-cli not found on PATH"
    
    # Check if go is installed
    if ! command -v go &> /dev/null; then
        echo ""
        echo "ERROR: Go is not installed. Please install Go first:"
        echo "  - macOS: brew install go"
        echo "  - Linux: https://go.dev/doc/install"
        echo ""
        echo "After installing Go, run this command again."
        exit 1
    fi
    
    echo "[*] Attempting to install via: go install github.com/AObuchow/pr-review-cli@latest"
    go install github.com/AObuchow/pr-review-cli@latest
    
    if [ $? -ne 0 ]; then
        echo "ERROR: Failed to install pr-review-cli"
        exit 1
    fi
    
    # Check if GOPATH/bin or GOBIN is in PATH
    GOBIN="${GOBIN:-${GOPATH:-$HOME/go}/bin}"
    if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
        echo ""
        echo "[!] Warning: $GOBIN is not in your PATH"
        echo "    Add the following to your shell config (~/.zshrc or ~/.bashrc):"
        echo ""
        echo "    export PATH=\"\$PATH:$GOBIN\""
        echo ""
        echo "    Then restart your terminal or source your config file."
        exit 1
    fi
    
    echo "[OK] pr-review-cli installed successfully"
fi
```

```bash
# Step 2: Check for GITHUB_TOKEN
if ! test -n "$GITHUB_TOKEN"; then
    echo ""
    echo "ERROR: GITHUB_TOKEN environment variable is not set."
    echo ""
    echo "To fix this:"
    echo "1. Create a GitHub personal access token at: https://github.com/settings/tokens"
    echo "   (Required scopes: repo for private repos, or public_repo for public repos only)"
    echo ""
    echo "2. Add it to your shell configuration:"
    if [[ "$SHELL" == *"zsh"* ]]; then
        echo "   echo 'export GITHUB_TOKEN=\"your_token_here\"' >> ~/.zshrc"
        echo ""
        echo "3. Then either restart your terminal or run: source ~/.zshrc"
    else
        echo "   echo 'export GITHUB_TOKEN=\"your_token_here\"' >> ~/.bashrc"
        echo ""
        echo "3. Then either restart your terminal or run: source ~/.bashrc"
    fi
    echo ""
    echo "4. Run this command again."
    exit 1
fi
```

```bash
# Step 3: Parse the URL and fetch comments
URL="{{url}}"
if [[ $URL =~ github\.com/([^/]+)/([^/]+)/pull/([0-9]+) ]]; then
    OWNER="${BASH_REMATCH[1]}"
    REPO="${BASH_REMATCH[2]}"
    PR_NUMBER="${BASH_REMATCH[3]}"

    echo "Fetching review threads for PR #$PR_NUMBER from $OWNER/$REPO..."
    echo ""
    echo "[*] Showing: Unresolved, non-outdated review threads (default)"
    echo "    To include resolved threads, add: --include-resolved"
    echo "    To include outdated threads, add: --include-outdated"
    echo "    To include general PR comments, add: --include-general"
    echo ""

    # Determine if we need to source shell config (in case token was just added)
    # This handles the case where GITHUB_TOKEN exists but pr-review-cli can't see it
    if [[ "$SHELL" == *"zsh"* ]] && [ -f "$HOME/.zshrc" ]; then
        source "$HOME/.zshrc" 2>/dev/null && pr-review-cli fetch --owner "$OWNER" --repo "$REPO" --pr "$PR_NUMBER" --format claude
    elif [ -f "$HOME/.bashrc" ]; then
        source "$HOME/.bashrc" 2>/dev/null && pr-review-cli fetch --owner "$OWNER" --repo "$REPO" --pr "$PR_NUMBER" --format claude
    else
        pr-review-cli fetch --owner "$OWNER" --repo "$REPO" --pr "$PR_NUMBER" --format claude
    fi
else
    echo "Invalid GitHub PR URL format. Expected: https://github.com/owner/repo/pull/123"
    exit 1
fi
```

The review threads have been fetched and formatted for analysis. I can now help you address any unresolved feedback.

**Available Options:**
- `--include-resolved`: Show both resolved and unresolved threads
- `--include-outdated`: Show threads on outdated code
- `--include-general`: Include general PR discussion comments (not attached to specific code)
- `--graphql=false`: Use legacy REST API (shows all comments, no status filtering)

{{else}}
Please provide a GitHub PR URL (e.g., https://github.com/owner/repo/pull/123) to fetch the review comments.

Usage: `/fetch-pr-comments url:https://github.com/AObuchow/Eclipse-Spectrum-Theme/pull/2`

**Requirements:**
- `GITHUB_TOKEN` environment variable with a GitHub personal access token
- `pr-review-cli` tool (will attempt auto-install via `go install` if missing)

**Default Behavior:** Fetches only unresolved, non-outdated review threads for focused analysis.
{{/if}}
