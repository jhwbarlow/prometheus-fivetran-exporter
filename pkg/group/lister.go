package group

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/jsonhttp"
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp/group"
)

type Lister interface {
	List() ([]*Group, error)
}

type APILister struct {
	APIURL string

	apiToken   string
	httpClient *http.Client
	url        *url.URL
}

func NewAPILister(APIKey, APISecret, APIURL string, timeout time.Duration) (*APILister, error) {
	url, err := url.Parse(APIURL + "/v1/groups?limit=1000")
	if err != nil {
		return nil, fmt.Errorf("parsing API URL: %w", err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &APILister{
		url:        url,
		apiToken:   apiToken,
		httpClient: httpClient,
	}, nil
}

func (l *APILister) List() ([]*Group, error) {
	listGroupsResp, err := jsonhttp.UnmarshallJSONFromHTTPGet[*apiresp.ListGroupsResp](l.url,
		l.apiToken,
		l.httpClient)
	if err != nil {
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	groups := make([]*Group, 0, len(listGroupsResp.Data.Items))
	for _, item := range listGroupsResp.Data.Items {
		group := &Group{
			ID:   item.ID,
			Name: item.Name,
		}

		groups = append(groups, group)
	}

	return groups, nil
}
