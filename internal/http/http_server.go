package http

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"mfe-worker/internal/di"
	"net/url"
)

type Server struct {
	di *di.DIContainer
}

func (h *Server) SetupHttpHandlers() error {
	e := echo.New()

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Static("static", h.di.FSDriver.ImagesPath)

	e.GET("/request-build/:projectID/:branch", h.RequestBuild)
	e.GET("/projects", h.GetProjects)
	e.GET("/branches/:projectID", h.GetBranches)
	e.GET("/revisions/:projectID/:branch", h.GetRevisions)
	e.GET("/images/:projectID/:branch/:revision", h.GetImages)

	u, err := url.Parse(h.di.ConfigMap.HttpBaseUrl)
	if err != nil {
		return err
	}

	return e.Start(fmt.Sprintf("%s", u.Host))
}

func NewHttpServer(di *di.DIContainer) (*Server, error) {
	return &Server{
		di: di,
	}, nil
}
