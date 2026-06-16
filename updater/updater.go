package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	GitHubOwner     = "Floren-Team"
	GitHubRepo      = "florenbot"
	VersionFile     = "version.txt"
	CURRENT_VERSION = "v5.3"
	UserAgent       = "FlorenBot-Updater/1.0"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func main() {
	targetPath := flag.String("target", "", "Path to the executable to update")
	updateURL := flag.String("url", "", "URL to download the update from")
	flag.Parse()

	if *targetPath != "" && *updateURL != "" {
		log.Println("Autonomous mode: starting update...")
		if err := runUpdateAutonomous(*targetPath, *updateURL); err != nil {
			log.Fatalf("Update failed: %v", err)
		}
		return
	}

	currentVer := readLocalVersion()
	if currentVer == "0.0.0" {
		currentVer = CURRENT_VERSION
	}

	log.Printf("Current version: %s", currentVer)
	RunUpdaterDaemon()
	select {}
}

func RunUpdaterDaemon() {
	ticker := time.NewTicker(3 * time.Minute)
	go func() {
		log.Println("Background updater started (3-minute interval).")
		for {
			currentVer := readLocalVersion()
			if currentVer == "0.0.0" {
				currentVer = CURRENT_VERSION
			}
			UpdateIfNeeded(GitHubOwner, GitHubRepo, currentVer)
			<-ticker.C
		}
	}()
}

func readLocalVersion() string {
	data, err := os.ReadFile(VersionFile)
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimSpace(string(data))
}

func updateLocalVersion(newVersion string) {
	_ = os.WriteFile(VersionFile, []byte(newVersion), 0644)
}

// Створення HTTP клієнта з User-Agent
func newHttpClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	return newHttpClient().Do(req)
}

func UpdateIfNeeded(owner, repo, currentVersion string) {
	release, err := GetLatestRelease(owner, repo)
	if err != nil {
		log.Printf("Failed to fetch release: %v", err)
		return
	}

	if release.TagName == currentVersion {
		return
	}

	log.Printf("New version found: %s (local: %s). Starting update...", release.TagName, currentVersion)
	
	osName := strings.ToLower(runtime.GOOS)
	archName := strings.ToLower(runtime.GOARCH)

	for _, asset := range release.Assets {
		assetName := strings.ToLower(asset.Name)
		if strings.Contains(assetName, osName) && strings.Contains(assetName, archName) {
			log.Printf("Downloading asset: %s", asset.Name)
			if runUpdateIntegrated(asset.BrowserDownloadURL) {
				updateLocalVersion(release.TagName)
				log.Println("Update successful.")
			}
			return
		}
	}
}

func GetLatestRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	return &release, err
}

func runUpdateIntegrated(url string) bool {
	tmpPath := filepath.Join(os.TempDir(), "app_update_tmp_"+time.Now().Format("20060102150405"))
	
	resp, err := doRequest(url)
	if err != nil {
		log.Printf("Download error: %v", err)
		return false
	}
	defer resp.Body.Close()

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY, 0755) // Права на виконання
	if err != nil {
		return false
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return false
	}

	return replaceBinary(tmpPath) == nil
}

func replaceBinary(newPath string) error {
	exePath, _ := os.Executable()
	oldPath := exePath + ".old"
	
	_ = os.Remove(oldPath)
	_ = os.Rename(exePath, oldPath)
	err := os.Rename(newPath, exePath)
	if err != nil {
		_ = os.Rename(oldPath, exePath)
		return err
	}
	
	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Start()
	os.Exit(0)
	return nil
}

func runUpdateAutonomous(targetPath, url string) error {
	tmpPath := targetPath + ".tmp"
	resp, err := doRequest(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()

	_ = os.Remove(targetPath + ".old")
	_ = os.Rename(targetPath, targetPath+".old")
	return os.Rename(tmpPath, targetPath)
}