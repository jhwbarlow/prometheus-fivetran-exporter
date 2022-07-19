package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	connectorCollector "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/connector"
	destinationCollector "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/destination"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/config"
	destinationDescriber "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/describer/destination"
	connectorLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/connector"
	groupLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/group"
	groupResolver "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resolver/group"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/connector"
	// "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/destination"
	// "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/group"
)

const (
	apiURL = "https://api.fivetran.com"
)

func main() {
	// 	testConnector := `{
	//     "code": "Success",
	//     "data": {
	//         "id": "connector_id",
	//         "group_id": "group_id",
	//         "service": "adwords",
	//         "service_version": 4,
	//         "schema": "adwords.schema",
	//         "paused": true,
	//         "pause_after_trial": true,
	//         "connected_by": "monitoring_assuring",
	//         "created_at": "2020-03-11T15:03:55.743708Z",
	//         "succeeded_at": "2020-03-17T12:31:40.870504Z",
	//         "failed_at": "2021-01-15T10:55:00.056497Z",
	//         "sync_frequency": 360,
	//         "schedule_type": "auto",
	//         "status": {
	//             "setup_state": "broken",
	//             "sync_state": "scheduled",
	//             "update_state": "delayed",
	//             "is_historical_sync": false,
	//             "tasks": [
	//                 {
	//                     "code": "reconnect",
	//                     "message": "Reconnect"
	//                 }
	//             ],
	//             "warnings": []
	//         },
	//         "config": {
	//             "sync_mode": "SpecificAccounts",
	//             "customer_id": "XXX-XXX-XXXX",
	//             "accounts": [
	//                 "1234567890"
	//             ],
	//             "conversion_window_size": 30,
	//             "report_type": "AD_PERFORMANCE_REPORT",
	//             "fields": [
	//                 "PolicySummary",
	//                 "AdType",
	//                 "Date"
	//             ]
	//         },
	//         "source_sync_details": {
	//             "accounts": [
	//                 "1234567890"
	//             ]
	//         }
	//     }
	// }`

	// 	testDestination := `{
	// 	"code":"Success",
	// 	"data":{
	// 			"id":"decent_dropsy",
	// 			"group_id":"decent_dropsy",
	// 			"service":"snowflake",
	// 			"region":"GCP_US_EAST4",
	// 			"time_zone_offset":"-5",
	// 			"setup_status":"connected",
	// 			"config":{
	// 					"host":"your-account.snowflakecomputing.com",
	// 					"port":443,
	// 					"database":"fivetran",
	// 					"auth":"PASSWORD",
	// 					"user":"fivetran_user",
	// 					"password":"******"
	// 			}
	// 	}
	// }`

	// 	testConnectorList := `{
	//     "code": "Success",
	//     "data": {
	//         "items": [
	//             {
	//                 "id": "iodize_impressive",
	//                 "group_id": "projected_sickle",
	//                 "service": "salesforce",
	//                 "service_version": 1,
	//                 "schema": "salesforce",
	//                 "connected_by": "concerning_batch",
	//                 "created_at": "2018-07-21T22:55:21.724201Z",
	//                 "succeeded_at": "2018-12-26T17:58:18.245Z",
	//                 "failed_at": "2018-08-24T15:24:58.872491Z",
	//                 "sync_frequency": 60,
	//                 "status": {
	//                     "setup_state": "connected",
	//                     "sync_state": "paused",
	//                     "update_state": "delayed",
	//                     "is_historical_sync": false,
	//                     "tasks": [],
	//                     "warnings": []
	//                 }
	//             }
	//         ],
	//         "next_cursor": "eyJza2lwIjoxfQ"
	//     }
	// }
	// `

	// 	testGroupsList := `{
	//     "code": "Success",
	//     "data": {
	//         "items": [
	//             {
	//                 "id": "projected_sickle",
	//                 "name": "Staging",
	//                 "created_at": "2018-12-20T11:59:35.089589Z"
	//             },
	//             {
	//                 "id": "schoolmaster_heedless",
	//                 "name": "Production",
	//                 "created_at": "2019-01-08T19:53:52.185146Z"
	//             }
	//         ],
	//         "next_cursor": "eyJza2lwIjoyfQ"
	//     }
	// }`

	// respConnector := new(connector.DescribeConnectorResp)

	// if err := json.Unmarshal([]byte(testConnector), respConnector); err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Printf("%#v", respConnector)

	// fmt.Println(respConnector.Data.Status.SetupState)

	// respDestination := new(destination.DescribeDestinationResp)

	// if err := json.Unmarshal([]byte(testDestination), respDestination); err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Printf("%#v", respDestination)

	// fmt.Println(respDestination.Data.SetupStatus)

	// respConnectorList := new(connector.ListConnectorsResp)

	// if err := json.Unmarshal([]byte(testConnectorList), respConnectorList); err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Printf("%#v", respConnectorList)

	// fmt.Println(respConnectorList.Data.Items[0].Status.SetupState)

	// respGroupsList := new(group.ListGroupsResp)

	// if err := json.Unmarshal([]byte(testGroupsList), respGroupsList); err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Printf("%#v", respGroupsList)

	// fmt.Println(respGroupsList.Data.Items[0].Name)

	// ************************************* //

	// fmt.Println("Testing fetch of groups from API...")
	// groupLister, err := groupLister.NewAPILister("97ERCwD6LPSXrkfl", "LnbQldRH5J4sEzzWoBzOGTxsX957d259", "https://api.fivetran.com", 10*time.Second)
	// if err != nil {
	// 	fmt.Printf("Group lister ctor error: %v\n", err)
	// }
	// groups, err := groupLister.List()
	// if err != nil {
	// 	fmt.Printf("Group lister list error: %v\n", err)
	// }
	// for _, group := range groups {
	// 	fmt.Printf("%#v\n", group)

	// 	fmt.Println("Testing fetch of connectors in group " + group.Name + " from API...")
	// 	connectorLister, err := connectorLister.NewAPILister("97ERCwD6LPSXrkfl", "LnbQldRH5J4sEzzWoBzOGTxsX957d259", "https://api.fivetran.com", group.ID, 10*time.Second)
	// 	if err != nil {
	// 		fmt.Printf("Connector lister ctor error: %v\n", err)
	// 	}
	// 	connectors, err := connectorLister.List()
	// 	if err != nil {
	// 		fmt.Printf("Connector lister list error: %v\n", err)
	// 	}
	// 	for _, connector := range connectors {
	// 		fmt.Printf("%#v\n", connector)
	// 	}
	// }

	apiKey, apiSecret, apiTimeout, collectedGroupNames, metricsPort, err := getConfig()
	if err != nil {
		log.Fatalf("Sourcing config error: %v\n", err) // TODO: Proper error logging
	}

	// TODO: List the groups so that we can reference by name and not ID
	groupLister, err := groupLister.NewAPILister(apiKey, apiSecret, apiURL, 10*time.Second)
	if err != nil {
		log.Fatalf("Group lister constructor error: %v\n", err)
	}
	groupResolver := groupResolver.NewDynamicLookupResolver(groupLister)
	//groupResolver, err := groupResolver.NewStaticMemoryLookupResolver(groupLister)
	//if err != nil {
	//	log.Fatalf("Group resolver constructor error: %v\n", err)
	//}

	// Loop through each group listed in the config (by name) and resolve to an ID
	// TODO: By statically resolving the group ID to group Name in the constructor,
	// we will not pick up if the group name changes. To solve this we need to dynamically
	// create the error counter in each scrape (store the count locally in this object)
	// and resolve the value on each scrape, as is done with the connectors
	//

	// TODO: Change to use Group Name to match definition in constructor where it is initialised to zero.
	// However, if the resolver dynamically looks up the connector name and that errors, we can not then
	// increment for that error as we do not know the name.
	// Thinking about it, we should probably use the group ID for the error metric as it is immutable, but
	// then again, in the config we pass a list of group names, not IDs so that may be a pointless concern
	connectorListers := make([]connectorLister.Lister, 0, len(collectedGroupNames))
	destinationDescribers := make([]destinationDescriber.Describer, 0, len(collectedGroupNames))
	for _, groupName := range collectedGroupNames {
		groupID, err := groupResolver.ResolveNameToID(groupName)
		if err != nil {
			log.Fatalf("Resolving group name %q to ID error\n", groupName)
		}

		// Construct a connector lister for each listed group
		connectorLister, err := connectorLister.NewAPILister(apiKey,
			apiSecret,
			apiURL,
			groupID,
			groupName,
			apiTimeout)
		if err != nil {
			log.Fatalf("Connector lister constructor error: %v\n", err)
		}
		connectorListers = append(connectorListers, connectorLister)

		// Construct a destination describer for each listed group
		destinationDescriber, err := destinationDescriber.NewAPIDescriber(apiKey,
			apiSecret,
			apiURL,
			groupID,
			groupName,
			apiTimeout)
		if err != nil {
			log.Fatalf("Destination describer constructor error: %v\n", err)
		}
		destinationDescribers = append(destinationDescribers, destinationDescriber)
	}

	connectorCollector, err := connectorCollector.NewConnectorCollector(connectorListers)
	if err != nil {
		log.Fatalf("Connector collector constructor error: %v\n", err)
	}

	destinationCollector := destinationCollector.NewDestinationCollector(destinationDescribers)

	prometheus.MustRegister(destinationCollector)
	prometheus.MustRegister(connectorCollector)

	if err := run(metricsPort); err != nil {
		log.Fatalf("Running exporter error: %v\n", err)
	}
}

func getConfig() (apiKey string,
	apiSecret string,
	apiCallTimeout time.Duration,
	collectedGroupNames []string,
	metricsPort uint16,
	err error) {
	var configSourcer config.Sourcer = config.NewEnvVarSourcer()

	apiKey, err = configSourcer.APIKey()
	if err != nil {
		err = fmt.Errorf("getting API Key from config: %w", err)
		return
	}

	apiSecret, err = configSourcer.APISecret()
	if err != nil {
		err = fmt.Errorf("getting API Secret from config: %w", err)
		return
	}

	apiCallTimeout, err = configSourcer.APICallTimeout()
	if err != nil {
		err = fmt.Errorf("getting API call timeout from config: %w", err)
		return
	}

	collectedGroupNames, err = configSourcer.CollectedGroupNames()
	if err != nil {
		err = fmt.Errorf("getting collected group names from config: %w", err)
		return
	}

	metricsPort, err = configSourcer.MetricsPort()
	if err != nil {
		err = fmt.Errorf("getting metrics port from config: %w", err)
		return
	}

	return
}

func run(metricsPort uint16) error {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
		return fmt.Errorf("running webserver: %w", err)
	}

	// Will never get here
	return nil
}
