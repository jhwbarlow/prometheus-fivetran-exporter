package main

import (
	"encoding/json"
	"fmt"
	"time"

	connectorLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/connector"
	groupLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/group"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/connector"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/destination"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/group"
)

func main() {
	testConnector := `{
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
}`

	testDestination := `{
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
}`

	testConnectorList := `{
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
`

	testGroupsList := `{
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
}`

	respConnector := new(connector.DescribeConnectorResp)

	if err := json.Unmarshal([]byte(testConnector), respConnector); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%#v", respConnector)

	fmt.Println(respConnector.Data.Status.SetupState)

	respDestination := new(destination.DescribeDestinationResp)

	if err := json.Unmarshal([]byte(testDestination), respDestination); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%#v", respDestination)

	fmt.Println(respDestination.Data.SetupStatus)

	respConnectorList := new(connector.ListConnectorsResp)

	if err := json.Unmarshal([]byte(testConnectorList), respConnectorList); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%#v", respConnectorList)

	fmt.Println(respConnectorList.Data.Items[0].Status.SetupState)

	respGroupsList := new(group.ListGroupsResp)

	if err := json.Unmarshal([]byte(testGroupsList), respGroupsList); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%#v", respGroupsList)

	fmt.Println(respGroupsList.Data.Items[0].Name)

	fmt.Println("Testing fetch of groups from API...")
	groupLister, err := groupLister.NewAPILister("", "", "https://api.fivetran.com", 10*time.Second)
	if err != nil {
		fmt.Printf("Group lister ctor error: %v\n", err)
	}
	groups, err := groupLister.List()
	if err != nil {
		fmt.Printf("Group lister list error: %v\n", err)
	}
	for _, group := range groups {
		fmt.Printf("%#v", group)

		fmt.Println("Testing fetch of connectors in group" + group.Name + " from API...")
		connectorLister, err := connectorLister.NewAPILister("", "", "https://api.fivetran.com", group.ID, 10*time.Second)
		if err != nil {
			fmt.Printf("Connector lister ctor error: %v\n", err)
		}
		connectors, err := connectorLister.List()
		if err != nil {
			fmt.Printf("Connector lister list error: %v\n", err)
		}
		for _, connector := range connectors {
			fmt.Printf("%#v", connector)
		}
	}
}
