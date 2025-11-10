package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	
	switch command {
	case "fetch":
		handleFetch(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleFetch(args []string) {
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)

	// Basic flags
	owner := fetchCmd.String("owner", "", "GitHub repository owner")
	repo := fetchCmd.String("repo", "", "GitHub repository name")
	prNumber := fetchCmd.Int("pr", 0, "Pull request number")
	format := fetchCmd.String("format", "claude", "Output format: json, human, claude")
	token := fetchCmd.String("token", "", "GitHub personal access token (optional if GITHUB_TOKEN env var is set)")

	// GraphQL flags
	useGraphQL := fetchCmd.Bool("graphql", true, "Use GraphQL API (default: true, set to false for legacy REST)")
	includeResolved := fetchCmd.Bool("include-resolved", false, "Include resolved review threads (default: only unresolved)")
	includeOutdated := fetchCmd.Bool("include-outdated", false, "Include outdated review threads (default: exclude outdated)")
	includeGeneral := fetchCmd.Bool("include-general", false, "Include general PR comments (default: only review threads)")

	fetchCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s fetch --owner OWNER --repo REPO --pr PR_NUMBER [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Fetch and parse GitHub PR review comments.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fetchCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_TOKEN    GitHub personal access token (not required if --token is used)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Basic usage (GraphQL, unresolved threads only)\n")
		fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Include resolved threads\n")
		fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --include-resolved\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Include outdated threads\n")
		fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --include-outdated\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Include general PR comments\n")
		fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --include-general\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use legacy REST API\n")
		fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --graphql=false\n", os.Args[0])
	}

	if err := fetchCmd.Parse(args); err != nil {
		os.Exit(1)
	}

	// Validate required arguments
	if *owner == "" || *repo == "" || *prNumber == 0 {
		fmt.Fprintf(os.Stderr, "Error: --owner, --repo, and --pr are required\n\n")
		fetchCmd.Usage()
		os.Exit(1)
	}

	// Validate token availability
	if *token == "" && os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Fprintf(os.Stderr, "Error: GitHub token is required. Provide via --token flag or GITHUB_TOKEN environment variable\n\n")
		fetchCmd.Usage()
		os.Exit(1)
	}

	// Validate format
	validFormats := map[string]bool{"json": true, "human": true, "claude": true}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Error: Invalid format '%s'. Must be one of: json, human, claude\n", *format)
		os.Exit(1)
	}

	var response *PRCommentsResponse

	if *useGraphQL {
		// GraphQL path (new default)
		client, err := NewGitHubGraphQLClient(*token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating GitHub GraphQL client: %v\n", err)
			os.Exit(1)
		}

		opts := FetchOptions{
			IncludeResolved: *includeResolved,
			IncludeOutdated: *includeOutdated,
			IncludeGeneral:  *includeGeneral,
		}

		threads, generalComments, err := client.FetchPRReviewThreads(
			context.Background(),
			*owner, *repo, *prNumber,
			opts,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching PR review threads: %v\n", err)
			os.Exit(1)
		}

		response = &PRCommentsResponse{
			PRNumber:        *prNumber,
			Owner:           *owner,
			Repo:            *repo,
			ReviewThreads:   threads,
			GeneralComments: generalComments,
			Summary:         GenerateThreadSummary(threads, generalComments, *owner, *repo, *prNumber),
		}

		// Format and output using V2 formatters
		output, err := FormatCommentsV2(response, *format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(output)
	} else {
		// REST path (legacy)
		client, err := NewGitHubClient(*token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating GitHub client: %v\n", err)
			os.Exit(1)
		}

		// Fetch comments
		comments, err := client.FetchPRComments(*owner, *repo, *prNumber)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching PR comments: %v\n", err)
			os.Exit(1)
		}

		// Parse comments
		parsedComments, err := ParseComments(comments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing comments: %v\n", err)
			os.Exit(1)
		}

		// Generate response
		response = &PRCommentsResponse{
			PRNumber: *prNumber,
			Owner:    *owner,
			Repo:     *repo,
			Comments: parsedComments,
			Summary:  GenerateSummary(parsedComments, *owner, *repo, *prNumber),
		}

		// Format and output using V1 formatters
		output, err := FormatComments(response, *format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(output)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "GitHub PR Review Comments CLI Tool\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s COMMAND [OPTIONS]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  fetch     Fetch and parse PR review comments\n")
	fmt.Fprintf(os.Stderr, "  help      Show this help message\n\n")
	fmt.Fprintf(os.Stderr, "For command-specific help, use: %s COMMAND --help\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Environment Variables:\n")
	fmt.Fprintf(os.Stderr, "  GITHUB_TOKEN    GitHub personal access token (not required if --token is used)\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --format human\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s fetch --owner AObuchow --repo Eclipse-Spectrum-Theme --pr 2 --token ghp_xxx\n", os.Args[0])
}

