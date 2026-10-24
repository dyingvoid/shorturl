package users_client

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	users   userv1.UserServiceClient
	timeout time.Duration
}

func Connect(_ context.Context, cfg Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.Addr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect users service: %w", err)
	}

	return &Client{
		conn:    conn,
		users:   userv1.NewUserServiceClient(conn),
		timeout: cfg.Timeout(),
	}, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close users service client: %w", err)
	}

	return nil
}

func (c *Client) GetLimit(ctx context.Context, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.users.GetLimit(ctx, &userv1.GetLimitRequest{UserId: userID})
	if err != nil {
		return 0, fmt.Errorf("get limit: %w", err)
	}

	return int(resp.GetLinksLimit()), nil
}
