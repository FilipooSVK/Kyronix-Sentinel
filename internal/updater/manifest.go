package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ManifestFile describes one file contained in a Sentinel release.
type ManifestFile struct {
	Name string `json:"name"`

	SHA256 string `json:"sha256"`
}

// Manifest describes a Sentinel release package.
type Manifest struct {
	Version string `json:"version"`

	OS string `json:"os"`

	Arch string `json:"arch"`

	Files []ManifestFile `json:"files"`
}

// LoadManifest loads and validates the manifest JSON syntax.
func LoadManifest(
	path string,
) (Manifest, error) {

	data, err := os.ReadFile(
		path,
	)

	if err != nil {
		return Manifest{}, err
	}

	decoder := json.NewDecoder(
		bytes.NewReader(data),
	)

	decoder.DisallowUnknownFields()

	var manifest Manifest

	if err := decoder.Decode(
		&manifest,
	); err != nil {

		return Manifest{}, err
	}

	var extra interface{}

	if err := decoder.Decode(
		&extra,
	); err != io.EOF {

		return Manifest{}, fmt.Errorf(
			"manifest contains trailing data",
		)
	}

	return manifest, nil
}

// ValidateManifest validates release metadata and binary hashes.
func ValidateManifest(
	manifest Manifest,
	root string,
	expectedVersion string,
	expectedOS string,
	expectedArch string,
) error {

	if !IsValidVersion(
		manifest.Version,
	) {

		return fmt.Errorf(
			"invalid manifest version: %s",
			manifest.Version,
		)
	}

	if expectedVersion != "" &&
		CompareVersions(
			manifest.Version,
			expectedVersion,
		) != 0 {

		return fmt.Errorf(
			"manifest version mismatch: expected %s, got %s",
			expectedVersion,
			manifest.Version,
		)
	}

	if manifest.OS != expectedOS {

		return fmt.Errorf(
			"manifest OS mismatch: expected %s, got %s",
			expectedOS,
			manifest.OS,
		)
	}

	if manifest.Arch != expectedArch {

		return fmt.Errorf(
			"manifest architecture mismatch: expected %s, got %s",
			expectedArch,
			manifest.Arch,
		)
	}

	files := make(
		map[string]ManifestFile,
	)

	for _, file := range manifest.Files {

		if file.Name == "" {

			return fmt.Errorf(
				"manifest contains empty file name",
			)
		}

		if filepath.Base(
			file.Name,
		) != file.Name ||
			strings.Contains(
				file.Name,
				"\\",
			) {

			return fmt.Errorf(
				"invalid manifest file name: %s",
				file.Name,
			)
		}

		if _, exists := files[file.Name]; exists {

			return fmt.Errorf(
				"duplicate manifest file: %s",
				file.Name,
			)
		}

		if !validSHA256(
			file.SHA256,
		) {

			return fmt.Errorf(
				"invalid SHA256 for %s",
				file.Name,
			)
		}

		files[file.Name] = file
	}

	for _, required := range []string{
		"sentineld",
		"sentinelctl",
	} {

		file, exists := files[required]

		if !exists {

			return fmt.Errorf(
				"required file missing from manifest: %s",
				required,
			)
		}

		path := filepath.Join(
			root,
			required,
		)

		info, err := os.Lstat(
			path,
		)

		if err != nil {

			return fmt.Errorf(
				"required file missing: %s: %w",
				required,
				err,
			)
		}

		if !info.Mode().IsRegular() {

			return fmt.Errorf(
				"required file is not regular: %s",
				required,
			)
		}

		actualHash, err := FileSHA256(
			path,
		)

		if err != nil {
			return err
		}

		if !strings.EqualFold(
			actualHash,
			file.SHA256,
		) {

			return fmt.Errorf(
				"SHA256 mismatch for %s",
				required,
			)
		}
	}

	return nil
}
