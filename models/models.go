package models

type Input struct {
	CloneURL string `json:"clone_url"`
	SizeMB   int    `json:"size"`
}

type FileOutput struct {
	Name     string `json:"name"`
	Size     int64  `json:"size_in_MB"`
	FullPath string
}

type BigFileOutput struct {
	Name string  `json:"name"`
	Size float64 `json:"size_mb"`
}

type SecretFileOutput struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

type ScanResult struct {
	TotalBig    int                `json:"total_big_files"`
	BigFiles    []BigFileOutput    `json:"big_files"`
	TotalSecret int                `json:"total_secret_files"`
	SecretFiles []SecretFileOutput `json:"secret_files"`
}
