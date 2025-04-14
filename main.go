package main

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

type RepoInput struct {
	CloneURL string `json:"clone_url"`
	SizeMB   int    `json:"size"` // in MB
}

type FileOutput struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ScanResult struct {
	Total int          `json:"total"`
	Files []FileOutput `json:"files"`
}

func main() {
	// Open the input file
	file, err := os.Open("input.json")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open input.json")
	}
	defer file.Close()

	// Decode the JSON input
	var input RepoInput
	if err := json.NewDecoder(file).Decode(&input); err != nil {
		logrus.WithError(err).Fatal("Failed to parse JSON input")
	}

	// Log the parsed data
	logrus.WithFields(logrus.Fields{
		"clone_url": input.CloneURL,
		"size_MB":   input.SizeMB,
	}).Info("Parsed input JSON successfully")

	repoPath, err := cloneRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to clone repository")
	}

	files, err := walkFiles(repoPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to walk through files")
	}

	logrus.WithField("file_count", len(files)).Info("Finished scanning files")

	largeFiles := filterLargeFiles(files, input.SizeMB)

	result := ScanResult{
		Total: len(largeFiles),
		Files: largeFiles,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to marshal result")
	}

	logrus.Info("Scan Result:\n" + string(output))

}
