package domain

import "time"

type DayClicks struct {
	Date   time.Time
	Clicks int64
}

type RefererClicks struct {
	Referer string
	Clicks  int64
}

type Analytics struct {
	ShortCode   string
	TotalClicks int64
	ClicksByDay []DayClicks
	TopReferers []RefererClicks
}

type ShortURL struct {
	ShortCode string
	UserID    string
	IsActive  bool
}
