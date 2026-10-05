package jwt

import (
	"fmt"
	"time"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	Email     string           `json:"email"`
	TokenType domain.TokenType `json:"token_type"`
	jwtlib.RegisteredClaims
}

type TokenService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenService(config Config) *TokenService {
	return &TokenService{
		secret:          []byte(config.Secret),
		accessTokenTTL:  config.AccessTokenTTL,
		refreshTokenTTL: config.RefreshTokenTTL,
	}
}

func (s *TokenService) Issue(user domain.User) (domain.TokenPair, error) {
	now := time.Now()

	accessToken, err := s.sign(user, now, s.accessTokenTTL, domain.TokenTypeAccess)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("sign access token: %w", err)
	}

	refreshToken, err := s.sign(user, now, s.refreshTokenTTL, domain.TokenTypeRefresh)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *TokenService) ParseAccess(token string) (domain.TokenClaims, error) {
	return s.parse(token, domain.TokenTypeAccess)
}

func (s *TokenService) ParseRefresh(token string) (domain.TokenClaims, error) {
	return s.parse(token, domain.TokenTypeRefresh)
}

func (s *TokenService) sign(
	user domain.User,
	now time.Time,
	ttl time.Duration,
	tokenType domain.TokenType,
) (string, error) {
	claims := Claims{
		Email:     user.Email,
		TokenType: tokenType,
		Subject:   user.ID.String(),
		IssuedAt:  jwtlib.NewNumericDate(now),
		ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (s *TokenService) parse(
	token string,
	expectedType domain.TokenType,
) (domain.TokenClaims, error) {
	claims := &Claims{}

	parsed, err := jwtlib.ParseWithClaims(
		token,
		claims,
		s.keyFunc,
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithExpirationRequired(),
	)
	if err != nil {
		return domain.TokenClaims{}, fmt.Errorf(
			"parse token: %v: %w", err, domain_errors.ErrUnauthenticated,
		)
	}

	if !parsed.Valid || claims.TokenType != expectedType {
		return domain.TokenClaims{}, fmt.Errorf(
			"invalid token: %w", domain_errors.ErrUnauthenticated,
		)
	}

	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return domain.TokenClaims{}, fmt.Errorf(
			"missing time claims: %w", domain_errors.ErrUnauthenticated,
		)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return domain.TokenClaims{}, fmt.Errorf(
			"parse subject: %v: %w", err, domain_errors.ErrUnauthenticated,
		)
	}

	return domain.TokenClaims{
		UserID:    userID,
		Email:     claims.Email,
		Type:      claims.TokenType,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (s *TokenService) keyFunc(token *jwtlib.Token) (any, error) {
	if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return s.secret, nil
}
