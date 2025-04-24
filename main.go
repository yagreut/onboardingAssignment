package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yagreut/onboardingAssignment/models"
	"github.com/yagreut/onboardingAssignment/repository"
	"github.com/yagreut/onboardingAssignment/service"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	if len(os.Args) < 2 {
		logrus.Fatal("Usage: go run main.go <input_file.json>")
	}

	inputFile := os.Args[1]

	logrus.Debugf("Reading input configuration from: %s", inputFile)
	input, err := repository.ReadInputFromFile(inputFile)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to read or parse input file: %s", inputFile)
	}

	logrus.Debug("Validating input configuration...")
	if err := service.ValidateInput(input); err != nil {
		logrus.WithError(err).Fatal("Invalid input configuration")
	}
	logrus.Debug("Input validation successful.")

	logrus.Infof("Starting scan for repository: %s (Size Threshold: %d MB)", input.CloneURL, input.SizeMB)

	repoDir, files, err := service.ScanRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to scan repository")
	}

	if repoDir != "" {
		defer func() {
			logrus.Infof("Cleaning up temporary directory: %s", repoDir)
			if err := os.RemoveAll(repoDir); err != nil {
				logrus.WithError(err).Errorf("Failed to remove temporary directory: %s", repoDir)
			} else {
				logrus.Debugf("Successfully removed temporary directory: %s", repoDir)
			}
		}()
	}

	logrus.Infof("Discovered %d files/entries in repository. Starting filtering...", len(files))

	largeFiles, secretFiles := service.FilterLargeAndSecretFiles(files, input.SizeMB)

	logrus.Infof("Filtering complete. Found %d large file(s) and %d file(s) with potential secrets.", len(largeFiles), len(secretFiles))

	result := models.ScanResult{
		TotalBig:    len(largeFiles),
		BigFiles:    largeFiles,
		TotalSecret: len(secretFiles),
		SecretFiles: secretFiles,
	}

	logrus.Debug("Marshalling results to JSON...")
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to marshal output to JSON")
	}

	fmt.Println(string(output))

	logrus.Info("Scan completed successfully.")
}
