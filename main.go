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

	// ✅ Validate the input
	if err := service.ValidateInput(input); err != nil {
		logrus.WithError(err).Fatal("Invalid input")
	}

	files, err := service.ScanRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to scan repository")
	}

	largeFiles := service.FilterLargeFiles(files, input.SizeMB)

	for i := range largeFiles {
		largeFiles[i].Size = largeFiles[i].Size / (1024 * 1024) // convert bytes to MB
	}
	

	result := models.ScanResult{
		Total: len(largeFiles),
		Files: largeFiles,
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}
