package model

import "time"

type URL struct {
	ShortCode string    `json:"short_code" db:"short_code"`
	LongURL   string    `json:"long_url" db:"long_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
