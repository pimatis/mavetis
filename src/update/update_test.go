package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"io"
	"strings"
	"testing"
)

func TestRunAlreadyUpToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/repos/pimatis/mavetis/releases/latest" {
			_, _ = writer.Write([]byte(`{"tag_name":"v0.1.4"}`))
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	previous := newUpdater
	newUpdater = func() *Updater {
		return &Updater{
			client:       server.Client(),
			apiBase:      server.URL,
			downloadBase: server.URL + "/releases/download",
		}
	}
	defer func() { newUpdater = previous }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Run(Spec{Check: false})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	out, _ := io.ReadAll(r)
	expected := "already at its glorious latest version"
	if !strings.Contains(string(out), expected) {
		t.Fatalf("expected output to contain %q, got %q", expected, string(out))
	}
}

func TestIsNewer(t *testing.T) {
	newer, err := IsNewer("0.1.0", "v0.2.0")
	if err != nil {
		t.Fatalf("compare versions: %v", err)
	}
	if !newer {
		t.Fatal("expected newer version")
	}
	newer, err = IsNewer("0.2.0", "v0.2.0")
	if err != nil {
		t.Fatalf("compare versions: %v", err)
	}
	if newer {
		t.Fatal("expected equal version to stay false")
	}
}

func TestLatestAndDownloadBinary(t *testing.T) {
	archive := makeTarGz(t, "mavetis", []byte("binary"))
	digest := sha256.Sum256(archive)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/repos/pimatis/mavetis/releases/latest" {
			_, _ = writer.Write([]byte(`{"tag_name":"v0.2.0"}`))
			return
		}
		if request.URL.Path == "/releases/download/v0.2.0/mavetis_linux_amd64.tar.gz" {
			_, _ = writer.Write(archive)
			return
		}
		if request.URL.Path == "/releases/download/v0.2.0/mavetis_linux_amd64.tar.gz.sha256" {
			_, _ = writer.Write([]byte(fmt.Sprintf("%x  mavetis_linux_amd64.tar.gz\n", digest)))
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	updater := &Updater{
		client:       server.Client(),
		apiBase:      server.URL,
		downloadBase: server.URL + "/releases/download",
	}
	tag, err := updater.Latest(context.Background())
	if err != nil {
		t.Fatalf("latest release: %v", err)
	}
	if tag != "v0.2.0" {
		t.Fatalf("unexpected tag: %s", tag)
	}
	dir := t.TempDir()
	binaryPath, err := updater.downloadBinary(context.Background(), tag, dir, "linux/amd64")
	if err != nil {
		t.Fatalf("download binary: %v", err)
	}
	content, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if string(content) != "binary" {
		t.Fatalf("unexpected binary content: %q", string(content))
	}
}

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "mavetis_windows_amd64.zip")
	body := bytes.NewBuffer(nil)
	archive := zip.NewWriter(body)
	writer, err := archive.Create("mavetis.exe")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := writer.Write([]byte("win-binary")); err != nil {
		t.Fatalf("write zip body: %v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := os.WriteFile(archivePath, body.Bytes(), 0o600); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	destination := filepath.Join(dir, "mavetis.exe")
	if err := extractBinary(archivePath, "mavetis.exe", destination); err != nil {
		t.Fatalf("extract zip: %v", err)
	}
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if string(content) != "win-binary" {
		t.Fatalf("unexpected destination content: %q", string(content))
	}
}

func makeTarGz(t *testing.T, name string, body []byte) []byte {
	t.Helper()
	buffer := bytes.NewBuffer(nil)
	zipper := gzip.NewWriter(buffer)
	archive := tar.NewWriter(zipper)
	header := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}
	if err := archive.WriteHeader(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := archive.Write(body); err != nil {
		t.Fatalf("write body: %v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := zipper.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buffer.Bytes()
}
