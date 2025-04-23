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
		logrus.WithError(err).Fatal("Failed to read input")
	}

	// Validate the input
	if err := service.ValidateInput(input); err != nil {
		logrus.WithError(err).Fatal("Invalid input")
	}

	repoDir, files, err := service.ScanRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to scan repository")
	}

	if repoDir != "" {
		defer func() {
			if err := os.RemoveAll(repoDir); err != nil {
				logrus.WithError(err).Error("Failed to remove temporary directory")
			}
		}()
	}

	largeFiles, secretFiles := service.FilterLargeAndSecretFiles(files, input.SizeMB)

	result := models.ScanResult{
		TotalBig:    len(largeFiles),
		BigFiles:    largeFiles,
		TotalSecret: len(secretFiles),
		SecretFiles: secretFiles,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to marshal output")
	}
	fmt.Println(string(output))
}
