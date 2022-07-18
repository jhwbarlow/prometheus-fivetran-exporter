package destination

import (
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/destination"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp"
)

/*
{
    "code":"Success",
    "data":{
        "id":"decent_dropsy",
        "group_id":"decent_dropsy",
        "service":"snowflake",
        "region":"GCP_US_EAST4",
        "time_zone_offset":"-5",
        "setup_status":"connected",
        "config":{
            "host":"your-account.snowflakecomputing.com",
            "port":443,
            "database":"fivetran",
            "auth":"PASSWORD",
            "user":"fivetran_user",
            "password":"******"
        }
    }
}

*/
type DescribeDestinationResp struct {
	Code resp.ResponseCode
	Data Data
}

type Data struct {
	ID          string
	GroupID     string `json:"group_id"`
	Service     string
	SetupStatus destination.SetupStatus `json:"setup_status"`
}
