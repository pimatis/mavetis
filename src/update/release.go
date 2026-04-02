package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Pimatis/mavetis/src/model"
)

const ownerRepo = "pimatis/mavetis"

const apiBase = "https://api.github.com"

const apiLimit = 1 << 20

type Updater struct {
	client       *http.Client
	apiBase      string
	downloadBase string
}

type releasePayload struct {
	TagName string `json:"tag_name"`
}

func New() *Updater {
	client := &http.Client{Timeout: 60 * time.Second}
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects during update")
		}
		if !trustedURL(request.URL.String()) {
			return fmt.Errorf("refusing untrusted update redirect: %s", request.URL.String())
		}
		return nil
	}
	return &Updater{
		client:       client,
		apiBase:      apiBase,
		downloadBase: model.Repository + "/releases/download",
	}
}

func (updater *Updater) Latest(ctx context.Context) (string, error) {
	endpoint := updater.apiBase + "/repos/" + ownerRepo + "/releases/latest"
	body, err := updater.get(ctx, endpoint, apiLimit)
	if err != nil {
		return "", err
	}
	payload := releasePayload{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode latest release: %w", err)
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("decode latest release: missing tag_name")
	}
	return payload.TagName, nil
}

func (updater *Updater) DownloadBinary(ctx context.Context, tag string, dir string) (string, error) {
	return updater.downloadBinary(ctx, tag, dir, runtimePlatform())
}

func (updater *Updater) downloadBinary(ctx context.Context, tag string, dir string, platform string) (string, error) {
	archiveName, binaryName, err := asset(platform)
	if err != nil {
		return "", err
	}
	archivePath := filepath.Join(dir, archiveName)
	checksumPath := archivePath + ".sha256"
	if err := updater.download(ctx, updater.assetURL(tag, archiveName), archivePath); err != nil {
		return "", err
	}
	if err := updater.download(ctx, updater.assetURL(tag, archiveName+".sha256"), checksumPath); err != nil {
		return "", err
	}
	if err := verifyChecksum(archivePath, checksumPath); err != nil {
		return "", err
	}
	binaryPath := filepath.Join(dir, binaryName)
	if err := extractBinary(archivePath, binaryName, binaryPath); err != nil {
		return "", err
	}
	return binaryPath, nil
}

func (updater *Updater) assetURL(tag string, name string) string {
	clean := strings.TrimSpace(tag)
	return updater.downloadBase + "/" + clean + "/" + name
}

func (updater *Updater) download(ctx context.Context, source string, destination string) error {
	if !trustedURL(source) {
		return fmt.Errorf("refusing untrusted update endpoint: %s", source)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", model.Name+"/"+model.Version)
	response, err := updater.client.Do(req)
	if err != nil {
		return fmt.Errorf("request update endpoint: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("update request failed: %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create download file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, response.Body); err != nil {
		return fmt.Errorf("write download file: %w", err)
	}
	return nil
}

func (updater *Updater) get(ctx context.Context, endpoint string, limit int64) ([]byte, error) {
	if !trustedURL(endpoint) {
		return nil, fmt.Errorf("refusing untrusted update endpoint: %s", endpoint)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", model.Name+"/"+model.Version)
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := updater.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request update endpoint: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("update request failed: %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	reader := io.Reader(response.Body)
	if limit > 0 {
		reader = io.LimitReader(response.Body, limit)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read update response: %w", err)
	}
	return body, nil
}

func trustedURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme == "http" && localHost(host) {
		return true
	}
	if parsed.Scheme != "https" {
		return false
	}
	allowed := []string{
		"api.github.com",
		"github.com",
		"release-assets.githubusercontent.com",
		"objects.githubusercontent.com",
	}
	for _, item := range allowed {
		if host == item {
			return true
		}
	}
	return false
}

func localHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	return false
}

func IsNewer(current string, latest string) (bool, error) {
	left, err := parseVersion(current)
	if err != nil {
		return false, err
	}
	right, err := parseVersion(latest)
	if err != nil {
		return false, err
	}
	for index := 0; index < 3; index++ {
		if right[index] > left[index] {
			return true, nil
		}
		if right[index] < left[index] {
			return false, nil
		}
	}
	return false, nil
}

func parseVersion(value string) ([3]int, error) {
	clean := strings.TrimSpace(strings.TrimPrefix(value, "v"))
	cut := strings.IndexAny(clean, "+-")
	if cut >= 0 {
		clean = clean[:cut]
	}
	parts := strings.Split(clean, ".")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid version: %s", value)
	}
	result := [3]int{}
	for index, item := range parts {
		number, err := strconv.Atoi(item)
		if err != nil {
			return [3]int{}, fmt.Errorf("invalid version: %s", value)
		}
		result[index] = number
	}
	return result, nil
}

func verifyChecksum(archivePath string, checksumPath string) error {
	expectedBody, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("read checksum: %w", err)
	}
	expected := strings.Fields(string(expectedBody))
	if len(expected) == 0 {
		return fmt.Errorf("read checksum: missing digest")
	}
	archive, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}
	defer archive.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, archive); err != nil {
		return fmt.Errorf("hash archive: %w", err)
	}
	actual := fmt.Sprintf("%x", hash.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(actual), []byte(strings.ToLower(expected[0]))) == 1 {
		return nil
	}
	return fmt.Errorf("checksum verification failed")
}

func extractBinary(archivePath string, binaryName string, destination string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, binaryName, destination)
	}
	if strings.HasSuffix(archivePath, ".tar.gz") {
		return extractTarGz(archivePath, binaryName, destination)
	}
	return fmt.Errorf("unsupported archive format: %s", archivePath)
}

func extractZip(archivePath string, binaryName string, destination string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if filepath.Base(file.Name) != binaryName {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return fmt.Errorf("open archive file: %w", err)
		}
		defer reader.Close()
		return writeExtracted(destination, reader)
	}
	return fmt.Errorf("binary not found in archive")
}

func extractTarGz(archivePath string, binaryName string, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	zipper, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer zipper.Close()
	archive := tar.NewReader(zipper)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}
		if filepath.Base(header.Name) != binaryName {
			continue
		}
		return writeExtracted(destination, archive)
	}
	return fmt.Errorf("binary not found in archive")
}

func writeExtracted(destination string, reader io.Reader) error {
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("create binary: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}
	return nil
}
