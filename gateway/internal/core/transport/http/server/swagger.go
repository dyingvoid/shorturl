package core_http_server

import (
	"net/http"

	"github.com/dyingvoid/shorturl/gateway/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (s *Server) RegisterSwagger() {
	s.router.Handle(
		"/swagger/*",
		httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
			httpSwagger.DefaultModelsExpandDepth(-1),
		),
	)

	s.router.HandleFunc(
		"/swagger/doc.json",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
		},
	)
}
