package group

import (
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp"
)

/*
{
    "code": "Success",
    "data": {
        "items": [
            {
                "id": "projected_sickle",
                "name": "Staging",
                "created_at": "2018-12-20T11:59:35.089589Z"
            },
            {
                "id": "schoolmaster_heedless",
                "name": "Production",
                "created_at": "2019-01-08T19:53:52.185146Z"
            }
        ],
        "next_cursor": "eyJza2lwIjoyfQ"
    }
}
*/

type ListGroupsResp struct {
	Code resp.ResponseCode
	Data ListGroupsRespData
}

type ListGroupsRespData struct {
	Items []ListGroupsRespDataItem
}

type ListGroupsRespDataItem struct {
	ID   string
	Name string
}
