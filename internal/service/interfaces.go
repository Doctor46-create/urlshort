package service

type Shortener interface {
	Shorten(originalURL string) (shortKey string, err error)
	GetOriginal(shortKey string) (originalURL string, err error)
}
