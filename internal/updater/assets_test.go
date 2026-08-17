package updater

import "testing"

func TestSelectReleaseAssetsARM64(
	t *testing.T,
) {

	release := Release{
		TagName: "v0.1.1",

		Assets: []Asset{
			{
				Name: "sentinel-v0.1.1-linux-amd64.tar.gz",
			},
			{
				Name: "sentinel-v0.1.1-linux-amd64.tar.gz.sha256",
			},
			{
				Name:               "sentinel-v0.1.1-linux-arm64.tar.gz",
				BrowserDownloadURL: "https://example.invalid/arm64.tar.gz",
			},
			{
				Name:               "sentinel-v0.1.1-linux-arm64.tar.gz.sha256",
				BrowserDownloadURL: "https://example.invalid/arm64.tar.gz.sha256",
			},
		},
	}

	assets, err := SelectReleaseAssets(
		release,
		"linux",
		"arm64",
	)

	if err != nil {
		t.Fatal(err)
	}

	if assets.Package.Name !=
		"sentinel-v0.1.1-linux-arm64.tar.gz" {

		t.Fatalf(
			"unexpected package asset: %s",
			assets.Package.Name,
		)
	}

	if assets.Checksum.Name !=
		"sentinel-v0.1.1-linux-arm64.tar.gz.sha256" {

		t.Fatalf(
			"unexpected checksum asset: %s",
			assets.Checksum.Name,
		)
	}

	if assets.OS != "linux" {
		t.Fatalf(
			"expected linux, got %s",
			assets.OS,
		)
	}

	if assets.Arch != "arm64" {
		t.Fatalf(
			"expected arm64, got %s",
			assets.Arch,
		)
	}
}

func TestSelectReleaseAssetsAMD64(
	t *testing.T,
) {

	release := Release{
		Assets: []Asset{
			{
				Name: "sentinel-v0.1.1-linux-amd64.tar.gz",
			},
			{
				Name: "sentinel-v0.1.1-linux-amd64.tar.gz.sha256",
			},
		},
	}

	assets, err := SelectReleaseAssets(
		release,
		"linux",
		"amd64",
	)

	if err != nil {
		t.Fatal(err)
	}

	if assets.Package.Name !=
		"sentinel-v0.1.1-linux-amd64.tar.gz" {

		t.Fatalf(
			"unexpected package asset: %s",
			assets.Package.Name,
		)
	}
}

func TestSelectReleaseAssetsMissingPackage(
	t *testing.T,
) {

	release := Release{
		Assets: []Asset{
			{
				Name: "sentinel-v0.1.1-linux-arm64.tar.gz.sha256",
			},
		},
	}

	_, err := SelectReleaseAssets(
		release,
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected missing package error",
		)
	}
}

func TestSelectReleaseAssetsMissingChecksum(
	t *testing.T,
) {

	release := Release{
		Assets: []Asset{
			{
				Name: "sentinel-v0.1.1-linux-arm64.tar.gz",
			},
		},
	}

	_, err := SelectReleaseAssets(
		release,
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected missing checksum error",
		)
	}
}

func TestSelectReleaseAssetsIgnoresOtherPlatforms(
	t *testing.T,
) {

	release := Release{
		Assets: []Asset{
			{
				Name: "sentinel-v0.1.1-linux-amd64.tar.gz",
			},
			{
				Name: "sentinel-v0.1.1-linux-amd64.tar.gz.sha256",
			},
		},
	}

	_, err := SelectReleaseAssets(
		release,
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected platform selection error",
		)
	}
}
