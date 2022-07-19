package destination

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/jsonhttp"
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp/destination"
)

type Describer interface {
	Describe() (*Destination, error)
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

func (d *APIDescriber) Describe() (*Destination, error) {
	describeDestinationResp, err := jsonhttp.UnmarshallJSONFromHTTPGet[*apiresp.DescribeDestinationResp](d.url,
		d.apiToken,
		d.httpClient)
	if err != nil {
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	setupStatus, err := convertSetupStatus(describeDestinationResp.Data.SetupStatus)
	if err != nil {
		return nil, fmt.Errorf("converting Setup Status: %w", err)
	}

	destination := &Destination{
		ID:   describeDestinationResp.Data.ID,
		Name: d.GroupName, // XXX: There is no way to get this info directly from the destination
		// as Fivetran does not have the concept of a group name separate from the destination name.
		// Instead, we imply this from the name of the group.
		GroupID:     d.GroupID,
		GroupName:   d.GroupName,
		Service:     describeDestinationResp.Data.Service,
		SetupStatus: setupStatus,
	}

	return destination, nil
}

func (d *APIDescriber) GetGroupID() string {
	return d.GroupID
}

func (d *APIDescriber) GetGroupName() string {
	return d.GroupName
}

func convertSetupStatus(apiSetupStatus apiresp.SetupStatus) (SetupStatus, error) {
	switch apiSetupStatus {
	case apiresp.SetupStatusIncomplete:
		return SetupStatusIncomplete, nil
	case apiresp.SetupStatusBroken:
		return SetupStatusBroken, nil
	case apiresp.SetupStatusConnected:
		return SetupStatusConnected, nil
	default:
		return SetupStatus(""), fmt.Errorf("illegal API Setup Status: %q", apiSetupStatus)
	}
}
