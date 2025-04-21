package models

type Input struct {
	CloneURL string `json:"clone_url"`
	SizeMB   int    `json:"size"`
}

type BigFileOutput struct {
	Name string `json:"name"`
	Size int64  `json:"size in MB"`
}

type SecretFileOutput struct {
	Name string `json:"name"`
	Size int64  `json:"Token was found in row"`
}

type ScanResult struct {
	TotalBig int          `json:"total big files"`
	BigFiles []FileOutput `json:"files"`
	TotalSecret int          `json:"total secret files"`
	SecretFiles []FileOutput `json:"files"`
}
