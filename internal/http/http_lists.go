package http

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"mfe-worker/internal/dbDriver"
	"net/http"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Server) GetProjects(c echo.Context) error {
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

func (h *Server) GetBranches(c echo.Context) error {
	projectID := c.Param("projectID")
	limit, offset := getPagination(c)

	branches, total, err := h.di.DBDriver.GetBranches(projectID, dbDriver.Pagination{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Meta: ResponseMeta{ErrorCode: ErrorServerSuck},
		})
	}

	response := Response{
		Meta: ResponseMeta{
			Total:  int(total),
			Limit:  limit,
			Offset: offset,
		},
		Payload: branches,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Server) GetRevisions(c echo.Context) error {
	projectId := c.Param("projectId")
	branchName := c.Param("branch")
	limit, offset := getPagination(c)

	revisions, total, err := h.di.DBDriver.GetRevisions(projectId, branchName, dbDriver.Pagination{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Meta: ResponseMeta{ErrorCode: ErrorServerSuck},
		})
	}

	response := Response{
		Meta: ResponseMeta{
			Total:  int(total),
			Limit:  limit,
			Offset: offset,
		},
		Payload: revisions,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Server) GetBuilds(c echo.Context) error {
	projectId := c.Param("projectId")
	branchName := c.Param("branch")
	revision := c.Param("revision")
	limit, offset := getPagination(c)

	branch, err := h.di.DBDriver.GetBranch(projectId, branchName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, Response{
			Meta: ResponseMeta{ErrorCode: ErrorDataNotFound},
		})
	}

	builds, total, err := h.di.DBDriver.GetBuilds(branch, revision, dbDriver.Pagination{
		Limit:  limit,
		Offset: offset,
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, Response{
			Meta: ResponseMeta{ErrorCode: ErrorDataNotFound},
		})
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Meta: ResponseMeta{ErrorCode: ErrorServerSuck},
		})
	}

	response := Response{
		Meta: ResponseMeta{
			Total:  int(total),
			Limit:  limit,
			Offset: offset,
		},
		Payload: builds,
	}

	return c.JSON(http.StatusOK, response)
}
