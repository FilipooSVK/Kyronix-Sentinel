package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeManifestTestFile(
	t *testing.T,
	root string,
	name string,
	content string,
) string {

	t.Helper()

	path := filepath.Join(
		root,
		name,
	)

	if err := os.WriteFile(
		path,
		[]byte(content),
		0644,
	); err != nil {

		t.Fatal(err)
	}

	hash, err := FileSHA256(
		path,
	)

	if err != nil {
		t.Fatal(err)
	}

	return hash
}

func validWorkerManifest(
	t *testing.T,
	root string,
) Manifest {

	t.Helper()

	sentineldSHA := writeManifestTestFile(
		t,
		root,
		"sentineld",
		"sentineld",
	)

	sentinelctlSHA := writeManifestTestFile(
		t,
		root,
		"sentinelctl",
		"sentinelctl",
	)

	workerSHA := writeManifestTestFile(
		t,
		root,
		"sentinel-update.service",
		"worker unit",
	)

	return Manifest{
		Version: "v0.1.4",
		OS:      "linux",
		Arch:    "arm64",

		Files: []ManifestFile{
			{
				Name:   "sentineld",
				SHA256: sentineldSHA,
			},
			{
				Name:   "sentinelctl",
				SHA256: sentinelctlSHA,
			},
			{
				Name:   "sentinel-update.service",
				SHA256: workerSHA,
			},
		},
	}
}

func TestValidateManifestRequiresWorkerUnit(
	t *testing.T,
) {

	root := t.TempDir()

	manifest := validWorkerManifest(
		t,
		root,
	)

	manifest.Files =
		manifest.Files[:2]

	err := ValidateManifest(
		manifest,
		root,
		"v0.1.4",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected missing worker unit error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"required file missing from manifest: sentinel-update.service",
	) {

		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestValidateManifestVerifiesWorkerUnitSHA256(
	t *testing.T,
) {

	root := t.TempDir()

	manifest := validWorkerManifest(
		t,
		root,
	)

	if err := os.WriteFile(
		filepath.Join(
			root,
			"sentinel-update.service",
		),
		[]byte(
			"tampered worker unit",
		),
		0644,
	); err != nil {

		t.Fatal(err)
	}

	err := ValidateManifest(
		manifest,
		root,
		"v0.1.4",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected worker unit SHA256 mismatch",
		)
	}

	if !strings.Contains(
		err.Error(),
		"SHA256 mismatch for sentinel-update.service",
	) {

		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestValidateManifestVerifiesAllDeclaredFiles(
	t *testing.T,
) {

	root := t.TempDir()

	manifest := validWorkerManifest(
		t,
		root,
	)

	extraSHA := writeManifestTestFile(
		t,
		root,
		"release-metadata",
		"metadata",
	)

	manifest.Files = append(
		manifest.Files,
		ManifestFile{
			Name:   "release-metadata",
			SHA256: extraSHA,
		},
	)

	if err := os.WriteFile(
		filepath.Join(
			root,
			"release-metadata",
		),
		[]byte(
			"tampered metadata",
		),
		0644,
	); err != nil {

		t.Fatal(err)
	}

	err := ValidateManifest(
		manifest,
		root,
		"v0.1.4",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected declared file SHA256 mismatch",
		)
	}

	if !strings.Contains(
		err.Error(),
		"SHA256 mismatch for release-metadata",
	) {

		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestValidateManifestAcceptsWorkerRelease(
	t *testing.T,
) {

	root := t.TempDir()

	manifest := validWorkerManifest(
		t,
		root,
	)

	if err := ValidateManifest(
		manifest,
		root,
		"v0.1.4",
		"linux",
		"arm64",
	); err != nil {

		t.Fatalf(
			"valid worker release rejected: %v",
			err,
		)
	}
}
