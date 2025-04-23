// Package models defines the data structures used for input, output, and scan results
// throughout the repository scanner application.
package models

// Input represents the data provided by the user, specifying the repository
// to scan and the size threshold for identifying large files.
type Input struct {
	// CloneURL is the URL of the GitHub repository to be scanned.
	// It should be a valid GitHub clone URL (HTTPS or SSH).
	CloneURL string `json:"clone_url"`
	// SizeMB is the size threshold in megabytes. Files larger than this
	// size will be reported as "big files".
	SizeMB int `json:"size"`
}

// FileOutput represents information about a single file discovered during the
// initial repository scan, before filtering based on size or content.
type FileOutput struct {
	// Name is the relative path of the file within the repository.
	Name string `json:"name"`
	// Size is the size of the file in bytes.
	Size int64 `json:"size_in_MB"`
	// FullPath is the absolute path to the file on the local filesystem
	// where the repository was cloned. This is used internally for scanning.
	FullPath string
}

// BigFileOutput represents a file identified as exceeding the specified size threshold.
type BigFileOutput struct {
	// Name is the relative path of the large file within the repository.
	Name string `json:"name"`
	// Size is the size of the file converted to megabytes.
	Size float64 `json:"size_mb"`
}

// SecretFileOutput represents a file identified as containing a potential GitHub token.
type SecretFileOutput struct {
	// Name is the relative path of the file containing the secret within the repository.
	Name string `json:"name"`
	// Line is the line number (1-based) where the first potential GitHub token was found.
	Line int `json:"line"`
}

// ScanResult aggregates the findings of the repository scan, including lists
// of large files and files containing potential secrets.
type ScanResult struct {
	// TotalBig is the total count of files found exceeding the size threshold.
	TotalBig int `json:"total_big_files"`
	// BigFiles is a slice containing details of each file exceeding the size threshold.
	BigFiles []BigFileOutput `json:"big_files"`
	// TotalSecret is the total count of files found containing potential GitHub tokens.
	TotalSecret int `json:"total_secret_files"`
	// SecretFiles is a slice containing details of each file with a potential GitHub token.
	SecretFiles []SecretFileOutput `json:"secret_files"`
}
