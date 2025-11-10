---
allowed-tools: Bash(pr-review-cli:*)
description: Fetch GitHub PR review comments in Claude-optimized format from a PR URL
---

I'll fetch the PR review comments from the provided GitHub PR URL and format them for analysis.

**Default Behavior:** By default, this fetches only **unresolved, non-outdated** review threads using GitHub's GraphQL API. This focuses on actionable feedback that still needs to be addressed.

{{#if url}}
The PR URL is: {{url}}

Let me extract the details and fetch the comments:

```bash
# Parse the URL to extract owner, repo, and PR number
URL="{{url}}"
if [[ $URL =~ github\.com/([^/]+)/([^/]+)/pull/([0-9]+) ]]; then
    OWNER="${BASH_REMATCH[1]}"
    REPO="${BASH_REMATCH[2]}"
    PR_NUMBER="${BASH_REMATCH[3]}"

    echo "Fetching review threads for PR #$PR_NUMBER from $OWNER/$REPO..."
    echo ""
    echo "📌 Showing: Unresolved, non-outdated review threads (default)"
    echo "   To include resolved threads, add: --include-resolved"
    echo "   To include outdated threads, add: --include-outdated"
    echo "   To include general PR comments, add: --include-general"
    echo ""

    # Fetch using GraphQL API (default)
    # Only shows unresolved, non-outdated threads by default
    pr-review-cli fetch --owner "$OWNER" --repo "$REPO" --pr "$PR_NUMBER" --format claude
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

**Default Behavior:** Fetches only unresolved, non-outdated review threads for focused analysis.
{{/if}}