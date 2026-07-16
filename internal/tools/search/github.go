package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nu/internal/contracts"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// GitHubContentTool extracts matching files from GitHub repositories.
type GitHubContentTool struct {
	client *github.Client
}

// GitHubFileContent is one extracted repository file.
type GitHubFileContent struct {
	FileName string `json:"file_name"`
	Path     string `json:"path"`
	Content  string `json:"content"`
}

// GitHubSearchParams controls repository traversal and extraction.
type GitHubSearchParams struct {
	RepositoryURL string   `json:"repository_url"`
	FilePatterns  []string `json:"file_patterns"`
	MaxFiles      int      `json:"max_files,omitempty"`
	MaxDepth      int      `json:"max_depth,omitempty"`
	MaxFileSize   int64    `json:"max_file_size,omitempty"`
	SpecificFiles []string `json:"specific_files,omitempty"`
}

// NewGitHubContent creates a GitHub content tool.
func NewGitHubContent(token string) (*GitHubContentTool, error) {
	var client *github.Client
	if token == "" {
		client = github.NewClient(nil)
	} else {
		source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = github.NewClient(oauth2.NewClient(context.Background(), source))
	}
	return &GitHubContentTool{client: client}, nil
}

func (t *GitHubContentTool) Name() string        { return "github_content_extractor" }
func (t *GitHubContentTool) DisplayName() string { return "GitHub Content Extractor" }
func (t *GitHubContentTool) Description() string {
	return "Extracts content from GitHub repositories based on file patterns"
}
func (t *GitHubContentTool) Internal() bool { return false }
func (t *GitHubContentTool) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"repository_url": {
			Type: "string", Description: "The GitHub repository URL to analyze", Required: true,
		},
		"file_patterns": {
			Type: "array", Description: "List of file patterns to match (e.g., ['.json', '.yaml']). Use only the following extensions: .json, .yaml, .yml, .go, .py, .js, .ts, .java, .rb, .php, .cs, .cpp, .c, .h, .hpp, .swift, .kt, .scala, .rs, .dart, .lua, .proto, .thrift, .avro, .graphql, .sql, .tf, .tfvars, .hcl, .tfstate, .k8s.yaml, .k8s.yml", Required: true,
			Items: &contracts.ParameterSpec{Type: "string"},
		},
		"max_files": {
			Type: "number", Description: "Maximum number of files to extract",
		},
		"max_depth": {
			Type: "number", Description: "Maximum depth of directories to traverse",
		},
		"max_file_size": {
			Type: "number", Description: "Maximum file size to extract (in bytes)",
		},
		"specific_files": {
			Type: "array", Description: "List of specific files to extract",
			Items: &contracts.ParameterSpec{Type: "string"},
		},
	}
}

func (t *GitHubContentTool) Run(ctx context.Context, input string) (string, error) {
	var params GitHubSearchParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}
	owner, repo, err := extractRepoInfo(params.RepositoryURL)
	if err != nil {
		return "", fmt.Errorf("failed to extract repository info: %w", err)
	}
	contents, err := t.getRepositoryContents(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get repository contents: %w", err)
	}
	files, err := t.findMatchingFiles(ctx, owner, repo, contents, params)
	if err != nil {
		return "", fmt.Errorf("failed to find matching files: %w", err)
	}
	result, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}
	return string(result), nil
}

func (t *GitHubContentTool) Execute(ctx context.Context, args string) (string, error) {
	return t.Run(ctx, args)
}

func (t *GitHubContentTool) findMatchingFiles(ctx context.Context, owner, repo string, contents []*github.RepositoryContent, params GitHubSearchParams) ([]GitHubFileContent, error) {
	var matches []GitHubFileContent
	var walk func([]*github.RepositoryContent, int) error
	walk = func(contents []*github.RepositoryContent, depth int) error {
		if params.MaxDepth > 0 && depth > params.MaxDepth {
			return nil
		}
		if params.MaxFiles > 0 && len(matches) >= params.MaxFiles {
			return errMaximumFiles
		}
		for _, content := range contents {
			if content == nil || content.Type == nil {
				continue
			}
			if content.GetType() == "dir" {
				file, directory, _, err := t.client.Repositories.GetContents(ctx, owner, repo, content.GetPath(), nil)
				if err != nil {
					return fmt.Errorf("failed to get contents of directory %s: %w", content.GetPath(), err)
				}
				children := directory
				if file != nil {
					children = []*github.RepositoryContent{file}
				}
				if err := walk(children, depth+1); err == errMaximumFiles {
					return err
				}
				continue
			}
			if content.GetType() != "file" || !matchesSpecificFile(content.GetPath(), params.SpecificFiles) || !matchesFilePattern(content.GetName(), params.FilePatterns) || exceedsFileSize(content, params.MaxFileSize) {
				continue
			}
			body, err := t.getFileContent(ctx, owner, repo, content.GetPath())
			if err != nil {
				return fmt.Errorf("failed to get content of file %s: %w", content.GetPath(), err)
			}
			path := content.GetPath()
			name := content.GetName()
			matches = append(matches, GitHubFileContent{FileName: name, Path: path[:len(path)-len(name)], Content: body})
			if params.MaxFiles > 0 && len(matches) >= params.MaxFiles {
				return errMaximumFiles
			}
		}
		return nil
	}
	if err := walk(contents, 0); err != nil && err != errMaximumFiles {
		return nil, err
	}
	return matches, nil
}

var errMaximumFiles = fmt.Errorf("maximum number of files reached")

func (t *GitHubContentTool) getFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	file, _, _, err := t.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", err
	}
	return file.GetContent()
}

func (t *GitHubContentTool) getRepositoryContents(ctx context.Context, owner, repo string) ([]*github.RepositoryContent, error) {
	_, contents, _, err := t.client.Repositories.GetContents(ctx, owner, repo, "", nil)
	if contents == nil {
		contents = []*github.RepositoryContent{}
	}
	return contents, err
}

func extractRepoInfo(repositoryURL string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(repositoryURL, "https://github.com/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format %s", repositoryURL)
	}
	return parts[0], parts[1], nil
}

func matchesSpecificFile(path string, files []string) bool {
	if len(files) == 0 {
		return true
	}
	for _, file := range files {
		if strings.HasSuffix(path, file) {
			return true
		}
	}
	return false
}

func matchesFilePattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasSuffix(name, pattern) {
			return true
		}
	}
	return false
}

func exceedsFileSize(content *github.RepositoryContent, max int64) bool {
	return max > 0 && content.Size != nil && int64(content.GetSize()) > max
}
