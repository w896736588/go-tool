package gsapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gshttp"
	"github.com/w896736588/go-tool/gstool"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GsGitLab struct {
	BaseUrl     string
	AccessToken string
}

type GsGitLabParam struct {
	State   string
	Sort    string
	Page    int
	PerPage int
	RefName string
}

func (h *GsGitLab) GetProjects(param GsGitLabParam) ([]map[string]any, error) {
	urlProjects := fmt.Sprintf(h.BaseUrl+"/projects%s", h.getSqlParams(param))
	projectLists, resErr := gshttp.Get(urlProjects).Headers(map[string]string{
		`PRIVATE-TOKEN`: h.AccessToken,
	}).Request(20 * time.Second).Result()
	if resErr != nil {
		return nil, resErr
	}
	projectList := make([]map[string]any, 0)
	dErr := gstool.JsonDecode(cast.ToString(projectLists), &projectList)
	if dErr != nil {
		return nil, dErr
	}
	return projectList, nil
}

func (h *GsGitLab) GetProjectCommits(projectId string, param GsGitLabParam) ([]map[string]any, error) {
	urlProjectCommits := fmt.Sprintf(h.BaseUrl+"/projects/%s/repository/commits%s", projectId, h.getSqlParams(param))
	commitLists, resErr := gshttp.Get(urlProjectCommits).Headers(map[string]string{
		`PRIVATE-TOKEN`: h.AccessToken,
	}).Request(20 * time.Second).Result()
	if resErr != nil {
		return nil, resErr
	}
	commitList := make([]map[string]any, 0)
	dErr := gstool.JsonDecode(cast.ToString(commitLists), &commitList)
	if dErr != nil {
		return nil, dErr
	}
	return commitList, nil
}

func (h *GsGitLab) GetMerges(projectId string, param GsGitLabParam) ([]map[string]any, error) {
	urlMerges := fmt.Sprintf(h.BaseUrl+"/projects/%s/merge_requests%s", projectId, h.getSqlParams(param))
	mergeLists, resErr := gshttp.Get(urlMerges).Headers(map[string]string{
		`PRIVATE-TOKEN`: h.AccessToken,
	}).Request(10 * time.Second).Result()
	if resErr != nil {
		return nil, resErr
	}
	mergeList := make([]map[string]any, 0)
	decodeErr := gstool.JsonDecode(cast.ToString(mergeLists), &mergeList)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return mergeList, nil
}

func (h *GsGitLab) getSqlParams(param GsGitLabParam) string {
	paramList := make([]string, 0)
	if param.RefName != `` {
		paramList = append(paramList, `ref_name=`+param.RefName)
	}
	if param.Sort != `` {
		paramList = append(paramList, `sort=`+param.Sort)
	}
	if param.State != `` {
		paramList = append(paramList, `state=`+param.State)
	}
	if param.Page > 0 {
		paramList = append(paramList, `page=`+cast.ToString(param.Page))
	}
	if param.PerPage > 0 {
		paramList = append(paramList, `per_page=`+cast.ToString(param.PerPage))
	}
	if len(paramList) > 0 {
		return `?` + strings.Join(paramList, `&`)
	} else {
		return ``
	}
}

func (h *GsGitLab) CreateMerge(projectId string, opts *gitlab.CreateMergeRequestOptions) (*gitlab.MergeRequest, *gitlab.Response, error) {
	git, err := gitlab.NewClient(h.AccessToken, gitlab.WithBaseURL(h.BaseUrl))
	if err != nil {
		return nil, nil, err
	}
	mr, res, mrErr := git.MergeRequests.CreateMergeRequest(projectId, opts)
	if mrErr != nil {
		return nil, nil, mrErr
	}
	return mr, res, err
}
