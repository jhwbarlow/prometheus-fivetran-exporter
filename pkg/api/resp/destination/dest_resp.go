package destination

import (
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp"
)

type DescribeDestinationResp struct {
	Code apiresp.ResponseCode
	Data Data
}

func (r *DescribeDestinationResp) GetCode() apiresp.ResponseCode {
	return r.Code
}

type Data struct {
	ID          string
	GroupID     string `json:"group_id"`
	Service     string
	SetupStatus SetupStatus `json:"setup_status"`
}
