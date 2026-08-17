package updater

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestReadExpectedSHA256(
	t *testing.T,
) {

	dir := t.TempDir()

	path := filepath.Join(
		dir,
		"package.sha256",
	)

	hash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	content := fmt.Sprintf(
		"%s  sentinel-v0.1.1-linux-arm64.tar.gz\n",
		hash,
	)

	if err := os.WriteFile(
		path,
		[]byte(content),
		0644,
	); err != nil {

		t.Fatal(err)
	}

	result, err := ReadExpectedSHA256(
		path,
		"sentinel-v0.1.1-linux-arm64.tar.gz",
	)

	if err != nil {
		t.Fatal(err)
	}

	if result != hash {

		t.Fatalf(
			"expected %s, got %s",
			hash,
			result,
		)
	}
}

func TestFileSHA256(
	t *testing.T,
) {

	dir := t.TempDir()

	path := filepath.Join(
		dir,
		"test.bin",
	)

	data := []byte(
		"Kyronix Sentinel",
	)

	if err := os.WriteFile(
		path,
		data,
		0644,
	); err != nil {

		t.Fatal(err)
	}

	expected := fmt.Sprintf(
		"%x",
		sha256.Sum256(data),
	)

	actual, err := FileSHA256(
		path,
	)

	if err != nil {
		t.Fatal(err)
	}

	if actual != expected {

		t.Fatalf(
			"expected %s, got %s",
			expected,
			actual,
		)
	}
}

func TestDownloadAndVerify(
	t *testing.T,
) {

	packageData := []byte(
		"Kyronix Sentinel update package",
	)

	hash := fmt.Sprintf(
		"%x",
		sha256.Sum256(packageData),
	)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {

				switch request.URL.Path {

				case "/package.tar.gz":

					writer.WriteHeader(
						http.StatusOK,
					)

					_, _ = writer.Write(
						packageData,
					)

				case "/package.tar.gz.sha256":

					writer.WriteHeader(
						http.StatusOK,
					)

					_, _ = fmt.Fprintf(
						writer,
						"%s  sentinel-v0.1.1-linux-arm64.tar.gz\n",
						hash,
					)

				default:

					http.NotFound(
						writer,
						request,
					)
				}
			},
		),
	)

	defer server.Close()

	assets := ReleaseAssets{
		Package: Asset{
			Name: "sentinel-v0.1.1-linux-arm64.tar.gz",

			BrowserDownloadURL: server.URL +
				"/package.tar.gz",

			Size: int64(
				len(packageData),
			),
		},

		Checksum: Asset{
			Name: "sentinel-v0.1.1-linux-arm64.tar.gz.sha256",

			BrowserDownloadURL: server.URL +
				"/package.tar.gz.sha256",
		},

		OS: "linux",

		Arch: "arm64",
	}

	result, err := DownloadAndVerify(
		context.Background(),
		assets,
		t.TempDir(),
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.SHA256 != hash {

		t.Fatalf(
			"expected hash %s, got %s",
			hash,
			result.SHA256,
		)
	}

	if _, err := os.Stat(
		result.PackagePath,
	); err != nil {

		t.Fatalf(
			"verified package does not exist: %v",
			err,
		)
	}

	if _, err := os.Stat(
		result.ChecksumPath,
	); err != nil {

		t.Fatalf(
			"checksum file does not exist: %v",
			err,
		)
	}
}

func TestDownloadAndVerifyRejectsChecksumMismatch(
	t *testing.T,
) {

	packageData := []byte(
		"tampered update package",
	)

	wrongHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {

				switch request.URL.Path {

				case "/package.tar.gz":

					_, _ = writer.Write(
						packageData,
					)

				case "/package.tar.gz.sha256":

					_, _ = fmt.Fprintf(
						writer,
						"%s  sentinel-v0.1.1-linux-arm64.tar.gz\n",
						wrongHash,
					)

				default:

					http.NotFound(
						writer,
						request,
					)
				}
			},
		),
	)

	defer server.Close()

	dir := t.TempDir()

	assets := ReleaseAssets{
		Package: Asset{
			Name: "sentinel-v0.1.1-linux-arm64.tar.gz",

			BrowserDownloadURL: server.URL +
				"/package.tar.gz",

			Size: int64(
				len(packageData),
			),
		},

		Checksum: Asset{
			Name: "sentinel-v0.1.1-linux-arm64.tar.gz.sha256",

			BrowserDownloadURL: server.URL +
				"/package.tar.gz.sha256",
		},
	}

	_, err := DownloadAndVerify(
		context.Background(),
		assets,
		dir,
	)

	if err == nil {

		t.Fatal(
			"expected checksum mismatch error",
		)
	}

	packagePath := filepath.Join(
		dir,
		assets.Package.Name,
	)

	if _, statErr := os.Stat(
		packagePath,
	); !os.IsNotExist(statErr) {

		t.Fatal(
			"package must be removed after checksum failure",
		)
	}
}

func TestSafeAssetNameRejectsPathTraversal(
	t *testing.T,
) {

	_, err := safeAssetName(
		"../../sentineld",
	)

	if err == nil {

		t.Fatal(
			"expected unsafe asset name to be rejected",
		)
	}
}
