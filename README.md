# GitHub Repo Scanner

A command-line tool written in Go that scans a given GitHub repository. It identifies files larger than a specified size and scans smaller files for potential GitHub personal access tokens.

The scanner takes a JSON file as input, specifying the repository's clone URL and a size threshold in megabytes. It outputs a JSON summary listing the large files found and any files potentially containing GitHub tokens, along with their line numbers.

## Key Features

*   Scans public or private (if local Git credentials are configured) GitHub repositories.
*   Identifies files exceeding a user-defined size threshold (in MB).
*   Scans smaller files for potential GitHub personal access tokens (`ghp_...`).
*   Outputs results in a structured JSON format to standard output.
*   Uses shallow clones (`git clone --depth=1`) for efficiency.
*   Automatically cleans up the temporary clone directory upon completion.

## Tech Stack

*   **Language:** Go (Golang)
*   **External Libraries:** `github.com/sirupsen/logrus` (for logging)
*   **External Dependencies:** `git` command-line tool

## Getting Started

### Prerequisites

*   [Go](https://go.dev/doc/install) (Version 1.18 or later recommended)
*   [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/yagreut/onboardingAssignment
    cd onboardingAssignment
    ```
2.  The project uses Go modules. Dependencies are typically handled automatically when building or running.

### Usage

1.  **Create an Input File:**
    Create a JSON file (e.g., `input.json`) with the following structure:

    ```json
    {
      "clone_url": "https://github.com/user/repository.git",
      "size": 10
    }
    ```
    *   `clone_url`: The HTTPS or SSH URL of the GitHub repository to scan.
    *   `size`: The file size threshold in Megabytes (MB). Files strictly larger than this will be reported as "big".

2.  **Run the Scanner:**

    *   **Using `go run`:**
        ```sh
        go run main.go input.json
        ```

    *   **Alternatively, build the executable first:**
        ```sh
        go build -o repo-scanner main.go
        ```
        Then run the compiled binary:
        ```sh
        ./repo-scanner input.json
        ```

## Output Format

The application outputs a JSON object to standard output upon successful completion.

**Example Output:**

```json
{
  "total_big_files": 1,
  "big_files": [
    {
      "name": "assets/large-video.mp4",
      "size_mb": 15.75
    }
  ],
  "total_secret_files": 1,
  "secret_files": [
    {
      "name": "config/dev.settings",
      "line": 23
    }
  ]
}