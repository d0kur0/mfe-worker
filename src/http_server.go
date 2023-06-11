package src

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/lo"
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

	projectFromConfig := lo.Reduce(h.di.configMap.Projects, func(agg *Project, item Project, index int) *Project {
		if item.ProjectID == projectID {
			return &item
		}
		return nil
	}, nil)

	if projectFromConfig == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "UNKNOWN_PROJECT_ID"})
	}

	if len(projectFromConfig.Branches) != 0 {
		hasBranch := lo.Filter(projectFromConfig.Branches, func(item string, index int) bool {
			return item == branch
		})

		if len(hasBranch) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"code": "UNKNOWN_BRANCH_OF_PROJECT"})
		}
	}

	h.di.queue.AddToQueue(func(wg *sync.WaitGroup) {
		defer wg.Done()

		getBranch, _, err := h.di.gitlabClient.Branches.GetBranch(projectID, branch)
		if err != nil {
			log.Printf("failed on get info of branch: %s", err)
			return
		}

		image := Image{
			Branch:    branch,
			Status:    ImageStatusQueued,
			Revision:  getBranch.Commit.ShortID,
			ProjectId: projectID,
		}

		if err := h.di.dbDriver.Save(&image); err != nil {
			log.Printf("failed on create image to db: %s", err)
			return
		}
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
