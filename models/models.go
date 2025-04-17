package models

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
