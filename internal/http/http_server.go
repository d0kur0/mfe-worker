package http

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/lo"
	"github.com/xanzy/go-gitlab"
	"log"
	"mfe-worker/internal/configMap"
	"mfe-worker/internal/dbDriver"
	"mfe-worker/internal/depsInjection"
	"mfe-worker/internal/fsDriver"
	"mfe-worker/internal/shell"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Server struct {
	di *depsInjection.DIContainer
}

func (h *Server) requestBuild(c echo.Context) error {
	branch := c.Param("branch")
	projectID := c.Param("projectID")

	projectFromConfig := lo.Reduce(h.di.ConfigMap.Projects, func(agg *configMap.Project, item configMap.Project, index int) *configMap.Project {
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

	gitBranch, _, err := h.di.GitlabClient.Branches.GetBranch(projectID, branch)
	if err != nil {
		log.Printf("failed on get info of branch: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "SERVER_WAS_SUCK"})
	}

	if h.di.DBDriver.IsRevisionExists(projectID, branch, gitBranch.Commit.ShortID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BRANCH_NOT_CHANGED"})
	}

	h.di.Queue.AddToQueue(func(wg *sync.WaitGroup) error {
		defer wg.Done()

		gitProject, _, err := h.di.GitlabClient.Projects.GetProject(
			projectID,
			&gitlab.GetProjectOptions{},
		)

		if err != nil {
			return errors.Join(fmt.Errorf("failed on get info of project (ID: %s)", projectID), err)
		}

		image := dbDriver.Image{
			Branch:    branch,
			Status:    dbDriver.ImageStatusQueued,
			Revision:  gitBranch.Commit.ShortID,
			ProjectId: projectID,
		}

		if err := h.di.DBDriver.Save(&image); err != nil {
			return errors.Join(errors.New("failed on create image to db"), err)
		}

		defer func(fsDriver *fsDriver.FSDriver, projectId string, branch string, revision string) {
			err := fsDriver.RemoveTmpDirForBuild(projectId, branch, revision)
			if err != nil {
				log.Printf("failed on clear tmp dir: %s", err)
			}
		}(h.di.FSDriver, projectID, branch, gitBranch.Commit.ShortID)

		tmpDirName := h.di.FSDriver.GetTmpPathForBuild(projectID, branch, gitBranch.Commit.ShortID)

		if h.di.FSDriver.HasTmpDirForBuild(projectID, branch, gitBranch.Commit.ShortID) {
			return fmt.Errorf("tmp dir already exists, skip: %s", tmpDirName)
		}

		clonePath := fmt.Sprintf(
			"%s://oauth2:%s@%s/%s/%s.git",
			"http",
			h.di.ConfigMap.GitlabToken,
			strings.Split(h.di.ConfigMap.GitlabUrl, "://")[1],
			gitProject.Namespace.FullPath,
			gitProject.Name,
		)

		cloneArgs := []string{"clone", "--single-branch", "--branch", branch, clonePath, tmpDirName}

		if _, err = shell.ExecShellCommand("git", cloneArgs, shell.ExecShellCommandArgs{}); err != nil {
			return errors.Join(fmt.Errorf("failed on clone project (args: %x)", cloneArgs), err)
		}

		for _, cmd := range projectFromConfig.BuildCommands {
			cmdSegments := strings.Split(cmd, " ")
			cmdName := cmdSegments[0]
			cmdArgs := lo.Slice(cmdSegments, 1, len(cmdSegments))

			if _, err = shell.ExecShellCommand(cmdName, cmdArgs, shell.ExecShellCommandArgs{Cwd: tmpDirName}); err != nil {
				return errors.Join(fmt.Errorf("failed on exec build command from cfg: %s ", cmd), err)
			}
		}

		projectExists := h.di.FSDriver.HasProjectDir(projectID)
		if !projectExists {
			if err := h.di.FSDriver.CreateProjectDir(projectID); err != nil {
				return err
			}
		}

		branchExists := h.di.FSDriver.HasProjectBranchDir(projectID, branch)
		if !branchExists {
			if err := h.di.FSDriver.CreateProjectBranchDir(projectID, branch); err != nil {
				return err
			}
		}

		branchRevisionExists := h.di.FSDriver.HasBranchRevisionDir(projectID, branch, gitBranch.Commit.ShortID)
		if !branchRevisionExists {
			if err := h.di.FSDriver.CreateBranchRevisionDir(projectID, branch, gitBranch.Commit.ShortID); err != nil {
				return err
			}
		}

		fileList, err := h.di.FSDriver.PickFilesToWebStorage(projectFromConfig, gitBranch, tmpDirName)
		if err != nil {
			return err
		}

		var imageFiles []dbDriver.ImageFile
		for _, filePath := range fileList {
			imageFiles = append(imageFiles, dbDriver.ImageFile{
				WebPath: filePath,
				ImageId: image.ID,
			})
		}

		image.Files = imageFiles
		image.Status = dbDriver.ImageStatusReady
		return h.di.DBDriver.Update(&image)
	})

	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Server) getProjectsList(c echo.Context) error {
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

func (h *Server) getProjectImagesList(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

func (h *Server) getProjectBranches(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

func (h *Server) getProjectBranchImages(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

func (h *Server) getProjectBranchRevisions(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
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

	e.GET("/request-build/:projectID/:branch", h.requestBuild)
	e.GET("/projects", h.getProjectsList)
	e.GET("/project-images/:projectID", h.getProjectImagesList)
	e.GET("/project-branches/:projectID", h.getProjectBranches)
	e.GET("/branches/:projectID", h.getProjectBranches)
	e.GET("/branch-images/:projectID/:branch", h.getProjectBranchImages)
	e.GET("/branch-revisions/:projectID/:branch", h.getProjectBranchRevisions)

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
