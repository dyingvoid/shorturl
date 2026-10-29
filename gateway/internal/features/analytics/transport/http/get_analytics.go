package analytics_transport_http

import (
	"fmt"
	"net/http"

	core_auth "github.com/dyingvoid/shorturl/gateway/internal/core/auth"
	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_request "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/request"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	analyticsv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/analytics/v1"
)

type GetAnalyticsResponse struct {
	ShortCode   string          `json:"short_code"`
	TotalClicks int64           `json:"total_clicks"`
	ClicksByDay []DayClicks     `json:"clicks_by_day"`
	TopReferers []RefererClicks `json:"top_referers"`
}

type DayClicks struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

type RefererClicks struct {
	Referer string `json:"referer"`
	Clicks  int64  `json:"clicks"`
}

// GetAnalytics returns click analytics for a short URL.
// @Summary Get analytics
// @Description Returns click analytics for a short URL owned by the authenticated user.
// @Tags analytics
// @Produce json
// @Security BearerAuth
// @Param code path string true "Short code"
// @Param days query int false "Number of days to aggregate" default(30)
// @Success 200 {object} GetAnalyticsResponse
// @Failure 400 {object} core_http_response.ErrorResponse
// @Failure 401 {object} core_http_response.ErrorResponse
// @Failure 403 {object} core_http_response.ErrorResponse
// @Failure 404 {object} core_http_response.ErrorResponse
// @Failure 500 {object} core_http_response.ErrorResponse
// @Router /analytics/{code} [get]
func (h *AnalyticsHTTPHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	code, err := core_http_request.PathValue(r, "code")
	if err != nil {
		response.Error(err, "failed to parse path parameters")
		return
	}

	days, err := core_http_request.QueryInt32(r, "days")
	if err != nil {
		response.Error(err, "failed to parse query parameters")
		return
	}
	if days != nil && (*days <= 0 || *days > 365) {
		response.Error(
			fmt.Errorf("days must be between 1 and 365: %w", core_errors.ErrInvalidArgument),
			"failed to parse query parameters",
		)
		return
	}

	var daysValue int32
	if days != nil {
		daysValue = *days
	}

	result, err := h.analytics.GetAnalytics(ctx, &analyticsv1.GetAnalyticsRequest{
		ShortCode: code,
		UserId:    core_auth.UserID(ctx),
		Days:      daysValue,
	})
	if err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to get analytics")
		return
	}

	clicksByDay := make([]DayClicks, 0, len(result.GetClicksByDay()))
	for _, day := range result.GetClicksByDay() {
		clicksByDay = append(clicksByDay, DayClicks{
			Date:   day.GetDate(),
			Clicks: day.GetClicks(),
		})
	}

	topReferers := make([]RefererClicks, 0, len(result.GetTopReferers()))
	for _, referer := range result.GetTopReferers() {
		topReferers = append(topReferers, RefererClicks{
			Referer: referer.GetReferer(),
			Clicks:  referer.GetClicks(),
		})
	}

	response.JSON(GetAnalyticsResponse{
		ShortCode:   result.GetShortCode(),
		TotalClicks: result.GetTotalClicks(),
		ClicksByDay: clicksByDay,
		TopReferers: topReferers,
	}, http.StatusOK)
}
