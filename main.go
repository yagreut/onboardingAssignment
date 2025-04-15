package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Input struct {
	CloneURL string `json:"clone_url"`
	SizeMB   int    `json:"size"`
}

type FileOutput struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ScanResult struct {
	Total int          `json:"total"`
	Files []FileOutput `json:"files"`
}

// Ask user for file name, then clone URL and size
func collectAndSaveInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the name of the file to create (e.g. input.json): ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)

	fmt.Print("Enter the GitHub clone URL: ")
	cloneURL, _ := reader.ReadString('\n')
	cloneURL = strings.TrimSpace(cloneURL)

	fmt.Print("Enter the size limit in MB: ")
	var size int
	fmt.Scanln(&size)

	input := Input{
		CloneURL: cloneURL,
		SizeMB:   size,
	}

	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	logrus.Info("✅ Input saved to ", filename)
	return filename, nil
}

// Read the input JSON file into a struct
func readInputFromFile(filename string) (Input, error) {
	var input Input
	data, err := os.ReadFile(filename)
	if err != nil {
		return input, err
	}

	err = json.Unmarshal(data, &input)
	return input, err
}

func main() {
	// Step 1: Ask user for input + save to a file
	inputFile, err := collectAndSaveInput()
	if err != nil {
		logrus.WithError(err).Fatal("❌ Failed to collect or save input")
	}

	// Step 2: Read input from that file
	input, err := readInputFromFile(inputFile)
	if err != nil {
		logrus.WithError(err).Fatal("❌ Failed to read input file")
	}

	// Step 3: Clone and scan
	files, err := ScanRepo(input.CloneURL)
	if err != nil {
		logrus.WithError(err).Fatal("❌ Failed to scan repo")
	}

	largeFiles := filterLargeFiles(files, input.SizeMB)

	result := ScanResult{
		Total: len(largeFiles),
		Files: largeFiles,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.WithError(err).Fatal("❌ Failed to marshal result")
	}

	logrus.Info("📊 Scan Result:\n" + string(output))
}
