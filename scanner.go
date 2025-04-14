package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func cloneRepo(url string) (string, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", err
	}

	logrus.WithField("path", tmpDir).Info("Created temp directory")

	// Run `git clone <url> <tmpDir>`
	cmd := exec.Command("git", "clone", url, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logrus.WithField("url", url).Info("Cloning repository...")
	if err := cmd.Run(); err != nil {
		return "", err
	}

	logrus.Info("Repository cloned successfully")
	return tmpDir, nil
}

func walkFiles(root string) ([]os.FileInfo, error) {
	var files []os.FileInfo

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).Warn("Error walking path")
			return nil // Skip error, but continue
		}
		if !info.IsDir() {
			files = append(files, info)
			logrus.WithFields(logrus.Fields{
				"name": info.Name(),
				"size": info.Size(),
			}).Debug("Found file")
		}
		return nil
	})

	return files, err
}

func filterLargeFiles(files []os.FileInfo, sizeLimitMB int) []FileOutput {
	var largeFiles []FileOutput
	sizeLimitBytes := int64(sizeLimitMB) * 1024 * 1024

	for _, file := range files {
		if file.Size() > sizeLimitBytes {
			largeFiles = append(largeFiles, FileOutput{
				Name: file.Name(),
				Size: file.Size(),
			})
		}
	}

	return largeFiles
}
