package domain

import "time"

type UserView struct {
	ShortCode string    `json:"short_code"`
	UserID    string    `json:"user_id"`
	ClickedAt time.Time `json:"clicked_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
}

func NewUserView(ip, userAgent, referer string) UserView {
	return UserView{
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
	}
}

func (v *UserView) SetLink(link Link) {
	v.ShortCode = link.ShortCode
	v.UserID = link.UserID
	v.ClickedAt = time.Now().UTC()
}
