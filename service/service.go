package service

import (
	"bufio"
	"errors"
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
		return errors.New("clone URL must be a valid GitHub link")
	}

	if input.SizeMB <= 0 {
		return errors.New("size must be a positive number")
	}

	return nil
}

func CloneRepo(cloneURL string) (string, error) {
	dir, err := os.MkdirTemp("/tmp", "repo-*")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", "--depth=1", cloneURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}

func ScanRepo(cloneURL string) (string, []models.FileOutput, error) {
	repoDir, err := CloneRepo(cloneURL)
	if err != nil {
		return "", nil, err
	}

	var results []models.FileOutput

	err = filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).Warnf("Error accessing path %q", path)
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		rel, err := filepath.Rel(repoDir, path)
		if err != nil {
			logrus.WithError(err).Warn("Could not get relative path, using full path instead.")
			rel = path
		}

		results = append(results, models.FileOutput{
			Name:     rel,
			Size:     info.Size(),
			FullPath: path,
		})
		return nil
	})

	return repoDir, results, err
}

func FilterLargeAndSecretFiles(files []models.FileOutput, sizeMB int) ([]models.BigFileOutput, []models.SecretFileOutput) {
	var large []models.BigFileOutput
	var secret []models.SecretFileOutput
	limit := int64(sizeMB) * 1024 * 1024
	const divisor = 1024.0 * 1024.0

	for _, f := range files {
		if f.Size > limit {
			convertedSize := float64(f.Size) / divisor
			large = append(large, models.BigFileOutput{
				Name: f.Name,
				Size: convertedSize,
			})
			continue
		} else if f.Size > 0 {
			ext := strings.ToLower(filepath.Ext(f.Name))

			if skipScanExtensions[ext] {
				logrus.Debugf("Skipping secret scan for extension %s: %s", ext, f.Name)
				continue
			}
			if line := FindGitHubTokenLineInFile(f.FullPath); line > 0 {
				secret = append(secret, models.SecretFileOutput{
					Name: f.Name,
					Line: line,
				})
			}
		}
	}

	return large, secret
}

func FindGitHubTokenLineInFile(path string) int {
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
