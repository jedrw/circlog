package circleci

import (
	"fmt"
	"time"

	"github.com/jedrw/circlog/config"
)

type Workflow struct {
	PipelineId     string    `json:"pipeline_id"`
	CanceledBy     string    `json:"canceled_by"`
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	ProjectSlug    string    `json:"project_slug"`
	ErroredBy      string    `json:"errored_by"`
	Tag            string    `json:"tag"`
	Status         string    `json:"status"`
	StartedBy      string    `json:"started_by"`
	PipelineNumber int       `json:"pipeline_number"`
	CreatedAt      time.Time `json:"created_at"`
	StoppedAt      time.Time `json:"stopped_at"`
}

func GetPipelineWorkflows(config config.CirclogConfig, pipelineId string, numPages int, nextPageToken string) ([]Workflow, string, error) {
	url := fmt.Sprintf("%s/pipeline/%s/workflow", CIRCLECI_ENDPOINT_V2, pipelineId)

	workflows, nextPageToken, err := MakeRequest[Workflow](url, config, numPages, nextPageToken)
	if err != nil {
		return []Workflow{}, nextPageToken, err
	}

	return workflows, nextPageToken, err
}
