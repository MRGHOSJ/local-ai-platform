package ollama

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type DownloadProgress struct {
	Downloaded int64  `json:"downloaded"`
	Total     int64  `json:"total"`
	Percent  float64 `json:"percent"`
}

var onDownloadProgress func(DownloadProgress)

func SetDownloadProgressCallback(cb func(DownloadProgress)) {
	onDownloadProgress = cb
}

var pinnedVersion = "v0.22.1"

var downloadURLs = map[string]string{
	"darwin-amd64":  "https://github.com/ollama/ollama/releases/download/" + pinnedVersion + "/ollama-darwin",
	"darwin-arm64":  "https://github.com/ollama/ollama/releases/download/" + pinnedVersion + "/ollama-darwin",
	"linux-amd64":   "https://github.com/ollama/ollama/releases/download/" + pinnedVersion + "/ollama-linux-amd64",
	"linux-arm64":   "https://github.com/ollama/ollama/releases/download/" + pinnedVersion + "/ollama-linux-arm64",
	"windows-amd64": "https://github.com/ollama/ollama/releases/download/" + pinnedVersion + "/ollama-windows-amd64",
}

func EnsureOllama(onProgress func(DownloadProgress)) (string, error) {
	if path, err := checkSystemOllama(); err == nil {
		return path, nil
	}

	if path, err := checkCachedOllama(); err == nil {
		return path, nil
	}

	if err := downloadOllama(onProgress); err != nil {
		return "", err
	}

	return checkCachedOllama()
}

func checkSystemOllama() (string, error) {
	binaryName := "ollama"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	path, err := exec.LookPath(binaryName)
	if err != nil {
		return "", fmt.Errorf("not in PATH: %w", err)
	}

	version, err := getOllamaVersion(path)
	if err != nil {
		return "", err
	}

	if !isVersionCompatible(version) {
		return "", fmt.Errorf("system Ollama %s too old, need >= 0.20.0", version)
	}

	return path, nil
}

func checkCachedOllama() (string, error) {
	cacheDir := getOllamaDir()
	binaryName := "ollama"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	cached := filepath.Join(cacheDir, binaryName)

	if _, err := os.Stat(cached); err != nil {
		return "", err
	}

	return cached, nil
}

func getOllamaVersion(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(string(out))
	parts := strings.Split(output, " ")
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[1], "v"), nil
	}

	if strings.Contains(output, "ollama version") {
		re := regexp.MustCompile(`ollama version (\d+\.\d+\.\d+)`)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not parse version from: %s", output)
}

func isVersionCompatible(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return false
	}

	major := 0
	minor := 0
	fmt.Sscanf(parts[0], "%d", &major)
	fmt.Sscanf(parts[1], "%d", &minor)

	if major > 0 {
		return true
	}

	return minor >= 20
}

func downloadOllama(onProgress func(DownloadProgress)) error {
	key := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	urlBase, ok := downloadURLs[key]
	if !ok {
		return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	assetURL := urlBase
	if runtime.GOOS == "windows" {
		assetURL += ".exe"
	}

	ext := ".exe"
	if runtime.GOOS != "windows" {
		ext = ""
		assetURL += ".tgz"
	}
	downloadURL := assetURL

	cacheDir := getOllamaDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	archivePath := filepath.Join(cacheDir, "download"+ext)
	if err := downloadFile(downloadURL, archivePath, onProgress); err != nil {
		os.Remove(archivePath)
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

func downloadFile(url, dest string, onProgress func(DownloadProgress)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	var downloaded int64
	total := resp.ContentLength
	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			if onProgress != nil && total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				onProgress(DownloadProgress{
					Downloaded: downloaded,
					Total:     total,
					Percent:  percent,
				})
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func getOllamaDir() string {
	return filepath.Join(getAppDir(), "ollama")
}

func GetPinnedVersion() string {
	return strings.TrimPrefix(pinnedVersion, "v")
}