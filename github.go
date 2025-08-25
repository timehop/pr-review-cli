package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	token string
}

// NewGitHubClient creates a new GitHub API client
// If token is provided, it will be used; otherwise falls back to GITHUB_TOKEN env var
func NewGitHubClient(token string) (*GitHubClient, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("GitHub token is required. Provide via --token flag or GITHUB_TOKEN environment variable")
		}
	}
	return &GitHubClient{token: token}, nil
}

// FetchPRComments fetches all review comments for a PR
func (c *GitHubClient) FetchPRComments(owner, repo string, prNumber int) ([]PRComment, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments", owner, repo, prNumber)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}
	
	var comments []PRComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	return comments, nil
}

// ParseDiffHunk parses a GitHub diff hunk string
func ParseDiffHunk(diffHunk string) (*DiffHunkInfo, error) {
	lines := strings.Split(diffHunk, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty diff hunk")
	}
	
	// Parse the @@ header line
	headerRegex := regexp.MustCompile(`@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@`)
	matches := headerRegex.FindStringSubmatch(lines[0])
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid diff hunk header: %s", lines[0])
	}
	
	oldStart, _ := strconv.Atoi(matches[1])
	oldCount := 1
	if matches[2] != "" {
		oldCount, _ = strconv.Atoi(matches[2])
	}
	
	newStart, _ := strconv.Atoi(matches[3])
	newCount := 1
	if matches[4] != "" {
		newCount, _ = strconv.Atoi(matches[4])
	}
	
	info := &DiffHunkInfo{
		OldStart: oldStart,
		OldCount: oldCount,
		NewStart: newStart,
		NewCount: newCount,
		Lines:    make([]DiffLine, 0),
	}
	
	// Parse the actual diff lines
	oldLine := oldStart
	newLine := newStart
	
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		
		diffLine := DiffLine{Content: line}
		
		switch {
		case strings.HasPrefix(line, "+"):
			diffLine.Type = "addition"
			diffLine.Content = line[1:] // Remove the + prefix
			diffLine.NewLine = &newLine
			newLine++
		case strings.HasPrefix(line, "-"):
			diffLine.Type = "deletion"
			diffLine.Content = line[1:] // Remove the - prefix
			diffLine.OldLine = &oldLine
			oldLine++
		case strings.HasPrefix(line, " "):
			diffLine.Type = "context"
			diffLine.Content = line[1:] // Remove the space prefix
			diffLine.OldLine = &oldLine
			diffLine.NewLine = &newLine
			oldLine++
			newLine++
		default:
			// Handle lines without prefix as context
			diffLine.Type = "context"
			diffLine.OldLine = &oldLine
			diffLine.NewLine = &newLine
			oldLine++
			newLine++
		}
		
		info.Lines = append(info.Lines, diffLine)
	}
	
	return info, nil
}

// ParseComments converts GitHub PR comments to Claude-friendly format
func ParseComments(comments []PRComment) ([]ParsedComment, error) {
	var parsed []ParsedComment
	
	for _, comment := range comments {
		parsedComment := ParsedComment{
			ID:        comment.ID,
			File:      comment.Path,
			Comment:   comment.Body,
			HTMLURL:   comment.HTMLURL,
			DiffHunk:  comment.DiffHunk,
			Author:    comment.User.Login,
			CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		
		// Set line numbers based on side
		if comment.Side == "RIGHT" {
			parsedComment.LineNew = comment.Line
			parsedComment.LineOld = comment.OriginalLine
		} else {
			parsedComment.LineNew = comment.OriginalLine
			parsedComment.LineOld = comment.Line
		}
		
		// Parse diff hunk to get more context
		if comment.DiffHunk != "" {
			diffInfo, err := ParseDiffHunk(comment.DiffHunk)
			if err == nil {
				// Find the line content and change type
				for _, diffLine := range diffInfo.Lines {
					// Match the line that the comment is on
					if (parsedComment.LineNew != nil && diffLine.NewLine != nil && *diffLine.NewLine == *parsedComment.LineNew) ||
						(parsedComment.LineOld != nil && diffLine.OldLine != nil && *diffLine.OldLine == *parsedComment.LineOld) {
						parsedComment.LineContent = diffLine.Content
						parsedComment.ChangeType = diffLine.Type
						break
					}
				}
				
				// Generate context description
				parsedComment.Context = generateContext(diffInfo, parsedComment)
			}
		}
		
		// Fallback change type determination
		if parsedComment.ChangeType == "" {
			if parsedComment.LineOld == nil {
				parsedComment.ChangeType = "addition"
			} else if parsedComment.LineNew == nil {
				parsedComment.ChangeType = "deletion"
			} else {
				parsedComment.ChangeType = "modification"
			}
		}
		
		parsed = append(parsed, parsedComment)
	}
	
	return parsed, nil
}

// generateContext creates a human-readable context description
func generateContext(diffInfo *DiffHunkInfo, comment ParsedComment) string {
	switch comment.ChangeType {
	case "addition":
		if comment.LineNew != nil {
			return fmt.Sprintf("New line %d added", *comment.LineNew)
		}
		return "Line addition"
	case "deletion":
		if comment.LineOld != nil {
			return fmt.Sprintf("Line %d deleted", *comment.LineOld)
		}
		return "Line deletion"
	case "modification":
		if comment.LineNew != nil && comment.LineOld != nil {
			return fmt.Sprintf("Line %d modified (was line %d)", *comment.LineNew, *comment.LineOld)
		}
		return "Line modification"
	default:
		return "Context line"
	}
}