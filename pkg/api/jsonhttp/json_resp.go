package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp"
)

type GetCoder interface {
	GetCode() apiresp.ResponseCode
}

func UnmarshallJSONFromHTTPGet[T GetCoder](url *url.URL, apiToken string, httpClient *http.Client) (T, error) {
	var genericZeroValue T

	httpReq := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL:    url,
	}
	httpReq.Header.Add("Authorization", "Basic "+apiToken)

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return genericZeroValue, fmt.Errorf("sending HTTP GET request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return genericZeroValue, fmt.Errorf("received HTTP status code %d", httpResp.StatusCode)
	}

	respBody := new(bytes.Buffer)
	if _, err := io.Copy(respBody, httpResp.Body); err != nil {
		return genericZeroValue, fmt.Errorf("copying HTTP response body: %w", err)
	}
	respBodyBytes := respBody.Bytes()
	respBodyStr := string(respBodyBytes)

	fmt.Printf("==>%#v\n", respBodyStr)

	if err := httpResp.Body.Close(); err != nil {
		return genericZeroValue, fmt.Errorf("closing HTTP response body: %w", err)
	}

	var respStruct T
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return genericZeroValue, fmt.Errorf("unmarshalling HTTP response body: %w", err)
	}

	if respStruct.GetCode() != apiresp.ResponseCodeSuccess {
		return genericZeroValue, fmt.Errorf("received response code %v", respStruct.GetCode())
	}

	return respStruct, nil
}
