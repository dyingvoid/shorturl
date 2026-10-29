package handler

import (
	"context"
	"fmt"

	analyticsv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/analytics/v1"
)

const dateLayout = "2006-01-02"

func (h *Handler) GetAnalytics(ctx context.Context, r *analyticsv1.GetAnalyticsRequest) (*analyticsv1.GetAnalyticsResponse, error) {
	result, err := h.analyticsService.GetAnalytics(ctx, r.ShortCode, r.UserId, r.Days)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to get analytics: %w", err))
	}

	clicksByDay := make([]*analyticsv1.DayClicks, 0, len(result.ClicksByDay))
	for _, day := range result.ClicksByDay {
		clicksByDay = append(clicksByDay, &analyticsv1.DayClicks{
			Date:   day.Date.Format(dateLayout),
			Clicks: day.Clicks,
		})
	}

	topReferers := make([]*analyticsv1.RefererClicks, 0, len(result.TopReferers))
	for _, referer := range result.TopReferers {
		topReferers = append(topReferers, &analyticsv1.RefererClicks{
			Referer: referer.Referer,
			Clicks:  referer.Clicks,
		})
	}

	return &analyticsv1.GetAnalyticsResponse{
		ShortCode:   result.ShortCode,
		TotalClicks: result.TotalClicks,
		ClicksByDay: clicksByDay,
		TopReferers: topReferers,
	}, nil
}
