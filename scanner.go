package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Clone the repo into a temporary directory
func CloneRepo(cloneURL string) (string, error) {
	dir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", "--depth=1", cloneURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logrus.Infof("⏳ Cloning repository into %s", dir)
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	logrus.Info("✅ Clone completed")
	return dir, nil
}

// Walk the repo and gather all files with their sizes
func ScanRepo(cloneURL string) ([]FileOutput, error) {
	repoDir, err := CloneRepo(cloneURL)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(repoDir)

	var files []FileOutput

	err = filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, repoDir+"/")
		files = append(files, FileOutput{
			Name: relPath,
			Size: info.Size(),
		})
		return nil
	})

	return files, err
}

// Filter out files smaller than the size limit
func filterLargeFiles(files []FileOutput, sizeMB int) []FileOutput {
	var largeFiles []FileOutput
	limit := int64(sizeMB) * 1024 * 1024

	for _, file := range files {
		if file.Size > limit {
			largeFiles = append(largeFiles, file)
		}
	}

	return largeFiles
}
