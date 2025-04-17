package service

import (
	"github.com/yagreut/onboardingAssignment/models"
)

func FilterLargeFiles(files []models.FileOutput, sizeMB int) []models.FileOutput {
	var large []models.FileOutput
	limit := int64(sizeMB) * 1024 * 1024

	for _, f := range files {
		if f.Size > limit {
			large = append(large, f)
		}
	}

	return large
}
