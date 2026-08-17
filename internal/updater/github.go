package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	defaultGitHubAPIBaseURL = "https://api.github.com"
	gitHubAPIVersion        = "2026-03-10"
	defaultHTTPTimeout      = 15 * time.Second
)

// GitHubClient retrieves Sentinel release information from GitHub.
type GitHubClient struct {
	owner string

	repository string

	baseURL string

	token string

	httpClient *http.Client
}

// NewGitHubClient creates a GitHub release client.
func NewGitHubClient(
	owner string,
	repository string,
) *GitHubClient {

	return &GitHubClient{
		owner: owner,

		repository: repository,

		baseURL: defaultGitHubAPIBaseURL,

		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

// SetToken configures optional GitHub authentication.
//
// Public repositories do not require a token, but authenticated
// requests may be used later when higher API limits are desirable.
func (c *GitHubClient) SetToken(
	token string,
) {

	c.token = strings.TrimSpace(
		token,
	)
}

// LatestRelease returns the latest published GitHub release.
func (c *GitHubClient) LatestRelease(
	ctx context.Context,
) (Release, error) {

	url := fmt.Sprintf(
		"%s/repos/%s/%s/releases/latest",
		strings.TrimRight(
			c.baseURL,
			"/",
		),
		c.owner,
		c.repository,
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)

	if err != nil {
		return Release{}, err
	}

	request.Header.Set(
		"Accept",
		"application/vnd.github+json",
	)

	request.Header.Set(
		"X-GitHub-Api-Version",
		gitHubAPIVersion,
	)

	request.Header.Set(
		"User-Agent",
		"Kyronix-Sentinel-Updater",
	)

	if c.token != "" {

		request.Header.Set(
			"Authorization",
			"Bearer "+c.token,
		)
	}

	response, err := c.httpClient.Do(
		request,
	)

	if err != nil {
		return Release{}, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		return Release{}, fmt.Errorf(
			"github release request failed: status %d",
			response.StatusCode,
		)
	}

	var release Release

	if err := json.NewDecoder(
		response.Body,
	).Decode(
		&release,
	); err != nil {

		return Release{}, err
	}

	if release.TagName == "" {

		return Release{}, fmt.Errorf(
			"github release response contains empty tag",
		)
	}

	return release, nil
}

// Check checks whether GitHub contains a newer Sentinel release.
func (c *GitHubClient) Check(
	ctx context.Context,
	currentVersion string,
) (CheckResult, error) {

	if !IsValidVersion(
		currentVersion,
	) {

		return CheckResult{}, fmt.Errorf(
			"invalid current version: %s",
			currentVersion,
		)
	}

	release, err := c.LatestRelease(
		ctx,
	)

	if err != nil {
		return CheckResult{}, err
	}

	if !IsValidVersion(
		release.TagName,
	) {

		return CheckResult{}, fmt.Errorf(
			"invalid release version: %s",
			release.TagName,
		)
	}

	return CheckResult{
		CurrentVersion: NormalizeVersion(
			currentVersion,
		),

		LatestVersion: NormalizeVersion(
			release.TagName,
		),

		UpdateAvailable: IsNewerVersion(
			currentVersion,
			release.TagName,
		),

		Release: release,
	}, nil
}
