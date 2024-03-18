package model

import "time"

type Article struct {
	ID          int64
	SourceID    int64
	Title       string
	Summary     string
	Link        string
	PublishedAt time.Time
	CreatedAt   time.Time
	PostedAt    time.Time
}
