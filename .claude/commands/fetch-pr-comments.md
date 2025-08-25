---
allowed-tools: Bash(pr-review-cli:*)
description: Fetch GitHub PR review comments in Claude-optimized format from a PR URL
---

I'll fetch the PR review comments from the provided GitHub PR URL and format them for analysis.

First, let me parse the PR URL to extract the owner, repo, and PR number:

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
    
    echo "Fetching comments for PR #$PR_NUMBER from $OWNER/$REPO..."
    pr-review-cli fetch --owner "$OWNER" --repo "$REPO" --pr "$PR_NUMBER" --format claude
else
    echo "Invalid GitHub PR URL format. Expected: https://github.com/owner/repo/pull/123"
    exit 1
fi
```

The comments have been fetched and formatted for analysis. I can now help you address any issues or feedback mentioned in the review comments.
{{else}}
Please provide a GitHub PR URL (e.g., https://github.com/owner/repo/pull/123) to fetch the review comments.

Usage: `/fetch-pr-comments url:https://github.com/AObuchow/Eclipse-Spectrum-Theme/pull/123`
{{/if}}