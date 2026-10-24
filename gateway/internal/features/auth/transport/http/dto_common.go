package auth_transport_http

import (
	"fmt"
	"strings"
)

func validateCredentials(email string, password string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}
