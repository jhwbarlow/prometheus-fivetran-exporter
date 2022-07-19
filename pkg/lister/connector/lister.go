package connector

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/connector"
	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/jsonhttp"
	groupResolver "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resolver/group"
	connectorResp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/connector"
)

type Lister interface {
	List() ([]*connector.Connector, error)
	GroupID() string
}

type APILister struct {
	GroupResolver groupResolver.Resolver

	groupID    string
	apiToken   string
	httpClient *http.Client
	url        *url.URL
}

func NewAPILister(APIKey, APISecret, APIURL, groupID string,
	timeout time.Duration,
	groupResolver groupResolver.Resolver) (*APILister, error) {
	url, err := url.Parse(fmt.Sprintf("%s/v1/groups/%s/connectors?limit=1000", APIURL, groupID))
	if err != nil {
		return nil, fmt.Errorf("parsing API URL: %w", err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &APILister{
		GroupResolver: groupResolver,
		groupID:       groupID,
		url:           url,
		apiToken:      apiToken,
		httpClient:    httpClient,
	}, nil
}

func (l *APILister) List() ([]*connector.Connector, error) {
	// TODO: All this HTTP sending, receiving and unmarshalling stuff is common and generic.
	// DRY this out, possibly with generics (or just interface{}/any)
	// httpReq := &http.Request{
	// 	Header: make(http.Header),
	// 	Method: http.MethodGet,
	// 	URL:    l.url,
	// }
	// httpReq.Header.Add("Authorization", "Basic "+l.apiToken)

	// httpResp, err := l.httpClient.Do(httpReq)
	// if err != nil {
	// 	return nil, fmt.Errorf("sending HTTP GET request: %w", err)
	// }

	// if httpResp.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("received HTTP status code %d", httpResp.StatusCode)
	// }

	// respBody := new(bytes.Buffer)
	// if _, err := io.Copy(respBody, httpResp.Body); err != nil {
	// 	return nil, fmt.Errorf("copying HTTP response body: %w", err)
	// }
	// respBodyBytes := respBody.Bytes()
	// respBodyStr := string(respBodyBytes)

	// fmt.Printf("==>%#v\n", respBodyStr)

	// if err := httpResp.Body.Close(); err != nil {
	// 	return nil, fmt.Errorf("closing HTTP response body: %w", err)
	// }

	// listConnectorsResp := new(connectorResp.ListConnectorsResp)
	// if err := json.Unmarshal(respBodyBytes, listConnectorsResp); err != nil {
	// 	return nil, fmt.Errorf("unmarshalling HTTP response body: %w", err)
	// }

	// if listConnectorsResp.Code != resp.ResponseCodeSuccess {
	// 	return nil, fmt.Errorf("received response code %v", listConnectorsResp.Code)
	// }

	listConnectorsResp, err := jsonhttp.UnmarshallJSONFromHTTPGet[*connectorResp.ListConnectorsResp](l.url,
		l.apiToken,
		l.httpClient)
	if err != nil {
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	groupName, err := l.GroupResolver.ResolveIDToName(l.groupID)
	if err != nil {
		return nil, fmt.Errorf("resolving group ID %q to group name", l.groupID)
	}

	connectors := make([]*connector.Connector, 0, len(listConnectorsResp.Data.Items))
	for _, item := range listConnectorsResp.Data.Items {
		fmt.Printf("-->%#v\n", item)
		group := &connector.Connector{
			ID:                item.ID,
			Name:              item.Schema,
			GroupID:           l.groupID,
			GroupName:         groupName,
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

func (l *APILister) GroupID() string {
	return l.groupID
}
