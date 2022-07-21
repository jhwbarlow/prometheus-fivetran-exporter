package connector

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/jsonhttp"
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp/connector"
	"go.uber.org/zap"
)

type Lister interface {
	List() ([]*Connector, error)
	GetGroupID() string
	GetGroupName() string
}

type APILister struct {
	GroupID      string
	GroupName    string
	logger       *zap.SugaredLogger
	unmarshaller *jsonhttp.JSONHTTPUnmarshaller[*apiresp.ListConnectorsResp]
}

func NewAPILister(logger *zap.SugaredLogger,
	APIKey, APISecret, APIURL, groupID, groupName string,
	timeout time.Duration) (*APILister, error) {
	logger = getComponentLogger(logger, "api_lister")

	url, err := url.Parse(fmt.Sprintf("%s/v1/groups/%s/connectors?limit=1000", APIURL, groupID))
	if err != nil {
		logger.Errorw("parsing API URL", "url", url, "error", err)
		return nil, fmt.Errorf("parsing API URL %q: %w", APIURL, err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	unmarshaller := jsonhttp.NewJSONHTTPUnmarshaller[*apiresp.ListConnectorsResp](logger,
		url,
		apiToken,
		httpClient)

	return &APILister{
		GroupID:      groupID,
		GroupName:    groupName,
		logger:       logger,
		unmarshaller: unmarshaller,
	}, nil
}

func (l *APILister) List() ([]*Connector, error) {
	listConnectorsResp, err := l.unmarshaller.UnmarshallJSONFromHTTPGet()
	if err != nil {
		l.logger.Errorw("getting JSON HTTP response", "error", err)
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	connectors := make([]*Connector, 0, len(listConnectorsResp.Data.Items))
	for _, item := range listConnectorsResp.Data.Items {
		id := item.ID
		name := item.Schema
		groupID := l.GroupID
		groupName := l.GroupName

		setupState, err := l.convertSetupState(item.Status.SetupState)
		if err != nil {
			l.logger.Errorw("converting Setup State",
				"id", id,
				"name", name,
				"group_id", groupID,
				"group_name", groupName,
				"setup_state", item.Status.SetupState,
				"error", err)
			return nil, fmt.Errorf("converting Setup State: %w", err)
		}

		syncState, err := l.convertSyncState(item.Status.SyncState)
		if err != nil {
			l.logger.Errorw("converting Sync State",
				"id", id,
				"name", name,
				"group_id", groupID,
				"group_name", groupName,
				"sync_state", item.Status.SyncState,
				"error", err)
			return nil, fmt.Errorf("converting Sync State: %w", err)
		}

		updateState, err := l.convertUpdateState(item.Status.UpdateState)
		if err != nil {
			l.logger.Errorw("converting Update State",
				"id", id,
				"name", name,
				"group_id", groupID,
				"group_name", groupName,
				"update_state", item.Status.UpdateState,
				"error", err)
			return nil, fmt.Errorf("converting Update State: %w", err)
		}

		group := &Connector{
			ID:                id,
			Name:              name,
			GroupID:           groupID,
			GroupName:         groupName,
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

		l.logger.Infow("discovered connector",
			"id", id,
			"name", name,
			"group_id", groupID,
			"group_name", groupName)
		connectors = append(connectors, group)
	}

	l.logger.Infow("listed connectors from API",
		"group_id", l.GroupID,
		"group_name", l.GroupName,
		"count", len(connectors))
	return connectors, nil
}

func (l *APILister) GetGroupID() string {
	return l.GroupID
}

func (l *APILister) GetGroupName() string {
	return l.GroupName
}

func (l *APILister) convertSetupState(apiSetupState apiresp.SetupState) (SetupState, error) {
	switch apiSetupState {
	case apiresp.SetupStateIncomplete:
		return SetupStateIncomplete, nil
	case apiresp.SetupStateBroken:
		return SetupStateBroken, nil
	case apiresp.SetupStateConnected:
		return SetupStateConnected, nil
	default:
		l.logger.Errorw("illegal API Setup State", "setup_state", apiSetupState)
		return SetupState(""), fmt.Errorf("illegal API Setup State: %q", apiSetupState)
	}
}

func (l *APILister) convertSyncState(apiSyncState apiresp.SyncState) (SyncState, error) {
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
		l.logger.Errorw("illegal API Sync State", "sync_state", apiSyncState)
		return SyncState(""), fmt.Errorf("illegal API Sync State: %q", apiSyncState)
	}
}

func (l *APILister) convertUpdateState(apiUpdateState apiresp.UpdateState) (UpdateState, error) {
	switch apiUpdateState {
	case apiresp.UpdateStateDelayed:
		return UpdateStateDelayed, nil
	case apiresp.UpdateStateOnSchedule:
		return UpdateStateOnSchedule, nil
	default:
		l.logger.Errorw("illegal API Update State", "update_state", apiUpdateState)
		return UpdateState(""), fmt.Errorf("illegal API Update State: %q", apiUpdateState)
	}
}
