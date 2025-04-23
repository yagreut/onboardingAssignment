// Package repository handles operations related to data retrieval, such as
// reading input configurations from files.
package repository

import (
	"encoding/json"
	"os"

	"github.com/yagreut/onboardingAssignment/models"
)

// ReadInputFromFile reads the scan configuration (repository URL and size threshold)
// from a JSON file specified by the path.
// It returns the parsed Input struct or an error if the file cannot be read
// or the JSON cannot be unmarshalled.
func ReadInputFromFile(path string) (models.Input, error) {
	var input models.Input

	// Read the entire file content.
	data, err := os.ReadFile(path)
	if err != nil {
		return input, err // Return error if file reading fails.
	}

	// Unmarshal the JSON data into the Input struct.
	err = json.Unmarshal(data, &input)
	// Return the parsed input and any unmarshalling error.
	return input, err
}
