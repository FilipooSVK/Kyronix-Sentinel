package updater

// Asset represents one downloadable GitHub release asset.
type Asset struct {
	Name string `json:"name"`

	BrowserDownloadURL string `json:"browser_download_url"`

	Size int64 `json:"size"`
}

// Release represents GitHub release metadata required by Sentinel.
type Release struct {
	TagName string `json:"tag_name"`

	Name string `json:"name"`

	Draft bool `json:"draft"`

	Prerelease bool `json:"prerelease"`

	Assets []Asset `json:"assets"`
}

// CheckResult represents the result of an update availability check.
type CheckResult struct {
	CurrentVersion string

	LatestVersion string

	UpdateAvailable bool

	Release Release
}
