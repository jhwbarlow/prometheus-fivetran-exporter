package group

import (
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp"
)

type ListGroupsResp struct {
	Code apiresp.ResponseCode
	Data ListGroupsRespData
}

func (r *ListGroupsResp) GetCode() apiresp.ResponseCode {
	return r.Code
}

type ListGroupsRespData struct {
	Items []ListGroupsRespDataItem
}

type ListGroupsRespDataItem struct {
	ID   string
	Name string
}
