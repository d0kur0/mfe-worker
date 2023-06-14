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
	"net/url"
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

	if h.di.dbDriver.IsRevisionExists(projectID, branch, gitBranch.Commit.ShortID) {
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
			err := fsDriver.RemoveTmpDirForBuild(projectId, branch, revision)
			if err != nil {
				log.Printf("failed on clear tmp dir: %s", err)
			}
		}(h.di.fsDriver, projectID, branch, gitBranch.Commit.ShortID)

		tmpDirName := h.di.fsDriver.GetTmpPathForBuild(projectID, branch, gitBranch.Commit.ShortID)

		if h.di.fsDriver.HasTmpDirForBuild(projectID, branch, gitBranch.Commit.ShortID) {
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

		projectExists := h.di.fsDriver.HasProjectDir(projectID)
		if !projectExists {
			if err := h.di.fsDriver.CreateProjectDir(projectID); err != nil {
				return err
			}
		}

		branchExists := h.di.fsDriver.HasProjectBranchDir(projectID, branch)
		if !branchExists {
			if err := h.di.fsDriver.CreateProjectBranchDir(projectID, branch); err != nil {
				return err
			}
		}

		branchRevisionExists := h.di.fsDriver.HasBranchRevisionDir(projectID, branch, gitBranch.Commit.ShortID)
		if !branchRevisionExists {
			if err := h.di.fsDriver.CreateBranchRevisionDir(projectID, branch, gitBranch.Commit.ShortID); err != nil {
				return err
			}
		}

		fileList, err := h.di.fsDriver.PickFilesToWebStorage(projectFromConfig, gitBranch, tmpDirName)
		if err != nil {
			return err
		}

		var imageFiles []ImageFile
		for _, filePath := range fileList {
			imageFiles = append(imageFiles, ImageFile{
				WebPath: filePath,
				ImageId: image.ID,
			})
		}

		image.Files = imageFiles
		image.Status = ImageStatusReady
		return h.di.dbDriver.Update(&image)
	})

	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}

func (h *HttpServer) SetupHttpHandlers() error {
	e := echo.New()

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Static("images", h.di.fsDriver.imagesPath)

	e.GET("/request-build/:projectID/:branch", h.requestBuild)

	u, err := url.Parse(h.di.configMap.HttpBaseUrl)
	if err != nil {
		return err
	}

	return e.Start(fmt.Sprintf("%s", u.Host))
}

func NewHttpServer(di *DIContainer) (*HttpServer, error) {
	return &HttpServer{
		di: di,
	}, nil
}
