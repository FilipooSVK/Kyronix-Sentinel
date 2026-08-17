package updater

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type testArchiveEntry struct {
	Name string

	Data []byte

	Type byte

	Linkname string
}

func createTestReleaseArchive(
	t *testing.T,
	version string,
	goos string,
	goarch string,
	mutateManifest func(*Manifest),
) string {

	t.Helper()

	sentineld := []byte(
		"test sentineld binary",
	)

	sentinelctl := []byte(
		"test sentinelctl binary",
	)

	sentineldHash := fmt.Sprintf(
		"%x",
		sha256.Sum256(sentineld),
	)

	sentinelctlHash := fmt.Sprintf(
		"%x",
		sha256.Sum256(sentinelctl),
	)

	manifest := Manifest{
		Version: version,

		OS: goos,

		Arch: goarch,

		Files: []ManifestFile{
			{
				Name:   "sentineld",
				SHA256: sentineldHash,
			},
			{
				Name:   "sentinelctl",
				SHA256: sentinelctlHash,
			},
		},
	}

	if mutateManifest != nil {
		mutateManifest(
			&manifest,
		)
	}

	manifestData, err := json.Marshal(
		manifest,
	)

	if err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(
		t.TempDir(),
		"release.tar.gz",
	)

	createTarGz(
		t,
		archive,
		[]testArchiveEntry{
			{
				Name: "manifest.json",
				Data: manifestData,
				Type: tar.TypeReg,
			},
			{
				Name: "sentineld",
				Data: sentineld,
				Type: tar.TypeReg,
			},
			{
				Name: "sentinelctl",
				Data: sentinelctl,
				Type: tar.TypeReg,
			},
		},
	)

	return archive
}

func createTarGz(
	t *testing.T,
	archivePath string,
	entries []testArchiveEntry,
) {

	t.Helper()

	file, err := os.Create(
		archivePath,
	)

	if err != nil {
		t.Fatal(err)
	}

	gzipWriter := gzip.NewWriter(
		file,
	)

	tarWriter := tar.NewWriter(
		gzipWriter,
	)

	for _, entry := range entries {

		header := &tar.Header{
			Name: entry.Name,

			Mode: 0644,

			Size: int64(
				len(entry.Data),
			),

			Typeflag: entry.Type,

			Linkname: entry.Linkname,
		}

		if entry.Type == 0 {
			header.Typeflag = tar.TypeReg
		}

		if entry.Type == tar.TypeSymlink {
			header.Size = 0
		}

		if err := tarWriter.WriteHeader(
			header,
		); err != nil {

			t.Fatal(err)
		}

		if header.Typeflag == tar.TypeReg {

			if _, err := tarWriter.Write(
				entry.Data,
			); err != nil {

				t.Fatal(err)
			}
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}

	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractAndValidateRelease(
	t *testing.T,
) {

	archive := createTestReleaseArchive(
		t,
		"v0.1.1",
		"linux",
		"arm64",
		nil,
	)

	result, err := ExtractAndValidateRelease(
		VerifiedArtifact{
			PackagePath: archive,
		},
		t.TempDir(),
		"v0.1.1",
		"linux",
		"arm64",
	)

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(
		result.Root,
	)

	if result.Manifest.Version != "v0.1.1" {

		t.Fatalf(
			"unexpected version: %s",
			result.Manifest.Version,
		)
	}

	for _, binary := range []string{
		result.SentineldPath,
		result.SentinelctlPath,
	} {

		info, err := os.Stat(
			binary,
		)

		if err != nil {
			t.Fatal(err)
		}

		if info.Mode().Perm() != 0755 {

			t.Fatalf(
				"expected executable mode 0755, got %o",
				info.Mode().Perm(),
			)
		}
	}
}

func TestExtractRejectsPathTraversal(
	t *testing.T,
) {

	archive := filepath.Join(
		t.TempDir(),
		"evil.tar.gz",
	)

	createTarGz(
		t,
		archive,
		[]testArchiveEntry{
			{
				Name: "../sentineld",
				Data: []byte("evil"),
				Type: tar.TypeReg,
			},
		},
	)

	_, err := ExtractAndValidateRelease(
		VerifiedArtifact{
			PackagePath: archive,
		},
		t.TempDir(),
		"v0.1.1",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected path traversal archive to be rejected",
		)
	}
}

func TestExtractRejectsSymlink(
	t *testing.T,
) {

	archive := filepath.Join(
		t.TempDir(),
		"evil.tar.gz",
	)

	createTarGz(
		t,
		archive,
		[]testArchiveEntry{
			{
				Name:     "sentineld",
				Type:     tar.TypeSymlink,
				Linkname: "/bin/sh",
			},
		},
	)

	_, err := ExtractAndValidateRelease(
		VerifiedArtifact{
			PackagePath: archive,
		},
		t.TempDir(),
		"v0.1.1",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected symlink archive to be rejected",
		)
	}
}

func TestManifestRejectsWrongPlatform(
	t *testing.T,
) {

	archive := createTestReleaseArchive(
		t,
		"v0.1.1",
		"linux",
		"amd64",
		nil,
	)

	_, err := ExtractAndValidateRelease(
		VerifiedArtifact{
			PackagePath: archive,
		},
		t.TempDir(),
		"v0.1.1",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected architecture mismatch",
		)
	}
}

func TestManifestRejectsWrongVersion(
	t *testing.T,
) {

	archive := createTestReleaseArchive(
		t,
		"v0.2.0",
		"linux",
		"arm64",
		nil,
	)

	_, err := ExtractAndValidateRelease(
		VerifiedArtifact{
			PackagePath: archive,
		},
		t.TempDir(),
		"v0.1.1",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected version mismatch",
		)
	}
}

func TestManifestRejectsBinaryHashMismatch(
	t *testing.T,
) {

	archive := createTestReleaseArchive(
		t,
		"v0.1.1",
		"linux",
		"arm64",
		func(
			manifest *Manifest,
		) {

			manifest.Files[0].SHA256 =
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		},
	)

	_, err := ExtractAndValidateRelease(
		VerifiedArtifact{
			PackagePath: archive,
		},
		t.TempDir(),
		"v0.1.1",
		"linux",
		"arm64",
	)

	if err == nil {

		t.Fatal(
			"expected binary SHA256 mismatch",
		)
	}
}
