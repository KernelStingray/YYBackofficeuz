package domain

// Media represents a stored file's metadata
type Media struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}
