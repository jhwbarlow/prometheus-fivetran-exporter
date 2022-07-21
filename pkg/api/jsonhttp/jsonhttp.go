package jsonhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp"
	"go.uber.org/zap"
)

type GetCoder interface {
	GetCode() apiresp.ResponseCode
}

type JSONHTTPUnmarshaller[T GetCoder] struct {
	URL        *url.URL
	APIToken   string
	HTTPClient *http.Client
	logger     *zap.SugaredLogger
}

func NewJSONHTTPUnmarshaller[T GetCoder](logger *zap.SugaredLogger,
	URL *url.URL,
	APIToken string,
	httpClient *http.Client) *JSONHTTPUnmarshaller[T] {
	logger = getComponentLogger(logger, "json-http-unmarshaller")

	return &JSONHTTPUnmarshaller[T]{
		URL:        URL,
		HTTPClient: httpClient,
		APIToken:   APIToken,
		logger:     logger,
	}
}

func (u JSONHTTPUnmarshaller[T]) UnmarshallJSONFromHTTPGet() (T, error) {
	var genericZeroValue T

	httpReq := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL:    u.URL,
	}
	httpReq.Header.Add("Authorization", "Basic "+u.APIToken)

	httpResp, err := u.HTTPClient.Do(httpReq)
	if err != nil {
		u.logger.Errorw("sending HTTP GET request", "url", u.URL, "error", err)
		return genericZeroValue, fmt.Errorf("sending HTTP GET request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		u.logger.Errorw("received unexpected HTTP status code", "url", u.URL, "status_code", httpResp.StatusCode)
		return genericZeroValue, fmt.Errorf("received unexpected HTTP status code %d", httpResp.StatusCode)
	}

	respBody := new(bytes.Buffer)
	if _, err := io.Copy(respBody, httpResp.Body); err != nil {
		u.logger.Errorw("copying HTTP response body", "url", u.URL, "error", err)
		return genericZeroValue, fmt.Errorf("copying HTTP response body: %w", err)
	}
	respBodyBytes := respBody.Bytes()
	//respBodyStr := string(respBodyBytes)
	//fmt.Printf("==>%#v\n", respBodyStr)

	if err := httpResp.Body.Close(); err != nil {
		u.logger.Errorw("closing HTTP response body", "url", u.URL, "error", err)
		return genericZeroValue, fmt.Errorf("closing HTTP response body: %w", err)
	}

	var respStruct T
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		u.logger.Errorw("unmarshalling HTTP response body", "url", u.URL, "error", err)
		return genericZeroValue, fmt.Errorf("unmarshalling HTTP response body: %w", err)
	}

	if respStruct.GetCode() != apiresp.ResponseCodeSuccess {
		u.logger.Errorw("received response code", "url", u.URL, "response_code", respStruct.GetCode())
		return genericZeroValue, fmt.Errorf("received response code %v", respStruct.GetCode())
	}

	u.logger.Infow("received and unmarshalled JSON HTTP response", "url", u.URL)
	return respStruct, nil
}
