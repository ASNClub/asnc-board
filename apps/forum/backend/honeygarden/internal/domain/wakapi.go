package domain

import (
	"time"

	"github.com/google/uuid"
)

type WakapiAccount struct {
	UserID      uuid.UUID `json:"userId"`
	InstanceURL string    `json:"instanceUrl"`
	APIKey      string    `json:"-"`
	Username    string    `json:"username"`
	CreatedAt   time.Time `json:"createdAt"`
}

type WakapiStats struct {
	TotalSeconds    float64         `json:"totalSeconds"`
	DailyAverage    float64         `json:"dailyAverage"`
	Days            []WakapiDay     `json:"days"`
	Languages       []WakapiLang    `json:"languages"`
}

type WakapiDay struct {
	Date    string  `json:"date"`
	Hours   float64 `json:"hours"`
}

type WakapiLang struct {
	Name    string  `json:"name"`
	Percent float64 `json:"percent"`
	Color   string  `json:"color"`
}
