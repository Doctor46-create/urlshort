// Package model
package model

type JSONRequest struct {
	URL string `json:"url" validate:"required,url"`  
}

type JSONResponse struct {
	Result string `json:"result"`
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
