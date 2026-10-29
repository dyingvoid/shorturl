package handler

import (
	"context"

	"github.com/dyingvoid/shorturl/analytics/internal/domain"
	analyticsv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/analytics/v1"
)

type Handler struct {
	analyticsv1.UnimplementedAnalyticsServiceServer
	analyticsService AnalyticsService
}

type AnalyticsService interface {
	GetAnalytics(
		ctx context.Context,
		shortCode string,
		userID string,
		days int32,
	) (domain.Analytics, error)
}

func New(analyticsService AnalyticsService) *Handler {
	return &Handler{
		analyticsService: analyticsService,
	}
}
