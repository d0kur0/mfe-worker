package http

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"mfe-worker/internal/dbDriver"
	"net/http"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Server) GetProjectsList(c echo.Context) error {
	limit, offset := getPagination(c)

	var projects []Project

	for _, project := range h.di.ConfigMap.Projects {
		projects = append(projects, Project{
			ID:   project.ProjectID,
			Name: project.ProjectName,
		})
	}

	response := Response{
		Meta: ResponseMeta{
			Total:  len(h.di.ConfigMap.Projects),
			Limit:  limit,
			Offset: offset,
		},
		Payload: lo.Slice(projects, offset, limit),
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Server) GetProjectImagesList(c echo.Context) error {
	projectID := c.Param("projectID")
	limit, offset := getPagination(c)

	images, err := h.di.DBDriver.GetImagesOfProject(projectID, dbDriver.Pagination{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "SERVER_WAS_SUCK"})
	}

	response := Response{
		Meta: ResponseMeta{
			Total:  len(h.di.ConfigMap.Projects),
			Limit:  limit,
			Offset: offset,
		},
		Payload: images,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Server) GetProjectBranches(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

func (h *Server) GetProjectBranchImages(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

func (h *Server) GetProjectBranchRevisions(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}
