package models

import "time"

type Entry struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	Story     string    `json:"story"`
	CreatedAt time.Time `json:"created_at"`
	Photos    []Photo   `json:"photos"`
}

type Photo struct {
	ID        string `json:"id"`
	EntryID   string `json:"entry_id"`
	RawPath   string `json:"raw_path"`
	ThumbPath string `json:"thumb_path"`
	Status    string `json:"status"`
}
