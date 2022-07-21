package group

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jsonhttp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/jsonhttp"
	apiresp "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/api/resp/group"
	"go.uber.org/zap"
)

type Lister interface {
	List() ([]*Group, error)
}

type APILister struct {
	logger       *zap.SugaredLogger
	unmarshaller *jsonhttp.JSONHTTPUnmarshaller[*apiresp.ListGroupsResp]
}

func NewAPILister(logger *zap.SugaredLogger, APIKey, APISecret, APIURL string, timeout time.Duration) (*APILister, error) {
	logger = getComponentLogger(logger, "api_lister")

	url, err := url.Parse(APIURL + "/v1/groups?limit=1000")
	if err != nil {
		logger.Errorw("parsing API URL", "url", APIURL, "error", err)
		return nil, fmt.Errorf("parsing API URL %q: %w", APIURL, err)
	}

	apiToken := base64.StdEncoding.EncodeToString([]byte(APIKey + ":" + APISecret))
	httpClient := &http.Client{
		Timeout: timeout,
	}

	unmarshaller := jsonhttp.NewJSONHTTPUnmarshaller[*apiresp.ListGroupsResp](logger,
		url,
		apiToken,
		httpClient)

	return &APILister{
		logger:       logger,
		unmarshaller: unmarshaller,
	}, nil
}

func (l *APILister) List() ([]*Group, error) {
	listGroupsResp, err := l.unmarshaller.UnmarshallJSONFromHTTPGet()
	if err != nil {
		l.logger.Errorw("getting JSON HTTP response", "error", err)
		return nil, fmt.Errorf("getting JSON HTTP response: %w", err)
	}

	groups := make([]*Group, 0, len(listGroupsResp.Data.Items))
	for _, item := range listGroupsResp.Data.Items {
		group := &Group{
			ID:   item.ID,
			Name: item.Name,
		}

		l.logger.Infow("discovered group", "id", group.ID, "name", group.Name)
		groups = append(groups, group)
	}

	l.logger.Infow("listed groups from API", "count", len(groups))
	return groups, nil
}
