package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	updateDownloadTimeout = 60 * time.Second

	maxPackageDownloadSize = 512 * 1024 * 1024

	maxChecksumDownloadSize = 64 * 1024
)

// VerifiedArtifact represents a downloaded and checksum-verified
// Sentinel update package.
type VerifiedArtifact struct {
	PackagePath string

	ChecksumPath string

	SHA256 string
}

// DownloadAndVerify downloads the release checksum and package,
// verifies SHA256 integrity and returns the verified artifact.
//
// No installation is performed by this function.
func DownloadAndVerify(
	ctx context.Context,
	assets ReleaseAssets,
	destination string,
) (VerifiedArtifact, error) {

	if destination == "" {
		return VerifiedArtifact{}, fmt.Errorf(
			"update destination is empty",
		)
	}

	if err := os.MkdirAll(
		destination,
		0755,
	); err != nil {

		return VerifiedArtifact{}, err
	}

	packageName, err := safeAssetName(
		assets.Package.Name,
	)

	if err != nil {
		return VerifiedArtifact{}, err
	}

	checksumName, err := safeAssetName(
		assets.Checksum.Name,
	)

	if err != nil {
		return VerifiedArtifact{}, err
	}

	packagePath := filepath.Join(
		destination,
		packageName,
	)

	checksumPath := filepath.Join(
		destination,
		checksumName,
	)

	client := &http.Client{
		Timeout: updateDownloadTimeout,
	}

	// Download checksum first.
	if err := downloadAsset(
		ctx,
		client,
		assets.Checksum,
		checksumPath,
		maxChecksumDownloadSize,
	); err != nil {

		return VerifiedArtifact{}, fmt.Errorf(
			"checksum download failed: %w",
			err,
		)
	}

	expectedHash, err := ReadExpectedSHA256(
		checksumPath,
		packageName,
	)

	if err != nil {

		_ = os.Remove(
			checksumPath,
		)

		return VerifiedArtifact{}, fmt.Errorf(
			"invalid checksum file: %w",
			err,
		)
	}

	if assets.Package.Size >
		maxPackageDownloadSize {

		_ = os.Remove(
			checksumPath,
		)

		return VerifiedArtifact{}, fmt.Errorf(
			"update package exceeds maximum allowed size",
		)
	}

	if err := downloadAsset(
		ctx,
		client,
		assets.Package,
		packagePath,
		maxPackageDownloadSize,
	); err != nil {

		_ = os.Remove(
			checksumPath,
		)

		return VerifiedArtifact{}, fmt.Errorf(
			"package download failed: %w",
			err,
		)
	}

	actualHash, err := FileSHA256(
		packagePath,
	)

	if err != nil {

		_ = os.Remove(
			packagePath,
		)

		_ = os.Remove(
			checksumPath,
		)

		return VerifiedArtifact{}, err
	}

	if !strings.EqualFold(
		expectedHash,
		actualHash,
	) {

		_ = os.Remove(
			packagePath,
		)

		_ = os.Remove(
			checksumPath,
		)

		return VerifiedArtifact{}, fmt.Errorf(
			"SHA256 mismatch: expected %s, got %s",
			expectedHash,
			actualHash,
		)
	}

	return VerifiedArtifact{
		PackagePath: packagePath,

		ChecksumPath: checksumPath,

		SHA256: actualHash,
	}, nil
}

func downloadAsset(
	ctx context.Context,
	client *http.Client,
	asset Asset,
	destination string,
	maxSize int64,
) error {

	if asset.BrowserDownloadURL == "" {

		return fmt.Errorf(
			"asset %s has no download URL",
			asset.Name,
		)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		asset.BrowserDownloadURL,
		nil,
	)

	if err != nil {
		return err
	}

	request.Header.Set(
		"User-Agent",
		"Kyronix-Sentinel-Updater",
	)

	response, err := client.Do(
		request,
	)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		return fmt.Errorf(
			"download returned status %d",
			response.StatusCode,
		)
	}

	if response.ContentLength > maxSize {

		return fmt.Errorf(
			"asset exceeds maximum allowed size",
		)
	}

	dir := filepath.Dir(
		destination,
	)

	tempFile, err := os.CreateTemp(
		dir,
		".sentinel-download-*",
	)

	if err != nil {
		return err
	}

	tempPath := tempFile.Name()

	cleanup := func() {

		_ = tempFile.Close()

		_ = os.Remove(
			tempPath,
		)
	}

	limitedReader := io.LimitReader(
		response.Body,
		maxSize+1,
	)

	written, err := io.Copy(
		tempFile,
		limitedReader,
	)

	if err != nil {

		cleanup()

		return err
	}

	if written > maxSize {

		cleanup()

		return fmt.Errorf(
			"asset exceeds maximum allowed size",
		)
	}

	if asset.Size > 0 &&
		written != asset.Size {

		cleanup()

		return fmt.Errorf(
			"asset size mismatch: expected %d bytes, got %d",
			asset.Size,
			written,
		)
	}

	if err := tempFile.Sync(); err != nil {

		cleanup()

		return err
	}

	if err := tempFile.Close(); err != nil {

		_ = os.Remove(
			tempPath,
		)

		return err
	}

	if err := os.Chmod(
		tempPath,
		0640,
	); err != nil {

		_ = os.Remove(
			tempPath,
		)

		return err
	}

	if err := os.Rename(
		tempPath,
		destination,
	); err != nil {

		_ = os.Remove(
			tempPath,
		)

		return err
	}

	return nil
}

// ReadExpectedSHA256 reads a sha256sum-compatible checksum file.
//
// Supported examples:
//
// abcdef...  sentinel-linux-arm64.tar.gz
// abcdef... *sentinel-linux-arm64.tar.gz
// abcdef...
func ReadExpectedSHA256(
	path string,
	packageName string,
) (string, error) {

	file, err := os.Open(
		path,
	)

	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(
		file,
	)

	for scanner.Scan() {

		line := strings.TrimSpace(
			scanner.Text(),
		)

		if line == "" {
			continue
		}

		fields := strings.Fields(
			line,
		)

		if len(fields) == 0 {
			continue
		}

		hash := strings.ToLower(
			fields[0],
		)

		if !validSHA256(
			hash,
		) {

			continue
		}

		// A checksum file containing only the hash is accepted.
		if len(fields) == 1 {
			return hash, nil
		}

		name := strings.TrimPrefix(
			fields[len(fields)-1],
			"*",
		)

		if filepath.Base(name) ==
			filepath.Base(packageName) {

			return hash, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf(
		"SHA256 checksum for %s not found",
		packageName,
	)
}

// FileSHA256 calculates SHA256 for a local file.
func FileSHA256(
	path string,
) (string, error) {

	file, err := os.Open(
		path,
	)

	if err != nil {
		return "", err
	}

	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(
		hash,
		file,
	); err != nil {

		return "", err
	}

	return hex.EncodeToString(
		hash.Sum(nil),
	), nil
}

func validSHA256(
	value string,
) bool {

	if len(value) != 64 {
		return false
	}

	_, err := hex.DecodeString(
		value,
	)

	return err == nil
}

func safeAssetName(
	name string,
) (string, error) {

	if name == "" {
		return "", fmt.Errorf(
			"asset name is empty",
		)
	}

	base := filepath.Base(
		name,
	)

	if base != name ||
		base == "." ||
		base == ".." {

		return "", fmt.Errorf(
			"invalid asset name: %s",
			name,
		)
	}

	return base, nil
}
