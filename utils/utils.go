package utils

import (
	"encoding/json"
	"os"

	"github.com/yagreut/onboardingAssignment/models"
)

func ReadInputFromFile(path string) (models.Input, error) {
	var input models.Input

	data, err := os.ReadFile(path)
	if err != nil {
		return input, err
	}

	err = json.Unmarshal(data, &input)
	return input, err
}
