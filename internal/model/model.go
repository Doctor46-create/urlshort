// Package model
package model

type JSONRequest struct {
	URL string `json:"url" validate:"required,url"`  
}

type JSONResponse struct {
	Result string `json:"result"`
}
