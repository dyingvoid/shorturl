package domain

import "time"

type Link struct {
	ID          int64
	UserID      string
	OriginalURL string
	ExpiresAt   *time.Time
}

func NewLink(
	userID, orignialURL string,
	expiresIn *time.Duration,
) Link {
	link := Link{
		UserID:      userID,
		OriginalURL: orignialURL,
	}

	if expiresIn != nil {
		exp := time.Now().Add(*expiresIn)
		link.ExpiresAt = &exp
	}

	return link
}
