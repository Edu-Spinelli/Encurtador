package repository

import "github.com/spinelli/encurtador-links/shared/model"

type URLRepository interface {
	Save(shortCode, longURL string) error
	FindByShortCode(shortCode string) (*model.URL, error)
}
