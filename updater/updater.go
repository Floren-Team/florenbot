package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	GitHubOwner    = "Floren-Team"
	GitHubRepo     = "florenbot"
	VersionFile    = "version.txt"
	UserAgent      = "FlorenBot-Updater/2.0"
	LATEST_VERSION = "v5.6"
)

func main() {
	log.Printf("[INFO] Initializing. OS: %s, Arch: %s", runtime.GOOS, runtime.GOARCH)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("[INFO] Shutting down gracefully...")
		os.Exit(0)
	}()

	for {
		log.Println("[INFO] Checking for updates...")
		checkAndUpdate()
		log.Println("[INFO] Waiting 3 minutes before next check...")
		time.Sleep(3 * time.Minute)
	}
}

func checkAndUpdate() bool {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ERROR] Network error: %v", err)
		return false
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	json.NewDecoder(resp.Body).Decode(&release)

	// Читаємо поточну версію
	var currentVer string
	localVerData, err := os.ReadFile(VersionFile)

	if err != nil {
		log.Printf("[INFO] Version file not found, assuming %s", LATEST_VERSION)
		currentVer = LATEST_VERSION
	} else {
		currentVer = strings.TrimSpace(string(localVerData))
	}

	log.Printf("[DEBUG] Local version: '%s', GitHub latest: '%s'", currentVer, release.TagName)

	// ПЕРЕВІРКА: якщо версії ідентичні - виходимо
	if release.TagName != "" && release.TagName == currentVer {
		log.Printf("[INFO] Version %s is already up to date. Skipping.", release.TagName)
		return false
	}

	log.Printf("[INFO] Update required: %s -> %s", currentVer, release.TagName)
	osTag, archTag := strings.ToLower(runtime.GOOS), runtime.GOARCH

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osTag) && strings.Contains(name, archTag) {
			log.Printf("[INFO] Downloading asset: %s", name)
			if performUpdate(asset.BrowserDownloadURL, name) {
				// Записуємо нову версію в файл тільки після успішного оновлення
				_ = os.WriteFile(VersionFile, []byte(release.TagName), 0644)
				log.Printf("[SUCCESS] Updated to %s", release.TagName)
				return true
			}
		}
	}
	return false
}

func performUpdate(url, filename string) bool {
	tmpPath := filepath.Join(os.TempDir(), filename)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ERROR] Download failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	out, _ := os.Create(tmpPath)
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		log.Printf("[ERROR] Write failed: %v", err)
		return false
	}

	log.Printf("[DEBUG] Extraction starting...")
	currDir, _ := os.Getwd()

	if strings.HasSuffix(filename, ".zip") {
		err = unzip(tmpPath, currDir)
	} else {
		err = untar(tmpPath, currDir)
	}

	os.Remove(tmpPath)
	if err != nil {
		log.Printf("[ERROR] Extraction error: %v", err)
		return false
	}

	return true
}

func untar(path, dest string) error {
	f, _ := os.Open(path)
	defer f.Close()
	gzr, _ := gzip.NewReader(f)
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if header.FileInfo().IsDir() {
			continue
		}

		target := filepath.Join(dest, header.Name)
		log.Printf("[DEBUG] Extracting: %s", header.Name)

		outFile, _ := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		io.Copy(outFile, tr)
		outFile.Close()
	}
	return nil
}

func unzip(path, dest string) error {
	r, _ := zip.OpenReader(path)
	defer r.Close()
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		target := filepath.Join(dest, f.Name)
		log.Printf("[DEBUG] Extracting: %s", f.Name)

		outFile, _ := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		rc, _ := f.Open()
		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}
