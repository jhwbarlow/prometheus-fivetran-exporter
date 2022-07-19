package destination

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/destination"
	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/jsonhttp"
	groupResolver "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resolver/group"
	destinationResp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resp/destination"
)

type Describer interface {
	Describe() (*destination.Destination, error)
	GroupID() string
}

type APIDescriber struct {
	GroupResolver groupResolver.Resolver

	groupID    string
	apiToken   string
	httpClient *http.Client
	url        *url.URL
}

func NewAPIDescriber(APIKey, APISecret, APIURL, groupID string,
	timeout time.Duration,
	groupResolver groupResolver.Resolver) (*APIDescriber, error) {
	url, err := url.Parse(fmt.Sprintf("%s/v1/destinations/%s", APIURL, groupID))
	if err != nil {
		return nil, fmt.Errorf("parsing API URL: %w", err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &APIDescriber{
		GroupResolver: groupResolver,
		groupID:       groupID,
		url:           url,
		apiToken:      apiToken,
		httpClient:    httpClient,
	}, nil
}

func (d *APIDescriber) Describe() (*destination.Destination, error) {
	// httpReq := &http.Request{
	// 	Header: make(http.Header),
	// 	Method: http.MethodGet,
	// 	URL:    d.url,
	// }
	// httpReq.Header.Add("Authorization", "Basic "+d.apiToken)

	// httpResp, err := d.httpClient.Do(httpReq)
	// if err != nil {
	// 	return nil, fmt.Errorf("sending HTTP GET request: %w", err)
	// }

	// if httpResp.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("received HTTP status code %d", httpResp.StatusCode)
	// }

	// respBody := new(bytes.Buffer)
	// if _, err := io.Copy(respBody, httpResp.Body); err != nil {
	// 	return nil, fmt.Errorf("copying HTTP response body: %w", err)
	// }
	// respBodyBytes := respBody.Bytes()
	// respBodyStr := string(respBodyBytes)

	// fmt.Printf("==>%#v\n", respBodyStr)

	// if err := httpResp.Body.Close(); err != nil {
	// 	return nil, fmt.Errorf("closing HTTP response body: %w", err)
	// }

	// describeDestinationResp := new(destinationResp.DescribeDestinationResp)
	// if err := json.Unmarshal(respBodyBytes, describeDestinationResp); err != nil {
	// 	return nil, fmt.Errorf("unmarshalling HTTP response body: %w", err)
	// }

	// if describeDestinationResp.Code != resp.ResponseCodeSuccess {
	// 	return nil, fmt.Errorf("received response code %v", describeDestinationResp.Code)
	// }

	describeDestinationResp, err := jsonhttp.UnmarshallJSONFromHTTPGet[*destinationResp.DescribeDestinationResp](d.url,
		d.apiToken,
		d.httpClient)
	if err != nil {
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	groupName, err := d.GroupResolver.ResolveIDToName(d.groupID)
	if err != nil {
		return nil, fmt.Errorf("resolving group ID %q to group name", d.groupID)
	}

	destination := &destination.Destination{
		ID:   describeDestinationResp.Data.ID,
		Name: groupName, // XXX: There is no way to get this info directly from the destination
		// as Fivetran does not have the concept of a group name separate from the destination name.
		// Instead, we imply this from the name of the group.
		GroupID:     d.groupID,
		GroupName:   groupName,
		Service:     describeDestinationResp.Data.Service,
		SetupStatus: describeDestinationResp.Data.SetupStatus,
	}

	return destination, nil
}

func (d *APIDescriber) GroupID() string {
	return d.groupID
}
