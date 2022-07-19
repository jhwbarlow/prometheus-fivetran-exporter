package destination

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/destination"
	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/jsonhttp"
	destinationResp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/destination"
)

type Describer interface {
	Describe() (*destination.Destination, error)
	GetGroupID() string
	GetGroupName() string
}

type APIDescriber struct {
	GroupID   string
	GroupName string

	apiToken   string
	httpClient *http.Client
	url        *url.URL
}

func NewAPIDescriber(APIKey, APISecret, APIURL, groupID, groupName string,
	timeout time.Duration) (*APIDescriber, error) {
	url, err := url.Parse(fmt.Sprintf("%s/v1/destinations/%s", APIURL, groupID))
	if err != nil {
		return nil, fmt.Errorf("parsing API URL: %w", err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &APIDescriber{
		GroupID:    groupID,
		GroupName:  groupName,
		url:        url,
		apiToken:   apiToken,
		httpClient: httpClient,
	}, nil
}

func (d *APIDescriber) Describe() (*destination.Destination, error) {
	describeDestinationResp, err := jsonhttp.UnmarshallJSONFromHTTPGet[*destinationResp.DescribeDestinationResp](d.url,
		d.apiToken,
		d.httpClient)
	if err != nil {
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	destination := &destination.Destination{
		ID:   describeDestinationResp.Data.ID,
		Name: d.GroupName, // XXX: There is no way to get this info directly from the destination
		// as Fivetran does not have the concept of a group name separate from the destination name.
		// Instead, we imply this from the name of the group.
		GroupID:     d.GroupID,
		GroupName:   d.GroupName,
		Service:     describeDestinationResp.Data.Service,
		SetupStatus: describeDestinationResp.Data.SetupStatus,
	}

	return destination, nil
}

func (d *APIDescriber) GetGroupID() string {
	return d.GroupID
}

func (d *APIDescriber) GetGroupName() string {
	return d.GroupName
}
