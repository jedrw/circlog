package circleci

import (
	"fmt"
	"time"

	"github.com/lupinelab/circlog/config"
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

func GetPipelineWorkflows(config config.CirclogConfig, project string, pipelineId string) ([]Workflow, error) {
	url := fmt.Sprintf("%s/pipeline/%s/workflow", CIRCLECI_ENDPOINT_V2, pipelineId)

	workflows, err := collectPaginatedResponses[Workflow](url, config)
	if err != nil {
		return []Workflow{}, err
	}

	return workflows, err
}
