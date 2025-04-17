package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yagreut/onboardingAssignment/models"
	"github.com/yagreut/onboardingAssignment/repository"
	"github.com/yagreut/onboardingAssignment/service"
	"ogithub.com/yagreut/onboardingAssignment/utils"

	"github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 2 {
		logrus.Fatal("Usage: go run main.go <input_file.json>")
	}

	inputFile := os.Args[1]

	input, err := utils.ReadInputFromFile(inputFile)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read input")
	}

	files, err := repository.ScanRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to scan repository")
	}

	largeFiles := service.FilterLargeFiles(files, input.SizeMB)

	result := models.ScanResult{
		Total: len(largeFiles),
		Files: largeFiles,
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}
