package model

import "time"

type App struct {
	ID         uint64    `json:"id"`
	AppCode    string    `json:"app_code"`
	Name       string    `json:"name"`
	AppKey     string    `json:"app_key"`
	AppSecret  string    `json:"app_secret,omitempty"`
	Status     string    `json:"status"`
	Scenario   []string  `json:"scenario"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}
