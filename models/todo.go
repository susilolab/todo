package models

import "time"

type Todo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CategoryId  int       `json:"category_id"`
	Done        int       `json:"done"`
	DateCreated time.Time `json:"date_created"`
	DateUpdated time.Time `json:"date_updated"`
}
