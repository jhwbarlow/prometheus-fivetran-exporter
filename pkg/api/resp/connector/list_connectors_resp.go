package connector

import (
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp"
)

type ListConnectorsResp struct {
	Code apiresp.ResponseCode
	Data ListConnectorsRespData
}

func (r *ListConnectorsResp) GetCode() apiresp.ResponseCode {
	return r.Code
}

type ListConnectorsRespData struct {
	Items []ListConnectorsRespDataItem
}

type ListConnectorsRespDataItem struct {
	ID                string
	GroupID           string `json:"group_id"`
	Service           string
	Schema            string
	Paused            bool
	SyncFrequencyMins int `json:"sync_frequency"`
	Status            Status
}

type Status struct {
	SetupState       SetupState  `json:"setup_state"`
	SyncState        SyncState   `json:"sync_state"`
	UpdateState      UpdateState `json:"update_state"`
	IsHistoricalSync bool        `json:"is_historical_sync"`
	Tasks            []Task
	Warnings         []Warning
}

type Task struct {
	Code    string
	Message string
}

type Warning struct {
	Code    string
	Message string
}
