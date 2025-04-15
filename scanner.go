package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

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

func ScanRepo(cloneURL string) ([]FileOutput, error) {
	repoDir, err := CloneRepo(cloneURL)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(repoDir)

	// Step 1: Collect all file paths
	var filePaths []string

	err = filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		filePaths = append(filePaths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Step 2: Prepare progress bar + concurrency
	bar := progressbar.Default(int64(len(filePaths)))
	fileChan := make(chan FileOutput)
	var wg sync.WaitGroup

	for _, path := range filePaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			info, err := os.Stat(p)
			if err == nil {
				relPath := strings.TrimPrefix(p, repoDir+"/")
				fileChan <- FileOutput{Name: relPath, Size: info.Size()}
			}
			bar.Add(1)
		}(path)
	}

	go func() {
		wg.Wait()
		close(fileChan)
	}()

	var files []FileOutput
	for file := range fileChan {
		files = append(files, file)
	}

	return files, nil
}

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
