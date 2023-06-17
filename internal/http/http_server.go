package http

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"mfe-worker/internal/depsInjection"
	"net/url"
)

type Server struct {
	di *depsInjection.DIContainer
}

func (h *Server) SetupHttpHandlers() error {
	e := echo.New()

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Static("images", h.di.FSDriver.ImagesPath)

	e.GET("/request-build/:projectID/:branch", h.RequestBuild)
	e.GET("/projects", h.GetProjectsList)
	e.GET("/project/:projectID/images", h.GetProjectImagesList)
	e.GET("/project/:projectID/branches", h.GetProjectBranches)
	e.GET("/branch-images/:projectID/:branch", h.GetProjectBranchImages)
	e.GET("/branch-revisions/:projectID/:branch", h.GetProjectBranchRevisions)

	u, err := url.Parse(h.di.ConfigMap.HttpBaseUrl)
	if err != nil {
		return err
	}

	return e.Start(fmt.Sprintf("%s", u.Host))
}

func NewHttpServer(di *depsInjection.DIContainer) (*Server, error) {
	return &Server{
		di: di,
	}, nil
}
