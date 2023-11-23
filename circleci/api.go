package circleci

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lupinelab/circlog/config"
)

const (
	// CircleCi Api Endpoints
	CIRCLECI_ENDPOINT_V1 = "https://circleci.com/api/v1.1"
	CIRCLECI_ENDPOINT_V2 = "https://circleci.com/api/v2"

	// Status types
	SUCCESS      = "success"
	RUNNING      = "running"
	NOT_RUN      = "not_run"
	FAILED       = "failed"
	ERROR        = "error"
	FAILING      = "failing"
	ONHOLD       = "on_hold"
	CANCELED     = "canceled"
	UNAUTHORIZED = "unauthorized"
)

type ResponseType interface {
	Pipeline | Workflow | Job | JobDetails
}

type ApiResponse[T ResponseType] struct {
	NextPageToken string `json:"next_page_token"`
	Items         []T    `json:"items"`
}

func parseResponseBody[T ResponseType](responseBody []byte) (*ApiResponse[T], error) {
	parsedApiResponse := new(ApiResponse[T])
	err := json.Unmarshal(responseBody, &parsedApiResponse)

	return parsedApiResponse, err
}

func MakeRequest[T ResponseType](url string, config config.CirclogConfig, numPages int, nextPageToken string) ([]T, string, error) {
	items := []T{}
	var branch string
	newItems := true
	page := 0

	for newItems && page != numPages {
		if nextPageToken != "" {
			nextPageToken = fmt.Sprintf("?page-token=%s", nextPageToken)
		}

		if nextPageToken != "" && config.Branch != "" {
			branch = fmt.Sprintf("&branch=%s", config.Branch)
		} else if config.Branch != "" {
			branch = fmt.Sprintf("?branch=%s", config.Branch)
		}

		endpoint := fmt.Sprintf("%s%s%s", url, nextPageToken, branch)

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return items, "", err
		}

		req.Header.Add("Circle-Token", config.Token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return items, "", err
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		parsedResponse, err := parseResponseBody[T](body)
		if err != nil {
			return items, "", err
		}

		items = append(items, parsedResponse.Items...)

		nextPageToken = parsedResponse.NextPageToken
		if nextPageToken == "" {
			newItems = false
		}

		page++
	}

	return items, nextPageToken, nil
}
