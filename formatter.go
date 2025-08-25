package main

import (
	"encoding/json"
	"fmt"
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
	
	for i, comment := range response.Comments {
		output.WriteString(fmt.Sprintf("💬 Comment %d of %d\n", i+1, len(response.Comments)))
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
	
	for file, comments := range fileComments {
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
	
	return summary
}