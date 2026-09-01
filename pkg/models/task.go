package models

import "time"

type Task struct {
	ID      int
	Title   string
	Comment string
	Date    time.Time
	Repeat  string
}
