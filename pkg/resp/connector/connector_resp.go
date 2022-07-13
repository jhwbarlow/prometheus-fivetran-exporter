package connector

import (
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp"
)

/*
{
    "code": "Success",
    "data": {
        "id": "connector_id",
        "group_id": "group_id",
        "service": "adwords",
        "service_version": 4,
        "schema": "adwords.schema",
        "paused": true,
        "pause_after_trial": true,
        "connected_by": "monitoring_assuring",
        "created_at": "2020-03-11T15:03:55.743708Z",
        "succeeded_at": "2020-03-17T12:31:40.870504Z",
        "failed_at": "2021-01-15T10:55:00.056497Z",
        "sync_frequency": 360,
        "schedule_type": "auto",
        "status": {
            "setup_state": "broken",
            "sync_state": "scheduled",
            "update_state": "delayed",
            "is_historical_sync": false,
            "tasks": [
                {
                    "code": "reconnect",
                    "message": "Reconnect"
                }
            ],
            "warnings": []
        },
        "config": {
            "sync_mode": "SpecificAccounts",
            "customer_id": "XXX-XXX-XXXX",
            "accounts": [
                "1234567890"
            ],
            "conversion_window_size": 30,
            "report_type": "AD_PERFORMANCE_REPORT",
            "fields": [
                "PolicySummary",
                "AdType",
                "Date"
            ]
        },
        "source_sync_details": {
            "accounts": [
                "1234567890"
            ]
        }
    }
}

*/
type DescribeConnectorResp struct {
	Code resp.ResponseCode
	Data DescribeConnectorRespData
}

type DescribeConnectorRespData struct {
	ID      string
	GroupID string `json:"group_id"`
    Service string
	Paused  bool
	Status  Status
}

/*
"status": {
	"setup_state": "broken",
	"sync_state": "scheduled",
	"update_state": "delayed",
	"is_historical_sync": false,
	"tasks": [
			{
					"code": "reconnect",
					"message": "Reconnect"
			}
	],
	"warnings": []
}
*/
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

type Warning interface{}
