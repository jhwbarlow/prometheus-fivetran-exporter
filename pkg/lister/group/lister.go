package group

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/group"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp"
	groupResp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/group"
)

type Lister interface {
	List() ([]*group.Group, error)
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

func (l *APILister) List() ([]*group.Group, error) {
	httpReq := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL:    l.url,
	}
	httpReq.Header.Add("Authorization", "Basic "+l.apiToken)

	httpResp, err := l.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP GET request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received HTTP status code %d", httpResp.StatusCode)
	}

	respBody := new(bytes.Buffer)
	if _, err := io.Copy(respBody, httpResp.Body); err != nil {
		return nil, fmt.Errorf("copying HTTP response body: %w", err)
	}

	if err := httpResp.Body.Close(); err != nil {
		return nil, fmt.Errorf("closing HTTP response body: %w", err)
	}

	listGroupsResp := new(groupResp.ListGroupsResp)
	if err := json.Unmarshal(respBody.Bytes(), listGroupsResp); err != nil {
		return nil, fmt.Errorf("unmarshalling HTTP response body: %w", err)
	}

	if listGroupsResp.Code != resp.ResponseCodeSuccess {
		return nil, fmt.Errorf("received response code %v", listGroupsResp.Code)
	}

	groups := make([]*group.Group, 0, len(listGroupsResp.Data.Items))
	for _, item := range listGroupsResp.Data.Items {
		group := &group.Group{
			ID:   item.ID,
			Name: item.Name,
		}

		groups = append(groups, group)
	}

	return groups, nil
}
