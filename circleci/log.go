package circleci

import (
	"fmt"
	"io"
	"net/http"

	"github.com/lupinelab/circlog/config"
)

func GetStepLogs(config config.CirclogConfig, jobNumber int64, stepNumber int64, stepIndex int64, allocationId string) (string, error) {
	url := fmt.Sprintf("%s/project/%s/%d/output/%d/%d?file=true&allocation-id=%s", CIRCLECI_ENDPOINT_V1, config.ProjectSlugV1(), jobNumber, stepNumber, stepIndex, allocationId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Circle-Token", config.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return string(body), err
}
