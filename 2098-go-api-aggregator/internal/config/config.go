package config

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"api-aggregator/internal/db"
	"api-aggregator/internal/formatter"
	"api-aggregator/internal/security"
)

type EndpointSpec struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Timeout int               `json:"timeout"`
	Retries int               `json:"retries"`
}

type AggregationSpec struct {
	Endpoints    []*EndpointSpec          `json:"endpoints"`
	FormatType   formatter.FormatType     `json:"format_type"`
	FormatConfig map[string]interface{}   `json:"format_config"`
	CacheTTL     int                      `json:"cache_ttl"`
}

func ParseFromQuery(query url.Values) (*AggregationSpec, error) {
	spec := &AggregationSpec{}

	urlList := query["url"]
	methodList := query["method"]
	timeoutList := query["timeout"]
	retryList := query["retry"]

	if len(urlList) == 0 {
		return nil, fmt.Errorf("at least one 'url' parameter is required")
	}

	spec.Endpoints = make([]*EndpointSpec, 0, len(urlList))
	for i := 0; i < len(urlList); i++ {
		ep := &EndpointSpec{
			URL:    urlList[i],
			Method: "GET",
		}

		if i < len(methodList) && methodList[i] != "" {
			ep.Method = strings.ToUpper(methodList[i])
		}
		if i < len(timeoutList) {
			var t int
			fmt.Sscanf(timeoutList[i], "%d", &t)
			ep.Timeout = t
		}
		if i < len(retryList) {
			var r int
			fmt.Sscanf(retryList[i], "%d", &r)
			ep.Retries = r
		}

		spec.Endpoints = append(spec.Endpoints, ep)
	}

	spec.FormatType = formatter.FormatType(query.Get("format"))
	if spec.FormatType == "" {
		spec.FormatType = formatter.FormatMerge
	}

	if ft := query.Get("format_type"); ft != "" {
		spec.FormatType = formatter.FormatType(ft)
	}

	spec.FormatConfig = make(map[string]interface{})
	for k, v := range query {
		if strings.HasPrefix(k, "extract_") && len(v) > 0 {
			fieldName := strings.TrimPrefix(k, "extract_")
			parts := strings.SplitN(v[0], ":", 2)
			if len(parts) == 2 {
				idx := 0
				fmt.Sscanf(parts[0], "%d", &idx)
				extracts := spec.FormatConfig["extracts"]
				if extracts == nil {
					extracts = make(map[string]interface{})
					spec.FormatConfig["extracts"] = extracts
				}
				extracts.(map[string]interface{})[fieldName] = map[string]interface{}{
					"endpoint_index": idx,
					"path":           parts[1],
				}
			}
		}
	}

	var ttl int
	fmt.Sscanf(query.Get("cache_ttl"), "%d", &ttl)
	spec.CacheTTL = ttl

	if err := ValidateSpec(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func ValidateSpec(spec *AggregationSpec) error {
	if len(spec.Endpoints) == 0 {
		return fmt.Errorf("no endpoints specified")
	}

	for _, ep := range spec.Endpoints {
		if err := security.ValidateURL(ep.URL); err != nil {
			return err
		}
		if security.IsInternalAddress(ep.URL) {
			return fmt.Errorf("internal address not allowed: %s", ep.URL)
		}
		if ep.Method == "" {
			ep.Method = "GET"
		}
	}

	switch spec.FormatType {
	case formatter.FormatMerge, formatter.FormatArray, formatter.FormatResponse:
	case formatter.FormatExtract:
		if spec.FormatConfig == nil {
			return fmt.Errorf("format_config required for 'extract' format")
		}
		if _, ok := spec.FormatConfig["extracts"]; !ok {
			return fmt.Errorf("format_config must contain 'extracts'")
		}
	default:
		return fmt.Errorf("invalid format type: %s", spec.FormatType)
	}

	return nil
}

func GenerateQueryString(query url.Values) string {
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		vs := query[k]
		sort.Strings(vs)
		for _, v := range vs {
			if sb.Len() > 0 {
				sb.WriteByte('&')
			}
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	return sb.String()
}

func EndpointToSpec(ep *db.Endpoint) *EndpointSpec {
	return &EndpointSpec{
		URL:     ep.URL,
		Method:  ep.Method,
		Headers: ep.Headers,
		Timeout: ep.Timeout,
		Retries: ep.Retries,
	}
}
