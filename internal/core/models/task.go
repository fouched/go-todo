package models

type Category string

const (
	CategoryPersonal Category = "PERSONAL"
	CategoryWork     Category = "WORK"
)

type Task struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    Category `json:"category"`
	IsCompleted bool     `json:"is_completed"`
	UserID      int64    `json:"user_id"`
}
