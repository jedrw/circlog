package circleci

type Project struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	ExternalUrl string `json:"external_url"`
}
