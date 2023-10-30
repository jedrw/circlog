package circleci

import (
	"fmt"

	"github.com/lupinelab/circlog/config"
)

type PipelineError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Actor struct {
	Login     string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
}

type PipelineTrigger struct {
	ReceivedAt string `json:"received_at"`
	Type       string `json:"type"`
	Actor      Actor  `json:"actor"`
}

type PipelineTriggerParameters struct {
	PropertyName interface{} `json:"property_name"`
}

type Commit struct {
	Body    string `json:"body"`
	Subject string `json:"subject"`
}

type Vcs struct {
	ProviderName        string `json:"provider_name"`
	TargetRepositoryUrl string `json:"target_repository_url"`
	Branch              string `json:"branch"`
	ReviewId            string `json:"review_id"`
	ReviewUrl           string `json:"revew_url"`
	Revision            string `json:"revision"`
	Tag                 string `json:"tag"`
	Commit              Commit `json:"commit"`
	OriginRepositoryUrl string `json:"origin_repository_url"`
}

type Pipeline struct {
	Id                string                    `json:"id"`
	Errors            []PipelineError           `json:"errors"`
	ProjectSlug       string                    `json:"project_slug"`
	UpdatedAt         string                    `json:"updated_at"`
	Number            int                       `json:"number"`
	TriggerParameters PipelineTriggerParameters `json:"trigger_parameters"`
	State             string                    `json:"state"`
	CreatedAt         string                    `json:"created_at"`
	Trigger           PipelineTrigger           `json:"trigger"`
	Vcs               Vcs                       `json:"vcs"`
}

func GetProjectPipelines(config config.CirclogConfig, project string) ([]Pipeline, error) {
	url := fmt.Sprintf("%s/project/%s/pipeline", CIRCLECI_ENDPOINT_V2, config.ProjectSlugV2(project))

	pipelines, err := collectPaginatedResponses[Pipeline](url, config)
	if err != nil {
		return []Pipeline{}, err
	}

	return pipelines, err
}
