package circleci

import (
	"fmt"
	"time"

	"github.com/lupinelab/circlog/config"
)

type Job struct {
	CanceledBy        string    `json:"canceled_by"`
	Dependencies      []string  `json:"dependencies"`
	JobNumber         int64     `json:"job_number"`
	Id                string    `json:"id"`
	StartedAt         time.Time `json:"started_at"`
	Name              string    `json:"name"`
	ApprovedBy        string    `json:"approved_by"`
	ProjectSlug       string    `json:"project_slug"`
	Status            string    `json:"status"`
	Type              string    `json:"type"`
	StoppedAt         time.Time `json:"stopped_at"`
	ApprovalRequestId string    `json:"approval_request_id"`
}

type ParallelRun struct {
	Index  int64  `json:"index"`
	Status string `json:"status"`
}

type Executor struct {
	ResourceClass string `json:"resource_class"`
	Type          string `json:"type"`
}

type Message struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

type Context struct {
	Name string `json:"name"`
}

type Organisation struct {
	Name string `json:"name"`
}

func GetWorkflowJobs(config config.CirclogConfig, workflowId string, numPages int, nextPageToken string) ([]Job, string, error) {
	url := fmt.Sprintf("%s/workflow/%s/job", CIRCLECI_ENDPOINT_V2, workflowId)

	jobs, nextPageToken, err := MakeRequest[Job](url, config, numPages, nextPageToken)
	if err != nil {
		return []Job{}, nextPageToken, err
	}

	return jobs, nextPageToken, err
}
