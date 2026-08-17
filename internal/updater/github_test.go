package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubClientLatestRelease(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {

				if request.URL.Path !=
					"/repos/kyronix/sentinel/releases/latest" {

					t.Fatalf(
						"unexpected path: %s",
						request.URL.Path,
					)
				}

				if request.Header.Get(
					"Accept",
				) != "application/vnd.github+json" {

					t.Fatalf(
						"unexpected Accept header: %s",
						request.Header.Get("Accept"),
					)
				}

				if request.Header.Get(
					"X-GitHub-Api-Version",
				) != gitHubAPIVersion {

					t.Fatalf(
						"unexpected API version: %s",
						request.Header.Get(
							"X-GitHub-Api-Version",
						),
					)
				}

				writer.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = writer.Write(
					[]byte(`{
						"tag_name":"v0.1.1",
						"name":"Kyronix Sentinel v0.1.1",
						"draft":false,
						"prerelease":false,
						"assets":[
							{
								"name":"sentinel-linux-arm64.tar.gz",
								"browser_download_url":"https://example.invalid/sentinel.tar.gz",
								"size":12345
							}
						]
					}`),
				)
			},
		),
	)

	defer server.Close()

	client := NewGitHubClient(
		"kyronix",
		"sentinel",
	)

	client.baseURL = server.URL

	release, err := client.LatestRelease(
		context.Background(),
	)

	if err != nil {
		t.Fatal(err)
	}

	if release.TagName != "v0.1.1" {

		t.Fatalf(
			"expected v0.1.1, got %s",
			release.TagName,
		)
	}

	if len(release.Assets) != 1 {

		t.Fatalf(
			"expected 1 asset, got %d",
			len(release.Assets),
		)
	}

	if release.Assets[0].Name !=
		"sentinel-linux-arm64.tar.gz" {

		t.Fatalf(
			"unexpected asset: %s",
			release.Assets[0].Name,
		)
	}
}

func TestGitHubClientCheckUpdateAvailable(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {

				writer.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = writer.Write(
					[]byte(`{
						"tag_name":"v0.1.1",
						"name":"Kyronix Sentinel v0.1.1",
						"draft":false,
						"prerelease":false,
						"assets":[]
					}`),
				)
			},
		),
	)

	defer server.Close()

	client := NewGitHubClient(
		"kyronix",
		"sentinel",
	)

	client.baseURL = server.URL

	result, err := client.Check(
		context.Background(),
		"0.1.0",
	)

	if err != nil {
		t.Fatal(err)
	}

	if !result.UpdateAvailable {

		t.Fatal(
			"expected update to be available",
		)
	}

	if result.CurrentVersion != "v0.1.0" {

		t.Fatalf(
			"unexpected current version: %s",
			result.CurrentVersion,
		)
	}

	if result.LatestVersion != "v0.1.1" {

		t.Fatalf(
			"unexpected latest version: %s",
			result.LatestVersion,
		)
	}
}

func TestGitHubClientCheckUpToDate(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {

				writer.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = writer.Write(
					[]byte(`{
						"tag_name":"v0.1.0",
						"name":"Kyronix Sentinel v0.1.0",
						"draft":false,
						"prerelease":false,
						"assets":[]
					}`),
				)
			},
		),
	)

	defer server.Close()

	client := NewGitHubClient(
		"kyronix",
		"sentinel",
	)

	client.baseURL = server.URL

	result, err := client.Check(
		context.Background(),
		"0.1.0",
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.UpdateAvailable {

		t.Fatal(
			"same version must be up to date",
		)
	}
}

func TestGitHubClientHTTPError(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {

				http.Error(
					writer,
					"not found",
					http.StatusNotFound,
				)
			},
		),
	)

	defer server.Close()

	client := NewGitHubClient(
		"kyronix",
		"sentinel",
	)

	client.baseURL = server.URL

	_, err := client.LatestRelease(
		context.Background(),
	)

	if err == nil {

		t.Fatal(
			"expected GitHub HTTP error",
		)
	}
}

func TestGitHubClientRejectsInvalidReleaseVersion(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {

				writer.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = writer.Write(
					[]byte(`{
						"tag_name":"latest",
						"name":"invalid",
						"assets":[]
					}`),
				)
			},
		),
	)

	defer server.Close()

	client := NewGitHubClient(
		"kyronix",
		"sentinel",
	)

	client.baseURL = server.URL

	_, err := client.Check(
		context.Background(),
		"0.1.0",
	)

	if err == nil {

		t.Fatal(
			"expected invalid release version error",
		)
	}
}
