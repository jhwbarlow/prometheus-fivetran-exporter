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

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/connector"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp"
	connectorResp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/connector"
)

type Lister interface {
	List() ([]*connector.Connector, error)
}

type APILister struct {
	apiToken   string
	httpClient *http.Client
	url        *url.URL
}

func NewAPILister(APIKey, APISecret, APIURL, groupID string, timeout time.Duration) (*APILister, error) {
	url, err := url.Parse(fmt.Sprintf("%s/v1/groups/%s/connectors?limit=1000", APIURL, groupID))
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

func (l *APILister) List() ([]*connector.Connector, error) {
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

	listConnectorsResp := new(connectorResp.ListConnectorsResp)
	if err := json.Unmarshal(respBody.Bytes(), listConnectorsResp); err != nil {
		return nil, fmt.Errorf("unmarshalling HTTP response body: %w", err)
	}

	if listConnectorsResp.Code != resp.ResponseCodeSuccess {
		return nil, fmt.Errorf("received response code %v", listConnectorsResp.Code)
	}

	connectors := make([]*connector.Connector, 0, len(listConnectorsResp.Data.Items))
	for _, item := range listConnectorsResp.Data.Items {
		group := &connector.Connector{
			ID:      item.ID,
			Name:    item.Schema,
			Service: item.Service,
		}

		connectors = append(connectors, group)
	}

	return connectors, nil
}
