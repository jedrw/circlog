package circleci

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lupinelab/circlog/config"
)

type Action struct {
	Index              int64     `json:"index"`
	Step               int64     `json:"step"`
	AllocationId       string    `json:"allocation_id"`
	Name               string    `json:"name"`
	Type               string    `json:"type"`
	StartTime          time.Time `json:"start_time"`
	Truncated          bool      `json:"truncated"`
	Parallel           bool      `json:"parallel"`
	BashCommand        string    `json:"bash_command"`
	Background         bool      `json:"background"`
	Insignificant      bool      `json:"insignificant"`
	HasOutput          bool      `json:"has_output"`
	Continue           bool      `json:"continue"`
	ExitCode           int64     `json:"exit_code"`
	OutputUrl          string    `json:"output_url"`
	Status             string    `json:"status"`
	Failed             bool      `json:"failed"`
	InfrastructureFail bool      `json:"infrastructure_fail"`
	Timedout           bool      `json:"timedout"`
	Canceled           bool      `json:"canceled"`
}

type Step struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

func GetJobSteps(config config.CirclogConfig, project string, jobNumber int64) (JobDetails, error) {
	url := fmt.Sprintf("%s/project/%s/%d", CIRCLECI_ENDPOINT_V1, config.ProjectSlugV1(project), jobNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return JobDetails{}, err
	}

	req.Header.Add("Circle-Token", config.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return JobDetails{}, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var jobDetails JobDetails
	err = json.Unmarshal(body, &jobDetails)
	if err != nil {
		return JobDetails{}, err
	}

	return jobDetails, err
}
