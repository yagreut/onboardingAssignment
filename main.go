// Package main is the entry point for the repository scanner application.
// It reads configuration from a file specified as a command-line argument,
// performs the repository scan, and prints the results as JSON to standard output.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	// Import necessary packages from the current project.
	"github.com/yagreut/onboardingAssignment/models"
	"github.com/yagreut/onboardingAssignment/repository"
	"github.com/yagreut/onboardingAssignment/service"

	// Import external libraries.
	"github.com/sirupsen/logrus" // Used for structured logging.
)

// main function orchestrates the repository scanning process.
func main() {
	// Check for the required command-line argument (input file path).
	if len(os.Args) < 2 {
		logrus.Fatal("Usage: go run main.go <input_file.json>")
	}

	// Get the input file path from the command-line arguments.
	inputFile := os.Args[1]

	// Read the input configuration (URL, size) from the specified JSON file.
	input, err := repository.ReadInputFromFile(inputFile)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to read input file: %s", inputFile)
	}

	// Validate the input data read from the file.
	if err := service.ValidateInput(input); err != nil {
		logrus.WithError(err).Fatal("Invalid input configuration")
	}

	// Log the start of the scanning process.
	logrus.Infof("Scanning repository: %s", input.CloneURL)

	// Perform the repository scan (clone and list files).
	// repoDir is the path to the temporary directory where the repo was cloned.
	// files contains the list of all discovered files.
	repoDir, files, err := service.ScanRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to scan repository")
	}

	// Ensure the temporary repository directory is cleaned up after the scan.
	// This deferred function runs just before main exits.
	if repoDir != "" {
		defer func() {
			logrus.Infof("Cleaning up temporary directory: %s", repoDir)
			if err := os.RemoveAll(repoDir); err != nil {
				// Log an error if cleanup fails, but don't terminate the program.
				logrus.WithError(err).Errorf("Failed to remove temporary directory: %s", repoDir)
			}
		}()
	}

	// Log the completion of the file discovery phase.
	logrus.Infof("Found %d files in repository. Filtering...", len(files))

	// Filter the discovered files into large files and files containing secrets.
	largeFiles, secretFiles := service.FilterLargeAndSecretFiles(files, input.SizeMB)

	// Prepare the final result structure.
	result := models.ScanResult{
		TotalBig:    len(largeFiles),
		BigFiles:    largeFiles,
		TotalSecret: len(secretFiles),
		SecretFiles: secretFiles,
	}

	// Marshal the result into a human-readable JSON format (indented).
	output, err := json.MarshalIndent(result, "", "  ") // Use two spaces for indentation.
	if err != nil {
		logrus.WithError(err).Fatal("Failed to marshal output to JSON")
	}

	// Print the final JSON output to standard output.
	fmt.Println(string(output))

	// Log successful completion.
	logrus.Info("Scan completed successfully.")
}
