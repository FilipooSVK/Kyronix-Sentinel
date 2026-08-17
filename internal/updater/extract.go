package updater

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	maxExtractedFiles = 64

	maxExtractedSize = 512 * 1024 * 1024
)

// ExtractedRelease represents a safely extracted and validated release.
type ExtractedRelease struct {
	Root string

	Manifest Manifest

	ManifestPath string

	SentineldPath string

	SentinelctlPath string
}

// ExtractAndValidateRelease safely extracts a verified update package
// and validates its manifest.
//
// No installation is performed.
func ExtractAndValidateRelease(
	artifact VerifiedArtifact,
	workDir string,
	expectedVersion string,
	expectedOS string,
	expectedArch string,
) (ExtractedRelease, error) {

	if artifact.PackagePath == "" {

		return ExtractedRelease{}, fmt.Errorf(
			"verified package path is empty",
		)
	}

	if err := os.MkdirAll(
		workDir,
		0755,
	); err != nil {

		return ExtractedRelease{}, err
	}

	root, err := os.MkdirTemp(
		workDir,
		".sentinel-release-*",
	)

	if err != nil {
		return ExtractedRelease{}, err
	}

	cleanup := func() {

		_ = os.RemoveAll(
			root,
		)
	}

	if err := extractTarGzSafe(
		artifact.PackagePath,
		root,
	); err != nil {

		cleanup()

		return ExtractedRelease{}, err
	}

	manifestPath := filepath.Join(
		root,
		"manifest.json",
	)

	manifest, err := LoadManifest(
		manifestPath,
	)

	if err != nil {

		cleanup()

		return ExtractedRelease{}, fmt.Errorf(
			"manifest load failed: %w",
			err,
		)
	}

	if err := ValidateManifest(
		manifest,
		root,
		expectedVersion,
		expectedOS,
		expectedArch,
	); err != nil {

		cleanup()

		return ExtractedRelease{}, fmt.Errorf(
			"manifest validation failed: %w",
			err,
		)
	}

	sentineldPath := filepath.Join(
		root,
		"sentineld",
	)

	sentinelctlPath := filepath.Join(
		root,
		"sentinelctl",
	)

	if err := os.Chmod(
		sentineldPath,
		0755,
	); err != nil {

		cleanup()
		return ExtractedRelease{}, err
	}

	if err := os.Chmod(
		sentinelctlPath,
		0755,
	); err != nil {

		cleanup()
		return ExtractedRelease{}, err
	}

	return ExtractedRelease{
		Root: root,

		Manifest: manifest,

		ManifestPath: manifestPath,

		SentineldPath: sentineldPath,

		SentinelctlPath: sentinelctlPath,
	}, nil
}

func extractTarGzSafe(
	archivePath string,
	destination string,
) error {

	file, err := os.Open(
		archivePath,
	)

	if err != nil {
		return err
	}

	defer file.Close()

	gzipReader, err := gzip.NewReader(
		file,
	)

	if err != nil {
		return err
	}

	defer gzipReader.Close()

	tarReader := tar.NewReader(
		gzipReader,
	)

	var totalSize int64

	fileCount := 0

	for {

		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if header.Size < 0 {

			return fmt.Errorf(
				"archive entry has invalid size: %s",
				header.Name,
			)
		}

		target, err := safeArchiveTarget(
			destination,
			header.Name,
		)

		if err != nil {
			return err
		}

		switch header.Typeflag {

		case tar.TypeDir:

			if err := os.MkdirAll(
				target,
				0755,
			); err != nil {

				return err
			}

		case tar.TypeReg,
			tar.TypeRegA:

			fileCount++

			if fileCount > maxExtractedFiles {

				return fmt.Errorf(
					"archive contains too many files",
				)
			}

			totalSize += header.Size

			if totalSize > maxExtractedSize {

				return fmt.Errorf(
					"archive exceeds maximum extracted size",
				)
			}

			if err := os.MkdirAll(
				filepath.Dir(target),
				0755,
			); err != nil {

				return err
			}

			output, err := os.OpenFile(
				target,
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
				0600,
			)

			if err != nil {
				return err
			}

			written, copyErr := io.Copy(
				output,
				tarReader,
			)

			closeErr := output.Close()

			if copyErr != nil {
				return copyErr
			}

			if closeErr != nil {
				return closeErr
			}

			if written != header.Size {

				return fmt.Errorf(
					"archive file size mismatch for %s",
					header.Name,
				)
			}

		default:

			return fmt.Errorf(
				"unsupported archive entry type for %s",
				header.Name,
			)
		}
	}

	return nil
}

func safeArchiveTarget(
	root string,
	name string,
) (string, error) {

	if name == "" {

		return "", fmt.Errorf(
			"archive contains empty path",
		)
	}

	if strings.Contains(
		name,
		"\\",
	) {

		return "", fmt.Errorf(
			"archive path contains backslash: %s",
			name,
		)
	}

	clean := path.Clean(
		name,
	)

	if path.IsAbs(clean) ||
		clean == ".." ||
		strings.HasPrefix(
			clean,
			"../",
		) {

		return "", fmt.Errorf(
			"unsafe archive path: %s",
			name,
		)
	}

	if clean == "." {

		return root, nil
	}

	target := filepath.Join(
		root,
		filepath.FromSlash(clean),
	)

	relative, err := filepath.Rel(
		root,
		target,
	)

	if err != nil {
		return "", err
	}

	if relative == ".." ||
		strings.HasPrefix(
			relative,
			".."+string(filepath.Separator),
		) {

		return "", fmt.Errorf(
			"archive path escapes destination: %s",
			name,
		)
	}

	return target, nil
}
