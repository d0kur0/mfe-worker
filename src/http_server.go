package src

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/lo"
	"github.com/xanzy/go-gitlab"
	"log"
	"net/http"
	"strings"
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

	gitBranch, _, err := h.di.gitlabClient.Branches.GetBranch(projectID, branch)
	if err != nil {
		log.Printf("failed on get info of branch: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "SERVER_WAS_SUCK"})
	}

	// TODO: move to db driver
	var hasImageWithSameRevision bool
	h.di.dbDriver.db.
		Model(&Image{}).
		Select("count(*) > 0").
		Where("revision = ? AND project_id = ? AND branch = ?", gitBranch.Commit.ShortID, projectID, branch).
		Find(&hasImageWithSameRevision)

	if hasImageWithSameRevision {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BRANCH_NOT_CHANGED"})
	}

	h.di.queue.AddToQueue(func(wg *sync.WaitGroup) error {
		defer wg.Done()

		gitProject, _, err := h.di.gitlabClient.Projects.GetProject(
			projectID,
			&gitlab.GetProjectOptions{},
		)

		if err != nil {
			return errors.Join(fmt.Errorf("failed on get info of project (ID: %s)", projectID), err)
		}

		image := Image{
			Branch:    branch,
			Status:    ImageStatusQueued,
			Revision:  gitBranch.Commit.ShortID,
			ProjectId: projectID,
		}

		if err := h.di.dbDriver.Save(&image); err != nil {
			return errors.Join(errors.New("failed on create image to db"), err)
		}

		defer func(fsDriver *FSDriver, projectId string, branch string, revision string) {
			err := fsDriver.removeTmpDirForBuild(projectId, branch, revision)
			if err != nil {
				log.Printf("failed on clear tmp dir: %s", err)
			}
		}(h.di.fsDriver, projectID, branch, gitBranch.Commit.ShortID)

		tmpDirName := h.di.fsDriver.getTmpPathForBuild(projectID, branch, gitBranch.Commit.ShortID)

		if h.di.fsDriver.hasTmpDirForBuild(projectID, branch, gitBranch.Commit.ShortID) {
			return fmt.Errorf("tmp dir already exists, skip: %s", tmpDirName)
		}

		clonePath := fmt.Sprintf(
			"%s://oauth2:%s@%s/%s/%s.git",
			"http",
			h.di.configMap.GitlabToken,
			strings.Split(h.di.configMap.GitlabUrl, "://")[1],
			gitProject.Namespace.FullPath,
			gitProject.Name,
		)

		cloneArgs := []string{"clone", "--single-branch", "--branch", branch, clonePath, tmpDirName}

		if _, err = ExecShellCommand("git", cloneArgs, ExecShellCommandArgs{}); err != nil {
			return errors.Join(fmt.Errorf("failed on clone project (args: %x)", cloneArgs), err)
		}

		for _, cmd := range projectFromConfig.BuildCommands {
			cmdSegments := strings.Split(cmd, " ")
			cmdName := cmdSegments[0]
			cmdArgs := lo.Slice(cmdSegments, 1, len(cmdSegments))

			if _, err = ExecShellCommand(cmdName, cmdArgs, ExecShellCommandArgs{Cwd: tmpDirName}); err != nil {
				return errors.Join(fmt.Errorf("failed on exec build command from cfg: %s ", cmd), err)
			}
		}

		projectExists := h.di.fsDriver.hasProject(projectID)
		if !projectExists {
			if err := h.di.fsDriver.createProject(projectID); err != nil {
				return err
			}
		}

		branchExists := h.di.fsDriver.hasProjectBranch(projectID, branch)
		if !branchExists {
			if err := h.di.fsDriver.createProjectBranch(projectID, branch); err != nil {
				return err
			}
		}

		branchRevisionExists := h.di.fsDriver.hasBranchRevision(projectID, branch, gitBranch.Commit.ShortID)
		if !branchRevisionExists {
			if err := h.di.fsDriver.createBranchRevision(projectID, branch, gitBranch.Commit.ShortID); err != nil {
				return err
			}
		}

		if err := h.di.fsDriver.pickFilesToWebStorage(projectFromConfig, gitBranch, tmpDirName); err != nil {
			return err
		}

		return nil
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
