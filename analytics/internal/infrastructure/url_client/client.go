package url_client

import (
	"context"
	"fmt"
	"time"

	"github.com/dyingvoid/shorturl/analytics/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/analytics/internal/domain/errors"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn    *grpc.ClientConn
	urls    urlv1.URLServiceClient
	timeout time.Duration
}

func Connect(cfg Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.Addr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect url service: %w", err)
	}

	return &Client{
		conn:    conn,
		urls:    urlv1.NewURLServiceClient(conn),
		timeout: cfg.Timeout(),
	}, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close url service client: %w", err)
	}

	return nil
}

func (c *Client) GetURL(ctx context.Context, shortCode string) (domain.ShortURL, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.urls.GetURL(ctx, &urlv1.GetURLRequest{ShortCode: shortCode})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ShortURL{}, fmt.Errorf("url not found: %w", domain_errors.ErrNotFound)
		}

		return domain.ShortURL{}, fmt.Errorf("get url: %w", err)
	}

	return domain.ShortURL{
		ShortCode: shortCode,
		UserID:    resp.GetUserId(),
		IsActive:  resp.GetIsActive(),
	}, nil
}
