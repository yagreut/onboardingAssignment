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
	if len(os.Args) < 2 {
		logrus.Fatal("Usage: go run main.go <input_file.json>")
	}

	inputFile := os.Args[1]

	input, err := repository.ReadInputFromFile(inputFile)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to read input file: %s", inputFile)
	}

	if err := service.ValidateInput(input); err != nil {
		logrus.WithError(err).Fatal("Invalid input configuration")
	}

	logrus.Infof("Scanning repository: %s", input.CloneURL)

	repoDir, files, err := service.ScanRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to scan repository")
	}

	if repoDir != "" {
		defer func() {
			logrus.Infof("Cleaning up temporary directory: %s", repoDir)
			if err := os.RemoveAll(repoDir); err != nil {
				logrus.WithError(err).Errorf("Failed to remove temporary directory: %s", repoDir)
			}
		}()
	}

	logrus.Infof("Found %d files in repository. Filtering...", len(files))

	largeFiles, secretFiles := service.FilterLargeAndSecretFiles(files, input.SizeMB)

	result := models.ScanResult{
		TotalBig:    len(largeFiles),
		BigFiles:    largeFiles,
		TotalSecret: len(secretFiles),
		SecretFiles: secretFiles,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to marshal output to JSON")
	}

	fmt.Println(string(output))

	logrus.Info("Scan completed successfully.")
}
