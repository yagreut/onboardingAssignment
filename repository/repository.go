package repository

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yagreut/onboardingAssignment/models"
)

func CloneRepo(cloneURL string) (string, error) {
	dir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", "--depth=1", cloneURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return dir, nil
}

func ScanRepo(cloneURL string) ([]models.FileOutput, error) {
	repoDir, err := CloneRepo(cloneURL)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(repoDir)

	var results []models.FileOutput

	err = filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(path, repoDir+"/")
		results = append(results, models.FileOutput{
			Name: rel,
			Size: info.Size(),
		})
		return nil
	})

	return results, err
}
