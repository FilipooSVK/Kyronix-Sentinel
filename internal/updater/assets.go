package updater

import (
	"fmt"
	"runtime"
	"strings"
)

// ReleaseAssets contains the update package and its checksum asset.
type ReleaseAssets struct {
	Package Asset

	Checksum Asset

	OS string

	Arch string
}

// CurrentPlatform returns the platform used by the running Sentinel.
func CurrentPlatform() (
	string,
	string,
) {

	return runtime.GOOS,
		runtime.GOARCH
}

// SelectReleaseAssets selects the update package and checksum
// matching the requested operating system and architecture.
func SelectReleaseAssets(
	release Release,
	goos string,
	goarch string,
) (ReleaseAssets, error) {

	platform := strings.ToLower(
		goos + "-" + goarch,
	)

	var packageAsset Asset

	var checksumAsset Asset

	for _, asset := range release.Assets {

		name := strings.ToLower(
			asset.Name,
		)

		if !strings.Contains(
			name,
			platform,
		) {
			continue
		}

		switch {

		case strings.HasSuffix(
			name,
			".tar.gz.sha256",
		):

			checksumAsset = asset

		case strings.HasSuffix(
			name,
			".tar.gz",
		):

			packageAsset = asset
		}
	}

	if packageAsset.Name == "" {

		return ReleaseAssets{}, fmt.Errorf(
			"no update package found for platform %s",
			platform,
		)
	}

	if checksumAsset.Name == "" {

		return ReleaseAssets{}, fmt.Errorf(
			"no checksum asset found for package %s",
			packageAsset.Name,
		)
	}

	return ReleaseAssets{
		Package: packageAsset,

		Checksum: checksumAsset,

		OS: goos,

		Arch: goarch,
	}, nil
}

// SelectCurrentPlatformAssets selects release assets for the
// operating system and architecture running Sentinel.
func SelectCurrentPlatformAssets(
	release Release,
) (ReleaseAssets, error) {

	goos, goarch := CurrentPlatform()

	return SelectReleaseAssets(
		release,
		goos,
		goarch,
	)
}
