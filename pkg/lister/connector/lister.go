package connector

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/connector"
	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/jsonhttp"
	connectorResp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/connector"
)

type Lister interface {
	List() ([]*connector.Connector, error)
	GetGroupID() string
	GetGroupName() string
}

type APILister struct {
	GroupID    string
	GroupName  string
	apiToken   string
	httpClient *http.Client
	url        *url.URL
}

func NewAPILister(APIKey, APISecret, APIURL, groupID, groupName string,
	timeout time.Duration) (*APILister, error) {
	url, err := url.Parse(fmt.Sprintf("%s/v1/groups/%s/connectors?limit=1000", APIURL, groupID))
	if err != nil {
		return nil, fmt.Errorf("parsing API URL: %w", err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &APILister{
		GroupID:    groupID,
		GroupName:  groupName,
		url:        url,
		apiToken:   apiToken,
		httpClient: httpClient,
	}, nil
}

func (l *APILister) List() ([]*connector.Connector, error) {
	listConnectorsResp, err := jsonhttp.UnmarshallJSONFromHTTPGet[*connectorResp.ListConnectorsResp](l.url,
		l.apiToken,
		l.httpClient)
	if err != nil {
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	connectors := make([]*connector.Connector, 0, len(listConnectorsResp.Data.Items))
	for _, item := range listConnectorsResp.Data.Items {
		fmt.Printf("-->%#v\n", item)
		group := &connector.Connector{
			ID:                item.ID,
			Name:              item.Schema,
			GroupID:           l.GroupID,
			GroupName:         l.GroupName,
			Service:           item.Service,
			Paused:            item.Paused,
			IsHistoricalSync:  item.Status.IsHistoricalSync,
			SyncFrequencyMins: item.SyncFrequencyMins,
			TaskCount:         len(item.Status.Tasks),
			WarningCount:      len(item.Status.Warnings),
			SetupState:        item.Status.SetupState,
			SyncState:         item.Status.SyncState,
			UpdateState:       item.Status.UpdateState,
		}

		connectors = append(connectors, group)
	}

	return connectors, nil
}

func (l *APILister) GetGroupID() string {
	return l.GroupID
}

func (l *APILister) GetGroupName() string {
	return l.GroupName
}
