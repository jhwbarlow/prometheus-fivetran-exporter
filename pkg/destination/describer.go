package destination

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/jsonhttp"
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp/destination"
	"go.uber.org/zap"
)

type Describer interface {
	Describe() (*Destination, error)
	GetGroupID() string
	GetGroupName() string
}

type APIDescriber struct {
	GroupID   string
	GroupName string

	unmarshaller *jsonhttp.JSONHTTPUnmarshaller[*apiresp.DescribeDestinationResp]
	logger       *zap.SugaredLogger
}

func NewAPIDescriber(logger *zap.SugaredLogger,
	APIKey, APISecret, APIURL, groupID, groupName string,
	timeout time.Duration) (*APIDescriber, error) {
	logger = getComponentLogger(logger, "api-describer")

	url, err := url.Parse(fmt.Sprintf("%s/v1/destinations/%s", APIURL, groupID))
	if err != nil {
		logger.Errorw("parsing API URL", "url", url, "error", err)
		return nil, fmt.Errorf("parsing API URL %q: %w", APIURL, err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	unmarshaller := jsonhttp.NewJSONHTTPUnmarshaller[*apiresp.DescribeDestinationResp](logger,
		url,
		apiToken,
		httpClient)

	return &APIDescriber{
		GroupID:      groupID,
		GroupName:    groupName,
		unmarshaller: unmarshaller,
		logger:       logger,
	}, nil
}

func (d *APIDescriber) Describe() (*Destination, error) {
	describeDestinationResp, err := d.unmarshaller.UnmarshallJSONFromHTTPGet()
	if err != nil {
		d.logger.Errorw("getting JSON HTTP response", "error", err)
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	id := describeDestinationResp.Data.ID
	name := d.GroupName // XXX: There is no way to get this info directly from the destination
	// as Fivetran does not have the concept of a group name separate from the destination name.
	// Instead, we imply this from the name of the group.
	groupID := d.GroupID
	groupName := d.GroupName

	setupStatus, err := d.convertSetupStatus(describeDestinationResp.Data.SetupStatus)
	if err != nil {
		d.logger.Errorw("converting Setup Status",
			"id", id,
			"name", name,
			"group_id", groupID,
			"group_name", groupName,
			"setup_status", describeDestinationResp.Data.SetupStatus,
			"error", err)
		return nil, fmt.Errorf("converting Setup Status: %w", err)
	}

	destination := &Destination{
		ID:          id,
		Name:        name,
		GroupID:     groupID,
		GroupName:   groupName,
		Service:     describeDestinationResp.Data.Service,
		SetupStatus: setupStatus,
	}

	d.logger.Infow("discovered destination",
		"id", id,
		"name", name,
		"group_id", groupID,
		"group_name", groupName)
	return destination, nil
}

func (d *APIDescriber) GetGroupID() string {
	return d.GroupID
}

func (d *APIDescriber) GetGroupName() string {
	return d.GroupName
}

func (d *APIDescriber) convertSetupStatus(apiSetupStatus apiresp.SetupStatus) (SetupStatus, error) {
	switch apiSetupStatus {
	case apiresp.SetupStatusIncomplete:
		return SetupStatusIncomplete, nil
	case apiresp.SetupStatusBroken:
		return SetupStatusBroken, nil
	case apiresp.SetupStatusConnected:
		return SetupStatusConnected, nil
	default:
		d.logger.Errorw("illegal API Setup Status", "setup_status", apiSetupStatus)
		return SetupStatus(""), fmt.Errorf("illegal API Setup Status: %q", apiSetupStatus)
	}
}
