package version

import "testing"

func TestCurrentVersion(t *testing.T) {

	info := Current()

	if info.Version == "" {
		t.Fatal(
			"version should not be empty",
		)
	}

	if info.Commit == "" {
		t.Fatal(
			"commit should not be empty",
		)
	}
}
