package domain

import (
	"fmt"
	"regexp"
	"unicode"

	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

const minPasswordLen = 6

type Email struct {
	value string
}

func NewEmail(value string) Email {
	return Email{value: value}
}

func (e Email) String() string {
	return e.value
}

func (e Email) Validate() error {
	if !emailRegex.MatchString(e.value) {
		return fmt.Errorf("%w: invalid email format", domain_errors.ErrInvalidArgument)
	}
	return nil
}

type Password struct {
	value string
}

func NewPassword(value string) Password {
	return Password{value: value}
}

func (p Password) String() string {
	return p.value
}

func (p Password) Validate() error {
	if len(p.value) < minPasswordLen {
		return fmt.Errorf(
			"%w: password must be at least %d characters",
			domain_errors.ErrInvalidArgument,
			minPasswordLen,
		)
	}

	var hasUpper, hasDigit bool
	for _, r := range p.value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
		if hasUpper && hasDigit {
			break
		}
	}

	if !hasUpper && !hasDigit {
		return fmt.Errorf(
			"%w: password must contain at least one uppercase letter or one digit",
			domain_errors.ErrInvalidArgument,
		)
	}

	return nil
}

type UserCredentials struct {
	Email    Email
	Password Password
}

func NewUserCredentials(email, password string) UserCredentials {
	return UserCredentials{
		Email:    NewEmail(email),
		Password: NewPassword(password),
	}
}

func (c *UserCredentials) Validate() error {
	if err := c.Email.Validate(); err != nil {
		return err
	}
	if err := c.Password.Validate(); err != nil {
		return err
	}
	return nil
}
