package main

import "time"

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// PRComment represents a GitHub PR review comment from the API
type PRComment struct {
	ID                   int        `json:"id"`
	PullRequestReviewID  int        `json:"pull_request_review_id"`
	DiffHunk             string     `json:"diff_hunk"`
	Path                 string     `json:"path"`
	CommitID             string     `json:"commit_id"`
	OriginalCommitID     string     `json:"original_commit_id"`
	User                 GitHubUser `json:"user"`
	Body                 string     `json:"body"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	HTMLURL              string     `json:"html_url"`
	PullRequestURL       string     `json:"pull_request_url"`
	StartLine            *int       `json:"start_line"`
	OriginalStartLine    *int       `json:"original_start_line"`
	Line                 *int       `json:"line"`
	OriginalLine         *int       `json:"original_line"`
	Side                 string     `json:"side"`
	OriginalPosition     *int       `json:"original_position"`
	Position             *int       `json:"position"`
}

// ParsedComment represents a comment with parsed diff context for Claude
type ParsedComment struct {
	ID            int    `json:"id"`
	File          string `json:"file"`
	LineNew       *int   `json:"line_new"`
	LineOld       *int   `json:"line_old"`
	ChangeType    string `json:"change_type"`
	LineContent   string `json:"line_content"`
	Comment       string `json:"comment"`
	Context       string `json:"context"`
	HTMLURL       string `json:"html_url"`
	DiffHunk      string `json:"diff_hunk"`
	Author        string `json:"author"`
	CreatedAt     string `json:"created_at"`
}

// DiffHunkInfo represents parsed diff hunk information
type DiffHunkInfo struct {
	OldStart   int
	OldCount   int
	NewStart   int
	NewCount   int
	Lines      []DiffLine
}

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    string // "addition", "deletion", "context"
	Content string
	OldLine *int
	NewLine *int
}

// PRCommentsResponse represents the tool's output
type PRCommentsResponse struct {
	PRNumber int             `json:"pr_number"`
	Owner    string          `json:"owner"`
	Repo     string          `json:"repo"`
	Comments []ParsedComment `json:"comments"`
	Summary  CommentsSummary `json:"summary"`
}

// CommentsSummary provides high-level information about the comments
type CommentsSummary struct {
	TotalComments int      `json:"total_comments"`
	FilesAffected []string `json:"files_affected"`
	Authors       []string `json:"authors"`
}