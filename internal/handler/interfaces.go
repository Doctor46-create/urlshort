package handler

import "net/http"

type URLHandler interface {
	ShortenURL(w http.ResponseWriter, r *http.Request)
	RedirectURL(w http.ResponseWriter, r *http.Request)
	ShortenURLJSON(w http.ResponseWriter, r *http.Request)
	PingDB(w http.ResponseWriter, r *http.Request)
	ShortenBatchURL(w http.ResponseWriter, r *http.Request)
	GetUserURLs(w http.ResponseWriter, r *http.Request)
	DeleteURLs(w http.ResponseWriter, r *http.Request)
}
