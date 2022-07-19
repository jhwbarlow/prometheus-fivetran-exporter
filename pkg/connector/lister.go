package connector

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/jsonhttp"
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp/connector"
)

type Lister interface {
	List() ([]*Connector, error)
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

func (l *APILister) List() ([]*Connector, error) {
	listConnectorsResp, err := jsonhttp.UnmarshallJSONFromHTTPGet[*apiresp.ListConnectorsResp](l.url,
		l.apiToken,
		l.httpClient)
	if err != nil {
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	connectors := make([]*Connector, 0, len(listConnectorsResp.Data.Items))
	for _, item := range listConnectorsResp.Data.Items {
		log.Printf("-->%#v\n", item) // TODO: Proper logging

		setupState, err := convertSetupState(item.Status.SetupState)
		if err != nil {
			return nil, fmt.Errorf("converting Setup State: %w", err)
		}

		syncState, err := convertSyncState(item.Status.SyncState)
		if err != nil {
			return nil, fmt.Errorf("converting Sync State: %w", err)
		}

		updateState, err := convertUpdateState(item.Status.UpdateState)
		if err != nil {
			return nil, fmt.Errorf("converting Update State: %w", err)
		}

		group := &Connector{
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
			SetupState:        setupState,
			SyncState:         syncState,
			UpdateState:       updateState,
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

func convertSetupState(apiSetupState apiresp.SetupState) (SetupState, error) {
	switch apiSetupState {
	case apiresp.SetupStateIncomplete:
		return SetupStateIncomplete, nil
	case apiresp.SetupStateBroken:
		return SetupStateBroken, nil
	case apiresp.SetupStateConnected:
		return SetupStateConnected, nil
	default:
		return SetupState(""), fmt.Errorf("illegal API Setup State: %q", apiSetupState)
	}
}

func convertSyncState(apiSyncState apiresp.SyncState) (SyncState, error) {
	switch apiSyncState {
	case apiresp.SyncStateScheduled:
		return SyncStateScheduled, nil
	case apiresp.SyncStateRescheduled:
		return SyncStateRescheduled, nil
	case apiresp.SyncStatePaused:
		return SyncStatePaused, nil
	case apiresp.SyncStateSyncing:
		return SyncStateSyncing, nil
	default:
		return SyncState(""), fmt.Errorf("illegal API Sync State: %q", apiSyncState)
	}
}

func convertUpdateState(apiUpdateState apiresp.UpdateState) (UpdateState, error) {
	switch apiUpdateState {
	case apiresp.UpdateStateDelayed:
		return UpdateStateDelayed, nil
	case apiresp.UpdateStateOnSchedule:
		return UpdateStateOnSchedule, nil
	default:
		return UpdateState(""), fmt.Errorf("illegal API Update State: %q", apiUpdateState)
	}
}
