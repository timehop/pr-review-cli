package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GitHubGraphQLClient handles GitHub GraphQL API interactions
type GitHubGraphQLClient struct {
	client *githubv4.Client
}

// NewGitHubGraphQLClient creates a new GitHub GraphQL API client
func NewGitHubGraphQLClient(token string) (*GitHubGraphQLClient, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("GitHub token is required")
		}
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	return &GitHubGraphQLClient{
		client: githubv4.NewClient(httpClient),
	}, nil
}

// FetchOptions configures what data to fetch
type FetchOptions struct {
	IncludeResolved bool
	IncludeOutdated bool
	IncludeGeneral  bool
}

// FetchPRReviewThreads fetches PR review threads with filtering
func (c *GitHubGraphQLClient) FetchPRReviewThreads(
	ctx context.Context,
	owner, repo string,
	prNumber int,
	opts FetchOptions,
) ([]ReviewThread, []GeneralComment, error) {

	var allThreads []ReviewThread
	var allGeneralComments []GeneralComment

	// Fetch review threads with pagination
	err := c.fetchReviewThreads(ctx, owner, repo, prNumber, opts, &allThreads)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching review threads: %w", err)
	}

	// Fetch general comments if requested
	if opts.IncludeGeneral {
		err = c.fetchGeneralComments(ctx, owner, repo, prNumber, &allGeneralComments)
		if err != nil {
			return nil, nil, fmt.Errorf("fetching general comments: %w", err)
		}
	}

	return allThreads, allGeneralComments, nil
}

// fetchReviewThreads fetches all review threads with pagination
func (c *GitHubGraphQLClient) fetchReviewThreads(
	ctx context.Context,
	owner, repo string,
	prNumber int,
	opts FetchOptions,
	allThreads *[]ReviewThread,
) error {
	var cursor *githubv4.String

	for {
		var query struct {
			Repository struct {
				PullRequest struct {
					ReviewThreads struct {
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
						Nodes []struct {
							ID         githubv4.String
							IsResolved bool
							IsOutdated bool
							Path       githubv4.String
							Line       *githubv4.Int
							StartLine  *githubv4.Int
							DiffSide   githubv4.String
							Comments   struct {
								PageInfo struct {
									EndCursor   githubv4.String
									HasNextPage bool
								}
								Nodes []struct {
									ID        githubv4.String
									Body      githubv4.String
									CreatedAt githubv4.DateTime
									Author    struct {
										Login githubv4.String
									}
									DiffHunk githubv4.String
									ReplyTo  *struct {
										ID githubv4.String
									}
									URL githubv4.URI
								}
							} `graphql:"comments(first: 100)"`
						}
					} `graphql:"reviewThreads(first: 100, after: $cursor)"`
				} `graphql:"pullRequest(number: $prNumber)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables := map[string]interface{}{
			"owner":    githubv4.String(owner),
			"name":     githubv4.String(repo),
			"prNumber": githubv4.Int(prNumber),
			"cursor":   cursor,
		}

		err := c.client.Query(ctx, &query, variables)
		if err != nil {
			return fmt.Errorf("GraphQL query error for PR #%d in %s/%s: %w", prNumber, owner, repo, err)
		}

		// Process and filter threads
		for _, thread := range query.Repository.PullRequest.ReviewThreads.Nodes {
			// Apply filtering logic
			if !opts.IncludeResolved && thread.IsResolved {
				continue
			}
			if !opts.IncludeOutdated && thread.IsOutdated {
				continue
			}

			reviewThread := ReviewThread{
				ID:         string(thread.ID),
				File:       string(thread.Path),
				IsResolved: thread.IsResolved,
				IsOutdated: thread.IsOutdated,
				Comments:   make([]ThreadComment, 0, len(thread.Comments.Nodes)),
			}

			// Calculate line numbers based on DiffSide
			// RIGHT = new file, LEFT = old file
			if thread.Line != nil {
				lineNum := int(*thread.Line)
				if strings.ToUpper(string(thread.DiffSide)) == "RIGHT" {
					reviewThread.LineNew = &lineNum
				} else {
					reviewThread.LineOld = &lineNum
				}
			}

			if thread.StartLine != nil {
				startLineNum := int(*thread.StartLine)
				if strings.ToUpper(string(thread.DiffSide)) == "RIGHT" {
					reviewThread.StartLineNew = &startLineNum
				} else {
					reviewThread.StartLineOld = &startLineNum
				}
			}

			// Process all comments in the thread
			for _, comment := range thread.Comments.Nodes {
				threadComment := ThreadComment{
					ID:        string(comment.ID),
					Body:      string(comment.Body),
					Author:    string(comment.Author.Login),
					CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
					HTMLURL:   comment.URL.String(),
					IsReply:   comment.ReplyTo != nil,
				}

				reviewThread.Comments = append(reviewThread.Comments, threadComment)

				// Capture diff hunk from first comment
				if len(reviewThread.Comments) == 1 && comment.DiffHunk != "" {
					reviewThread.DiffHunk = string(comment.DiffHunk)
				}
			}

			// Fetch additional comment pages if needed
			if thread.Comments.PageInfo.HasNextPage {
				if err := c.fetchAdditionalThreadComments(
					ctx,
					thread.ID,
					thread.Comments.PageInfo.EndCursor,
					&reviewThread,
				); err != nil {
					return fmt.Errorf("fetching additional comments for thread %s: %w", thread.ID, err)
				}
			}

			*allThreads = append(*allThreads, reviewThread)
		}

		if !query.Repository.PullRequest.ReviewThreads.PageInfo.HasNextPage {
			break
		}
		cursor = githubv4.NewString(query.Repository.PullRequest.ReviewThreads.PageInfo.EndCursor)
	}

	return nil
}

// fetchAdditionalThreadComments paginates through remaining comments for a thread
func (c *GitHubGraphQLClient) fetchAdditionalThreadComments(
	ctx context.Context,
	threadID githubv4.String,
	startCursor githubv4.String,
	reviewThread *ReviewThread,
) error {
	var cursor *githubv4.String
	if startCursor != "" {
		cursor = githubv4.NewString(startCursor)
	}

	for {
		var query struct {
			Node struct {
				PullRequestReviewThread struct {
					Comments struct {
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
						Nodes []struct {
							ID        githubv4.String
							Body      githubv4.String
							CreatedAt githubv4.DateTime
							Author    struct {
								Login githubv4.String
							}
							DiffHunk githubv4.String
							ReplyTo  *struct {
								ID githubv4.String
							}
							URL githubv4.URI
						}
					} `graphql:"comments(first: 100, after: $cursor)"`
				} `graphql:"... on PullRequestReviewThread"`
			} `graphql:"node(id: $threadID)"`
		}

		variables := map[string]interface{}{
			"threadID": githubv4.ID(threadID),
			"cursor":   cursor,
		}

		if err := c.client.Query(ctx, &query, variables); err != nil {
			return fmt.Errorf("GraphQL thread comment query error: %w", err)
		}

		node := query.Node.PullRequestReviewThread
		for _, comment := range node.Comments.Nodes {
			threadComment := ThreadComment{
				ID:        string(comment.ID),
				Body:      string(comment.Body),
				Author:    string(comment.Author.Login),
				CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
				HTMLURL:   comment.URL.String(),
				IsReply:   comment.ReplyTo != nil,
			}

			reviewThread.Comments = append(reviewThread.Comments, threadComment)

			if len(reviewThread.Comments) == 1 && comment.DiffHunk != "" {
				reviewThread.DiffHunk = string(comment.DiffHunk)
			}
		}

		if !node.Comments.PageInfo.HasNextPage {
			break
		}
		cursor = githubv4.NewString(node.Comments.PageInfo.EndCursor)
	}

	return nil
}

// fetchGeneralComments fetches all general PR comments with pagination
func (c *GitHubGraphQLClient) fetchGeneralComments(
	ctx context.Context,
	owner, repo string,
	prNumber int,
	allComments *[]GeneralComment,
) error {
	var cursor *githubv4.String

	for {
		var query struct {
			Repository struct {
				PullRequest struct {
					Comments struct {
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
						Nodes []struct {
							ID        githubv4.String
							Body      githubv4.String
							CreatedAt githubv4.DateTime
							Author    struct {
								Login githubv4.String
							}
							URL githubv4.URI
						}
					} `graphql:"comments(first: 100, after: $cursor)"`
				} `graphql:"pullRequest(number: $prNumber)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables := map[string]interface{}{
			"owner":    githubv4.String(owner),
			"name":     githubv4.String(repo),
			"prNumber": githubv4.Int(prNumber),
			"cursor":   cursor,
		}

		err := c.client.Query(ctx, &query, variables)
		if err != nil {
			return fmt.Errorf("GraphQL query error for PR #%d general comments: %w", prNumber, err)
		}

		// Process general comments
		for _, comment := range query.Repository.PullRequest.Comments.Nodes {
			generalComment := GeneralComment{
				ID:        string(comment.ID),
				Body:      string(comment.Body),
				Author:    string(comment.Author.Login),
				CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
				HTMLURL:   comment.URL.String(),
			}
			*allComments = append(*allComments, generalComment)
		}

		if !query.Repository.PullRequest.Comments.PageInfo.HasNextPage {
			break
		}
		cursor = githubv4.NewString(query.Repository.PullRequest.Comments.PageInfo.EndCursor)
	}

	return nil
}
