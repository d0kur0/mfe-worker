package http

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
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
	requestBranch := c.Param("branch")
	requestProjectId := c.Param("projectId")

	projectFromConfig := lo.Reduce(h.di.ConfigMap.Projects, func(agg *configMap.Project, item configMap.Project, index int) *configMap.Project {
		if item.ProjectID == requestProjectId {
			return &item
		}
		return agg
	}, nil)

	if projectFromConfig == nil {
		return c.JSON(http.StatusBadRequest, Response{
			Meta:    ResponseMeta{ErrorCode: ErrorUnknownProject},
			Payload: nil,
		})
	}

	if len(projectFromConfig.Branches) != 0 {
		hasBranch := lo.Filter(projectFromConfig.Branches, func(item string, index int) bool {
			return item == requestBranch
		})

		if len(hasBranch) == 0 {
			return c.JSON(http.StatusBadRequest, Response{
				Meta:    ResponseMeta{ErrorCode: ErrorBranchNotAllowed},
				Payload: nil,
			})
		}
	}

	gitlabBranch, _, err := h.di.GitlabClient.Branches.GetBranch(requestProjectId, requestBranch)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, Response{
			Meta: ResponseMeta{ErrorCode: ErrorServerSuck},
		})
	}

	latestGitlabCommit := gitlabBranch.Commit

	branch, err := h.di.DBDriver.GetBranch(requestProjectId, requestBranch)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, Response{
			Meta: ResponseMeta{ErrorCode: ErrorServerSuck},
		})
	}

	if branch == nil {
		branch, err = h.di.DBDriver.CreateBranch(&dbDriver.Branch{
			Name:      requestBranch,
			ProjectId: requestProjectId,
		})

		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, Response{
				Meta: ResponseMeta{ErrorCode: ErrorServerSuck},
			})
		}
	}

	for _, revision := range branch.Revisions {
		if revision.Name == latestGitlabCommit.ID {
			return c.JSON(http.StatusConflict, Response{
				Meta: ResponseMeta{ErrorCode: ErrorRevisionExists},
			})
		}
	}

	revision, err := h.di.DBDriver.CreateRevision(&dbDriver.Revision{
		Name:     latestGitlabCommit.ID,
		BranchId: branch.ID,
	})

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, Response{
			Meta: ResponseMeta{ErrorCode: ErrorServerSuck},
		})
	}

	h.di.Queue.AddToQueue(func(wg *sync.WaitGroup) error {
		defer wg.Done()

		gitProject, _, err := h.di.GitlabClient.Projects.GetProject(
			requestProjectId,
			&gitlab.GetProjectOptions{},
		)

		if err != nil {
			return err
		}

		build, err := h.di.DBDriver.CreateBuild(&dbDriver.Build{
			Status:     dbDriver.BuildStatusInProgress,
			RevisionId: revision.ID,
		})

		if err != nil {
			return err
		}

		defer func(fsDriver *fsDriver.FSDriver, projectId string, branch string, revision string) {
			err := fsDriver.RemoveTmpDirForBuild(projectId, branch, revision)
			if err != nil {
				log.Printf("failed on clear tmp dir: %s", err)
			}
		}(h.di.FSDriver, requestProjectId, requestBranch, revision.Name)

		tmpDirName := h.di.FSDriver.GetTmpPathForBuild(requestProjectId, requestBranch, gitlabBranch.Commit.ID)

		if h.di.FSDriver.HasTmpDirForBuild(requestProjectId, requestProjectId, gitlabBranch.Commit.ID) {
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

		cloneArgs := []string{"clone", "--single-branch", "--branch", requestBranch, clonePath, tmpDirName}

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

		projectExists := h.di.FSDriver.HasProjectDir(requestProjectId)
		if !projectExists {
			if err := h.di.FSDriver.CreateProjectDir(requestProjectId); err != nil {
				return err
			}
		}

		branchExists := h.di.FSDriver.HasProjectBranchDir(requestProjectId, requestBranch)
		if !branchExists {
			if err := h.di.FSDriver.CreateProjectBranchDir(requestProjectId, requestBranch); err != nil {
				return err
			}
		}

		branchRevisionExists := h.di.FSDriver.HasBranchRevisionDir(requestProjectId, requestBranch, gitlabBranch.Commit.ID)
		if !branchRevisionExists {
			if err := h.di.FSDriver.CreateBranchRevisionDir(requestProjectId, requestBranch, gitlabBranch.Commit.ID); err != nil {
				return err
			}
		}

		pickedFiles, err := h.di.FSDriver.PickFilesToWebStorage(projectFromConfig, gitlabBranch, tmpDirName)
		if err != nil {
			return err
		}

		var buildFiles []dbDriver.BuildFiles
		for _, file := range pickedFiles {
			buildFiles = append(buildFiles, dbDriver.BuildFiles{
				Path:    file.Path,
				WebPath: file.WebPath,
				BuildId: build.ID,
			})
		}

		build.Files = buildFiles
		build.Status = dbDriver.BuildStatusReady
		_, err = h.di.DBDriver.UpdateBuild(build)
		return err
	})

	return c.JSON(http.StatusOK, Response{
		Payload: map[string]string{"code": "ADDED_TO_QUEUE"},
	})
}
