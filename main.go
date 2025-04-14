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
}
