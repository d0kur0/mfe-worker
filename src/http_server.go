package src

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"sync"
)

type HttpServer struct {
	di *DIContainer
}

func (h *HttpServer) requestBuild(c echo.Context) error {
	branch := c.Param("branch")
	projectID := c.Param("projectID")

	_ = branch
	_ = projectID

	h.di.queue.AddToQueue(func(wg *sync.WaitGroup) {
		defer wg.Done()

		image := Image{
			Branch:    branch,
			Status:    ImageStatusQueued,
			Revision:  "",
			ProjectId: projectID,
		}

		_ = image
	})

	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

func (h *HttpServer) SetupHttpHandlers() error {
	e := echo.New()

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.GET("/request-build/:projectID/:branch", h.requestBuild)

	log.Printf("Server started at http://localhost:%d", h.di.configMap.HttpPort)
	return e.Start(fmt.Sprintf(":%d", h.di.configMap.HttpPort))
}

func NewHttpServer(di *DIContainer) (*HttpServer, error) {
	return &HttpServer{
		di: di,
	}, nil
}
