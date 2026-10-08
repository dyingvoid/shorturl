package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type Link struct {
	ID          string
	UserID      string
	OriginalURL string
	ShortCode   string
	ExpiresAt   *time.Time
}

func NewLink(
	userID, orignialURL string,
	expiresIn *time.Duration,
) (Link, error) {
	short, err := randomBase62(7)
	if err != nil {
		return Link{}, err
	}

	link := Link{
		UserID:      userID,
		OriginalURL: orignialURL,
		ShortCode:   short,
	}

	if expiresIn != nil {
		exp := time.Now().Add(*expiresIn)
		link.ExpiresAt = &exp
	}

	return link, nil
}

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func randomBase62(n int) (string, error) {
	out := make([]byte, n)
	max := big.NewInt(int64(len(base62Alphabet)))
	for i := range out {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("error generating random int: %w", err)
		}
		out[i] = base62Alphabet[idx.Int64()]
	}
	return string(out), nil
}
