package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// FormatComments formats comments for human-readable output
func FormatComments(response *PRCommentsResponse, format string) (string, error) {
	switch format {
	case "json":
		return formatJSON(response)
	case "human":
		return formatHuman(response)
	case "claude":
		return formatClaude(response)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatJSON outputs the response as JSON
func formatJSON(response *PRCommentsResponse) (string, error) {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	return string(data), nil
}

// formatHuman outputs human-readable format
func formatHuman(response *PRCommentsResponse) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("🔍 PR #%d Review Comments (%s/%s)\n",
		response.PRNumber, response.Owner, response.Repo))
	output.WriteString(fmt.Sprintf("📊 %d comments across %d files\n",
		response.Summary.TotalComments, len(response.Summary.FilesAffected)))
	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	sortedComments := sortParsedComments(response.Comments)
	for i, comment := range sortedComments {
		output.WriteString(fmt.Sprintf("💬 Comment %d of %d\n", i+1, len(sortedComments)))
		output.WriteString(fmt.Sprintf("📁 %s", comment.File))

		// Add line information
		if comment.LineNew != nil {
			output.WriteString(fmt.Sprintf(":%d", *comment.LineNew))
		}

		// Add change type indicator
		switch comment.ChangeType {
		case "addition":
			output.WriteString(" ➕ (new)")
		case "deletion":
			output.WriteString(" ➖ (deleted)")
		case "modification":
			output.WriteString(" ✏️  (modified)")
		default:
			output.WriteString(" 📝 (context)")
		}
		output.WriteString("\n")

		// Add context
		output.WriteString(fmt.Sprintf("📍 %s\n", comment.Context))

		// Add line content if available
		if comment.LineContent != "" {
			output.WriteString(fmt.Sprintf("📄 Line: `%s`\n", strings.TrimSpace(comment.LineContent)))
		}

		// Add comment body
		output.WriteString(fmt.Sprintf("💭 %s: %s\n", comment.Author, comment.Comment))

		// Add link
		output.WriteString(fmt.Sprintf("🔗 %s\n", comment.HTMLURL))

		// Add separator
		if i < len(response.Comments)-1 {
			output.WriteString("───────────────────────────────────────────────────────────────\n\n")
		}
	}

	if len(response.Comments) == 0 {
		output.WriteString("✅ No review comments found for this PR\n")
	}

	return output.String(), nil
}

// formatClaude outputs Claude-optimized format for processing
func formatClaude(response *PRCommentsResponse) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("# PR #%d Review Comments Analysis\n\n", response.PRNumber))
	output.WriteString(fmt.Sprintf("**Repository:** %s/%s\n", response.Owner, response.Repo))
	output.WriteString(fmt.Sprintf("**Total Comments:** %d\n", response.Summary.TotalComments))
	output.WriteString(fmt.Sprintf("**Files Affected:** %s\n\n", strings.Join(response.Summary.FilesAffected, ", ")))

	if len(response.Comments) == 0 {
		output.WriteString("✅ **No review comments to address**\n")
		return output.String(), nil
	}

	output.WriteString("## Comments to Address:\n\n")

	// Group comments by file
	fileComments := make(map[string][]ParsedComment)
	for _, comment := range response.Comments {
		fileComments[comment.File] = append(fileComments[comment.File], comment)
	}

	files := make([]string, 0, len(fileComments))
	for file := range fileComments {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		comments := sortParsedComments(fileComments[file])
		output.WriteString(fmt.Sprintf("### 📁 `%s`\n\n", file))

		for _, comment := range comments {
			output.WriteString(fmt.Sprintf("**Line %s** - %s\n",
				formatLineInfo(comment), comment.Context))

			// Add diff context if helpful
			if comment.LineContent != "" {
				output.WriteString(fmt.Sprintf("```\n%s\n```\n", comment.LineContent))
			}

			output.WriteString(fmt.Sprintf("**Review Comment:** %s\n\n", comment.Comment))
			output.WriteString(fmt.Sprintf("**Action Required:** Address the feedback on line %s\n",
				formatLineInfo(comment)))
			output.WriteString(fmt.Sprintf("**Reference:** [View on GitHub](%s)\n\n", comment.HTMLURL))

			output.WriteString("---\n\n")
		}
	}

	// Add summary of next steps
	output.WriteString("## Next Steps:\n\n")
	output.WriteString("1. Review each comment above\n")
	output.WriteString("2. Make the necessary changes to address the feedback\n")
	output.WriteString("3. Commit and push your changes\n")
	output.WriteString("4. Notify the reviewer that changes are ready for re-review\n")

	return output.String(), nil
}

// formatLineInfo formats line number information for display
func formatLineInfo(comment ParsedComment) string {
	if comment.LineNew != nil {
		return fmt.Sprintf("%d", *comment.LineNew)
	}
	if comment.LineOld != nil {
		return fmt.Sprintf("%d (deleted)", *comment.LineOld)
	}
	return "unknown"
}

// GenerateSummary creates a summary of the comments
func GenerateSummary(comments []ParsedComment, owner, repo string, prNumber int) CommentsSummary {
	summary := CommentsSummary{
		TotalComments: len(comments),
		FilesAffected: make([]string, 0),
		Authors:       make([]string, 0),
	}

	// Track unique files and authors
	fileSet := make(map[string]bool)
	authorSet := make(map[string]bool)

	for _, comment := range comments {
		if !fileSet[comment.File] {
			fileSet[comment.File] = true
			summary.FilesAffected = append(summary.FilesAffected, comment.File)
		}

		if !authorSet[comment.Author] {
			authorSet[comment.Author] = true
			summary.Authors = append(summary.Authors, comment.Author)
		}
	}

	sort.Strings(summary.FilesAffected)
	sort.Strings(summary.Authors)

	return summary
}

// ============================================================================
// V2 Formatters for Thread-based GraphQL responses
// ============================================================================

// FormatCommentsV2 formats thread-based comments from GraphQL API
func FormatCommentsV2(response *PRCommentsResponse, format string) (string, error) {
	switch format {
	case "json":
		return formatJSONV2(response)
	case "human":
		return formatHumanV2(response)
	case "claude":
		return formatClaudeV2(response)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatJSONV2 outputs thread-based response as JSON
func formatJSONV2(response *PRCommentsResponse) (string, error) {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	return string(data), nil
}

// formatHumanV2 outputs human-readable thread format
func formatHumanV2(response *PRCommentsResponse) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("🔍 PR #%d Review Comments (%s/%s)\n",
		response.PRNumber, response.Owner, response.Repo))

	// Enhanced summary with thread counts
	output.WriteString(fmt.Sprintf("📊 %d unresolved threads, %d resolved\n",
		response.Summary.UnresolvedThreads,
		response.Summary.ResolvedThreads))

	if response.Summary.OutdatedThreads > 0 {
		output.WriteString(fmt.Sprintf("⚠️  %d outdated threads\n",
			response.Summary.OutdatedThreads))
	}

	if response.Summary.GeneralComments > 0 {
		output.WriteString(fmt.Sprintf("💬 %d general comments\n",
			response.Summary.GeneralComments))
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Display general comments first (if any)
	if len(response.GeneralComments) > 0 {
		output.WriteString("## General PR Comments\n\n")
		for i, comment := range response.GeneralComments {
			output.WriteString(fmt.Sprintf("💬 Comment %d\n", i+1))
			output.WriteString(fmt.Sprintf("👤 %s: %s\n", comment.Author, comment.Body))
			output.WriteString(fmt.Sprintf("🔗 %s\n", comment.HTMLURL))
			output.WriteString("───────────────────────────────────────────────────────────────\n\n")
		}
	}

	// Display review threads
	sortedThreads := sortReviewThreads(response.ReviewThreads)
	if len(sortedThreads) > 0 {
		output.WriteString("## Review Threads\n\n")

		for i, thread := range sortedThreads {
			// Thread header with status
			statusEmoji := "❌"
			statusText := "[UNRESOLVED]"
			if thread.IsResolved {
				statusEmoji = "✅"
				statusText = "[RESOLVED] ✓"
			}

			output.WriteString(fmt.Sprintf("%s Thread %d of %d %s\n",
				statusEmoji, i+1, len(response.ReviewThreads), statusText))

			if thread.IsOutdated {
				output.WriteString("⚠️  [OUTDATED]\n")
			}

			output.WriteString(fmt.Sprintf("📁 %s", thread.File))
			if thread.LineNew != nil {
				output.WriteString(fmt.Sprintf(":%d", *thread.LineNew))
			}
			output.WriteString("\n")

			writeDiffSnippet(&output, thread.DiffHunk)

			// Display all comments in thread
			for j, comment := range thread.Comments {
				indent := ""
				if comment.IsReply {
					indent = "  ↳ "
				}
				output.WriteString(fmt.Sprintf("%s💭 %s: %s\n",
					indent, comment.Author, comment.Body))

				if j == 0 {
					output.WriteString(fmt.Sprintf("  🔗 %s\n", comment.HTMLURL))
				}
			}

			output.WriteString("───────────────────────────────────────────────────────────────\n\n")
		}
	}

	if len(sortedThreads) == 0 && len(response.GeneralComments) == 0 {
		output.WriteString("✅ No comments found for this PR\n")
	}

	return output.String(), nil
}

// formatClaudeV2 outputs Claude-optimized thread format
func formatClaudeV2(response *PRCommentsResponse) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("# PR #%d Review Comments Analysis\n\n", response.PRNumber))
	output.WriteString(fmt.Sprintf("**Repository:** %s/%s\n\n", response.Owner, response.Repo))

	// Enhanced statistics
	output.WriteString("## Summary Statistics\n\n")
	output.WriteString(fmt.Sprintf("- **Unresolved Threads:** %d\n", response.Summary.UnresolvedThreads))
	output.WriteString(fmt.Sprintf("- **Resolved Threads:** %d\n", response.Summary.ResolvedThreads))

	if response.Summary.OutdatedThreads > 0 {
		output.WriteString(fmt.Sprintf("- **Outdated Threads:** %d\n", response.Summary.OutdatedThreads))
	}

	if response.Summary.GeneralComments > 0 {
		output.WriteString(fmt.Sprintf("- **General Comments:** %d\n", response.Summary.GeneralComments))
	}

	output.WriteString(fmt.Sprintf("- **Files Affected:** %s\n\n", strings.Join(response.Summary.FilesAffected, ", ")))

	// Show general comments first
	if len(response.GeneralComments) > 0 {
		output.WriteString("## General PR Discussion\n\n")
		for _, comment := range response.GeneralComments {
			output.WriteString(fmt.Sprintf("**%s** wrote:\n\n", comment.Author))
			output.WriteString(fmt.Sprintf("> %s\n\n", comment.Body))
			output.WriteString(fmt.Sprintf("**Reference:** [View on GitHub](%s)\n\n", comment.HTMLURL))
			output.WriteString("---\n\n")
		}
	}

	// Early exit if no threads to address
	if len(response.ReviewThreads) == 0 {
		if len(response.GeneralComments) == 0 {
			output.WriteString("✅ **No review comments to address**\n")
		}
		return output.String(), nil
	}

	output.WriteString("## Review Threads to Address\n\n")

	// Group threads by file
	fileThreads := make(map[string][]ReviewThread)
	for _, thread := range response.ReviewThreads {
		fileThreads[thread.File] = append(fileThreads[thread.File], thread)
	}

	files := make([]string, 0, len(fileThreads))
	for file := range fileThreads {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		threads := sortReviewThreads(fileThreads[file])
		output.WriteString(fmt.Sprintf("### 📁 `%s`\n\n", file))

		for _, thread := range threads {
			// Thread status header
			if thread.IsResolved {
				output.WriteString("#### [RESOLVED] ✓ ")
			} else {
				output.WriteString("#### [UNRESOLVED] ")
			}

			// Line information
			if thread.LineNew != nil {
				output.WriteString(fmt.Sprintf("Line %d", *thread.LineNew))
			} else {
				output.WriteString("General file comment")
			}

			if thread.IsOutdated {
				output.WriteString(" ⚠️ **[OUTDATED]**")
			}
			output.WriteString("\n\n")

			writeDiffSnippet(&output, thread.DiffHunk)

			// Thread conversation
			output.WriteString("**Conversation:**\n\n")
			for i, comment := range thread.Comments {
				if comment.IsReply {
					output.WriteString(fmt.Sprintf("↳ **%s** replied:\n", comment.Author))
				} else {
					output.WriteString(fmt.Sprintf("**%s** commented:\n", comment.Author))
				}
				output.WriteString(fmt.Sprintf("> %s\n\n", comment.Body))

				// Only show link for first comment
				if i == 0 {
					output.WriteString(fmt.Sprintf("**Reference:** [View on GitHub](%s)\n\n", comment.HTMLURL))
				}
			}

			// Action item (only for unresolved threads)
			if !thread.IsResolved {
				output.WriteString("**Action Required:**\n")
				lineInfo := "this section"
				if thread.LineNew != nil {
					lineInfo = fmt.Sprintf("line %d", *thread.LineNew)
				}
				output.WriteString(fmt.Sprintf("- Address the feedback on %s\n", lineInfo))
				output.WriteString("- Reply to the thread when changes are made\n\n")
			}

			output.WriteString("---\n\n")
		}
	}

	// Next steps section
	if response.Summary.UnresolvedThreads > 0 {
		output.WriteString("## Next Steps\n\n")
		output.WriteString("1. Review each unresolved thread above\n")
		output.WriteString("2. Make the necessary code changes to address the feedback\n")
		output.WriteString("3. Reply to threads as you resolve them\n")
		output.WriteString("4. Push your changes and notify reviewers\n")
	}

	return output.String(), nil
}

// writeDiffSnippet renders a diff hunk with consistent formatting
func writeDiffSnippet(builder *strings.Builder, diffHunk string) {
	trimmed := strings.TrimSpace(diffHunk)
	if trimmed == "" {
		return
	}

	builder.WriteString("```diff\n")
	builder.WriteString(trimmed)
	builder.WriteString("\n```\n\n")
}

// GenerateThreadSummary creates summary for thread-based response
func GenerateThreadSummary(
	threads []ReviewThread,
	generalComments []GeneralComment,
	owner, repo string,
	prNumber int,
) CommentsSummary {

	summary := CommentsSummary{
		TotalComments:     0,
		UnresolvedThreads: 0,
		ResolvedThreads:   0,
		OutdatedThreads:   0,
		GeneralComments:   len(generalComments),
		FilesAffected:     make([]string, 0),
		Authors:           make([]string, 0),
	}

	fileSet := make(map[string]bool)
	authorSet := make(map[string]bool)

	// Process review threads
	for _, thread := range threads {
		if thread.IsResolved {
			summary.ResolvedThreads++
		} else {
			summary.UnresolvedThreads++
		}

		if thread.IsOutdated {
			summary.OutdatedThreads++
		}

		// Count all comments in thread
		summary.TotalComments += len(thread.Comments)

		// Track files
		if !fileSet[thread.File] {
			fileSet[thread.File] = true
			summary.FilesAffected = append(summary.FilesAffected, thread.File)
		}

		// Track authors
		for _, comment := range thread.Comments {
			if !authorSet[comment.Author] {
				authorSet[comment.Author] = true
				summary.Authors = append(summary.Authors, comment.Author)
			}
		}
	}

	// Process general comments
	for _, comment := range generalComments {
		if !authorSet[comment.Author] {
			authorSet[comment.Author] = true
			summary.Authors = append(summary.Authors, comment.Author)
		}
	}

	summary.TotalComments += len(generalComments)

	sort.Strings(summary.FilesAffected)
	sort.Strings(summary.Authors)

	return summary
}

// sortReviewThreads returns a deterministically ordered copy of the threads
func sortReviewThreads(threads []ReviewThread) []ReviewThread {
	sorted := make([]ReviewThread, len(threads))
	copy(sorted, threads)

	sort.SliceStable(sorted, func(i, j int) bool {
		a := sorted[i]
		b := sorted[j]

		if a.File != b.File {
			return a.File < b.File
		}

		if a.IsResolved != b.IsResolved {
			return !a.IsResolved && b.IsResolved
		}

		lineA := threadLineNumber(a)
		lineB := threadLineNumber(b)
		if lineA != lineB {
			return lineA < lineB
		}

		return a.ID < b.ID
	})

	return sorted
}

func sortParsedComments(comments []ParsedComment) []ParsedComment {
	sorted := make([]ParsedComment, len(comments))
	copy(sorted, comments)

	sort.SliceStable(sorted, func(i, j int) bool {
		a := sorted[i]
		b := sorted[j]

		if a.File != b.File {
			return a.File < b.File
		}

		lineA := commentLineNumber(a)
		lineB := commentLineNumber(b)
		if lineA != lineB {
			return lineA < lineB
		}

		return a.ID < b.ID
	})

	return sorted
}

func threadLineNumber(thread ReviewThread) int {
	if thread.LineNew != nil {
		return *thread.LineNew
	}
	if thread.LineOld != nil {
		return *thread.LineOld
	}
	if thread.StartLineNew != nil {
		return *thread.StartLineNew
	}
	if thread.StartLineOld != nil {
		return *thread.StartLineOld
	}
	return math.MaxInt32
}

func commentLineNumber(comment ParsedComment) int {
	if comment.LineNew != nil {
		return *comment.LineNew
	}
	if comment.LineOld != nil {
		return *comment.LineOld
	}
	return math.MaxInt32
}
