package http

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"mfe-worker/internal/di"
	"net/url"
)

type Server struct {
	di *di.Container
}

func (h *Server) SetupHttpHandlers() error {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Static("static", h.di.FSDriver.ImagesPath)

	e.GET("/request-build/:projectId/:branch", h.RequestBuild)
	e.GET("/projects", h.GetProjects)
	e.GET("/branches/:projectId", h.GetBranches)
	e.GET("/revisions/:projectId/:branch", h.GetRevisions)
	e.GET("/builds/:projectId/:branch/:revision", h.GetBuilds)

	u, err := url.Parse(h.di.ConfigMap.HttpBaseUrl)
	if err != nil {
		return err
	}

	return e.Start(fmt.Sprintf("%s", u.Host))
}

func NewHttpServer(di *di.Container) (*Server, error) {
	return &Server{
		di: di,
	}, nil
}
