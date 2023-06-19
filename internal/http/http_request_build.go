package http

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/xanzy/go-gitlab"
	"log"
	"mfe-worker/internal/configMap"
	"mfe-worker/internal/dbDriver"
	"mfe-worker/internal/fsDriver"
	"mfe-worker/internal/shell"
	"net/http"
	"strings"
	"sync"
)

func (h *Server) RequestBuild(c echo.Context) error {
	branch := c.Param("branch")
	projectID := c.Param("projectID")

	projectFromConfig := lo.Reduce(h.di.ConfigMap.Projects, func(agg *configMap.Project, item configMap.Project, index int) *configMap.Project {
		if item.ProjectID == projectID {
			return &item
		}
		return agg
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

		pickedFiles, err := h.di.FSDriver.PickFilesToWebStorage(projectFromConfig, gitBranch, tmpDirName)
		if err != nil {
			return err
		}

		var imageFiles []dbDriver.ImageFile
		for _, file := range pickedFiles {
			imageFiles = append(imageFiles, dbDriver.ImageFile{
				Path:    file.Path,
				WebPath: file.WebPath,
				ImageId: image.ID,
			})
		}

		image.Files = imageFiles
		image.Status = dbDriver.ImageStatusReady
		return h.di.DBDriver.Update(&image)
	})

	return c.JSON(http.StatusOK, map[string]string{"code": "ADDED_TO_QUEUE"})
}
