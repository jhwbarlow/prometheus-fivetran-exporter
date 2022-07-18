package connector

import (
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp"
)

/*
{
    "code": "Success",
    "data": {
        "items": [
            {
                "id": "iodize_impressive",
                "group_id": "projected_sickle",
                "service": "salesforce",
                "service_version": 1,
                "schema": "salesforce",
                "connected_by": "concerning_batch",
                "created_at": "2018-07-21T22:55:21.724201Z",
                "succeeded_at": "2018-12-26T17:58:18.245Z",
                "failed_at": "2018-08-24T15:24:58.872491Z",
                "sync_frequency": 60,
                "status": {
                    "setup_state": "connected",
                    "sync_state": "paused",
                    "update_state": "delayed",
                    "is_historical_sync": false,
                    "tasks": [],
                    "warnings": []
                }
            }
        ],
        "next_cursor": "eyJza2lwIjoxfQ"
    }
}

*/

type ListConnectorsResp struct {
	Code resp.ResponseCode
	Data ListConnectorsRespData
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
