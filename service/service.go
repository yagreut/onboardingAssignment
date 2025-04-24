package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/yagreut/onboardingAssignment/models"
)

var githubURLPattern = regexp.MustCompile(`^(https:\/\/|git@)github\.com[/:][\w\-]+\/[\w\-]+(\.git)?$`)
var githubTokenPattern = regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`)
var skipScanExtensions = map[string]bool{
	".pdf":   true,
	".exe":   true,
	".dll":   true,
	".so":    true,
	".a":     true,
	".o":     true,
	".jar":   true,
	".class": true,
	".zip":   true,
	".gz":    true,
	".tar":   true,
	".rar":   true,
	".7z":    true,
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".bmp":   true,
	".tiff":  true,
	".ico":   true,
	".svg":   true,
	".mp3":   true,
	".wav":   true,
	".mp4":   true,
	".avi":   true,
	".mov":   true,
	".wmv":   true,
	".ttf":   true,
	".otf":   true,
	".woff":  true,
	".woff2": true,
}

func ValidateInput(input models.Input) error {
	if input.CloneURL == "" {
		return errors.New("clone URL is required")
	}
	if !githubURLPattern.MatchString(input.CloneURL) {
		return errors.New("clone URL is not a valid GitHub repository URL (HTTPS or SSH)")
	}
	if input.SizeMB <= 0 {
		return errors.New("size threshold must be a positive number")
	}
	return nil
}

func CloneRepo(cloneURL string) (string, error) {
	dir, err := os.MkdirTemp("/tmp", "repo-*")
	if err != nil {
		logrus.WithError(err).Error("Failed to create temporary directory for clone")
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	logrus.Debugf("Created temporary directory for clone: %s", dir)

	logrus.Infof("Attempting to clone repository %s into %s...", cloneURL, dir)

	var stderr bytes.Buffer
	cmd := exec.Command("git", "clone", "--depth=1", cloneURL, dir)
	cmd.Stderr = &stderr

	logrus.Info("Cloning repository...")

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		errorMsg := fmt.Sprintf("git clone command failed for %s: %v\nGit stderr: %s", cloneURL, err, stderr.String())
		logrus.Error(errorMsg)
		return "", errors.New(errorMsg)
	}

	logrus.Infof("Successfully cloned repository %s", cloneURL)
	return dir, nil
}

func ScanRepo(cloneURL string) (string, []models.FileOutput, error) {
	repoDir, err := CloneRepo(cloneURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed during repository cloning phase: %w", err)
	}

	var results []models.FileOutput
	logrus.Infof("Scanning cloned repository structure at: %s", repoDir)

	walkErr := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).Warnf("Error accessing path %q during scan", path)
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" {
				logrus.Debugf("Skipping .git directory: %s", path)
				return filepath.SkipDir
			}
			logrus.Debugf("Entering directory: %s", path)
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			logrus.Debugf("Skipping hidden file: %s", path)
			return nil
		}

		logrus.Debugf("Found file: %s (Size: %d bytes)", path, info.Size())

		rel, err := filepath.Rel(repoDir, path)
		if err != nil {
			logrus.WithError(err).Warnf("Could not get relative path for %q, using full path instead.", path)
			rel = path
		}

		results = append(results, models.FileOutput{
			Name:     rel,
			Size:     info.Size(),
			FullPath: path,
		})
		return nil
	})

	if walkErr != nil {
		logrus.WithError(walkErr).Errorf("Error occurred during directory walk of %s", repoDir)
		return repoDir, results, fmt.Errorf("filesystem scan failed: %w", walkErr)
	}

	logrus.Infof("Filesystem scan complete for %s.", repoDir)
	return repoDir, results, nil
}

func FilterLargeAndSecretFiles(files []models.FileOutput, sizeMB int) ([]models.BigFileOutput, []models.SecretFileOutput) {
	var large []models.BigFileOutput
	var secret []models.SecretFileOutput
	limit := int64(sizeMB) * 1024 * 1024
	const divisor = 1024.0 * 1024.0

	logrus.Infof("Starting analysis of %d files against size limit (%d MB) and secret patterns...", len(files), sizeMB)

	for _, f := range files {
		logrus.Debugf("Analyzing file: %s (Size: %d bytes)", f.Name, f.Size)

		if f.Size > limit {
			convertedSize := float64(f.Size) / divisor
			large = append(large, models.BigFileOutput{
				Name: f.Name,
				Size: convertedSize,
			})
			continue
		}

		if f.Size > 0 {
			ext := strings.ToLower(filepath.Ext(f.Name))

			if skipScanExtensions[ext] {
				logrus.Debugf("Skipping secret scan for file with extension %s: %s", ext, f.Name)
				continue
			}

			logrus.Debugf("Scanning for secrets in: %s", f.Name)
			if line := FindGitHubTokenLineInFile(f.FullPath); line > 0 {
				secret = append(secret, models.SecretFileOutput{
					Name: f.Name,
					Line: line,
				})
			} else {
				logrus.Debugf("No secrets found in: %s", f.Name)
			}
		} else {
			logrus.Debugf("Skipping empty file: %s", f.Name)
		}
	}

	return large, secret
}

func FindGitHubTokenLineInFile(path string) int {
	logrus.Debugf("Opening file for secret scan: %s", path)
	file, err := os.Open(path)
	if err != nil {
		logrus.WithError(err).Warnf("Could not open file %q for secret scanning", path)
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		if githubTokenPattern.MatchString(scanner.Text()) {
			return lineNum
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		logrus.WithError(err).Warnf("Error reading file %q during secret scan", path)
	}

	return 0
}
