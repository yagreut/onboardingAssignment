// Package service contains the core business logic for validating input,
// cloning repositories, scanning files, and filtering results based on size
// and content (secrets).
package service

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/yagreut/onboardingAssignment/models"
)

// githubURLPattern defines the regex for validating GitHub repository URLs (HTTPS or SSH).
var githubURLPattern = regexp.MustCompile(`^(https:\/\/|git@)github\.com[/:][\w\-]+\/[\w\-]+(\.git)?$`)

// githubTokenPattern defines the regex for finding potential GitHub personal access tokens.
// Note: This is a basic pattern and might have false positives/negatives.
var githubTokenPattern = regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`)

var skipScanExtensions = map[string]bool{
	// Binary / Compiled
	".pdf":   true,
	".exe":   true,
	".dll":   true,
	".so":    true,
	".a":     true,
	".o":     true,
	".jar":   true,
	".class": true,
	// Archives
	".zip": true,
	".gz":  true,
	".tar": true,
	".rar": true,
	".7z":  true,
	// Images
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".bmp":  true,
	".tiff": true,
	".ico":  true,
	// Vector Images (often problematic for line scanning)
	".svg": true,
	// Audio/Video
	".mp3": true,
	".wav": true,
	".mp4": true,
	".avi": true,
	".mov": true,
	".wmv": true,
	// Fonts
	".ttf":   true,
	".otf":   true,
	".woff":  true,
	".woff2": true,
}

// ValidateInput checks if the provided Input data is valid.
// It ensures the CloneURL is present and matches the GitHub URL pattern,
// and that SizeMB is a positive integer.
// It returns nil if the input is valid, otherwise returns an error describing the issue.
func ValidateInput(input models.Input) error {
	if input.CloneURL == "" {
		return errors.New("clone URL is required")
	}

	if !githubURLPattern.MatchString(input.CloneURL) {
		return errors.New("clone URL must be a valid GitHub link")
	}

	if input.SizeMB <= 0 {
		return errors.New("size must be a positive number")
	}

	// Input is valid.
	return nil
}

// CloneRepo clones the specified Git repository into a temporary directory.
// It uses `git clone --depth=1` for a shallow clone to save time and bandwidth.
// It returns the path to the temporary directory where the repo was cloned,
// or an error if the cloning process fails.
// The caller is responsible for cleaning up the temporary directory.
func CloneRepo(cloneURL string) (string, error) {
	// Create a temporary directory to store the cloned repository.
	// The directory name will have the prefix "repo-".
	dir, err := os.MkdirTemp("/tmp", "repo-*")
	if err != nil {
		return "", err // Return error if temp dir creation fails.
	}

	// Prepare the git clone command.
	// --depth=1 creates a shallow clone, downloading only the latest commit.
	cmd := exec.Command("git", "clone", "--depth=1", cloneURL, dir)
	// Redirect command's stdout and stderr to the application's streams for visibility.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command.
	if err := cmd.Run(); err != nil {
		// If cloning fails, attempt to remove the partially created directory.
		_ = os.RemoveAll(dir) // Ignore error during cleanup on failure.
		return "", err        // Return the original error from cmd.Run().
	}

	// Return the path to the directory containing the cloned repository.
	return dir, nil
}

// ScanRepo clones the repository specified by cloneURL and then walks the
// filesystem to find all regular files, excluding hidden files/directories
// and the .git directory itself.
// It returns the path to the temporary clone directory (for later cleanup),
// a slice of FileOutput structs representing the found files, and any error
// encountered during cloning or filesystem traversal.
func ScanRepo(cloneURL string) (string, []models.FileOutput, error) {
	// Clone the repository first.
	repoDir, err := CloneRepo(cloneURL)
	if err != nil {
		// If cloning fails, return the error immediately. repoDir will be "".
		return "", nil, err
	}

	var results []models.FileOutput

	// Walk the directory tree starting from the repository root.
	err = filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		// Handle potential errors during walk.
		if err != nil {
			logrus.WithError(err).Warnf("Error accessing path %q", path)
			return err // Propagate the error to stop the walk if needed.
		}
		// Skip directories.
		if info.IsDir() {
			// Specifically skip the .git directory to avoid scanning repository metadata.
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			// Skip other directories silently.
			return nil
		}
		// Skip hidden files (files starting with '.').
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Calculate the relative path within the repository.
		rel, err := filepath.Rel(repoDir, path)
		if err != nil {
			// Log a warning if relative path calculation fails, but use the full path as fallback.
			logrus.WithError(err).Warn("Could not get relative path, using full path instead.")
			rel = path // Fallback to full path.
		}

		// Append file information to the results slice.
		results = append(results, models.FileOutput{
			Name:     rel,         // Relative path within the repo.
			Size:     info.Size(), // Size in bytes.
			FullPath: path,        // Absolute path on the local system.
		})
		return nil // Continue walking.
	})

	// Return the clone directory path, the list of found files, and any error from filepath.Walk.
	return repoDir, results, err
}

// FilterLargeAndSecretFiles processes a list of files found in the repository.
// It categorizes each file based on its size relative to `sizeMB` and whether
// it contains a potential GitHub token pattern.
// Files larger than `sizeMB` (converted to bytes) are added to the `BigFileOutput` list.
// Files smaller than or equal to the limit but not empty are scanned for GitHub tokens;
// if a token is found, the file is added to the `SecretFileOutput` list.
// It returns two slices: one for large files and one for files containing secrets.
func FilterLargeAndSecretFiles(files []models.FileOutput, sizeMB int) ([]models.BigFileOutput, []models.SecretFileOutput) {
	var large []models.BigFileOutput
	var secret []models.SecretFileOutput
	// Calculate the size limit in bytes.
	limit := int64(sizeMB) * 1024 * 1024
	// Define the divisor to convert bytes to megabytes.
	const divisor = 1024.0 * 1024.0

	for _, f := range files {
		// Check if the file size exceeds the limit.
		if f.Size > limit {
			// Convert size to megabytes for reporting.
			convertedSize := float64(f.Size) / divisor
			large = append(large, models.BigFileOutput{
				Name: f.Name,
				Size: convertedSize,
			})
			continue // Skip further checks for large files.
		} else if f.Size > 0 { // Only scan non-empty files that are not large.
			// Scan the file content for a GitHub token pattern.
			// FindGitHubTokenLineInFile returns the line number (1-based) or 0.
			// Get the file extension (lowercase)
			ext := strings.ToLower(filepath.Ext(f.Name))

			// Check if the extension is in our skip list
			if skipScanExtensions[ext] {
				logrus.Debugf("Skipping secret scan for extension %s: %s", ext, f.Name)
				continue // Skip to the next file
			}
			if line := FindGitHubTokenLineInFile(f.FullPath); line > 0 {
				secret = append(secret, models.SecretFileOutput{
					Name: f.Name,
					Line: line,
				})
			}
		}
		// Files that are empty or within the size limit and contain no token are ignored.
	}

	return large, secret
}

// FindGitHubTokenLineInFile scans the file at the given path line by line
// for the first occurrence of the GitHub token pattern (githubTokenPattern).
// It returns the 1-based line number where the pattern is first found.
// If the pattern is not found, or if the file cannot be opened or read,
// it returns 0 and logs a warning in case of errors.
func FindGitHubTokenLineInFile(path string) int {
	// Attempt to open the file for reading.
	file, err := os.Open(path)
	if err != nil {
		// Log a warning if the file cannot be opened.
		logrus.WithError(err).Warnf("Could not open file %q for secret scanning", path)
		return 0 // Return 0 as token cannot be found.
	}
	// Ensure the file is closed when the function returns.
	defer file.Close()

	// Create a scanner to read the file line by line.
	scanner := bufio.NewScanner(file)

	lineNum := 1 // Initialize line number counter (1-based).
	// Iterate through each line of the file.
	for scanner.Scan() {
		// Check if the current line matches the GitHub token pattern.
		if githubTokenPattern.MatchString(scanner.Text()) {
			return lineNum // Return the current line number if a match is found.
		}
		lineNum++ // Increment line number for the next iteration.
	}

	// Check for errors encountered during scanning (e.g., read errors).
	if err := scanner.Err(); err != nil {
		logrus.WithError(err).Warnf("Error reading file %q during secret scan", path)
	}

	// Return 0 if the loop completes without finding the token pattern.
	return 0
}
