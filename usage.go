package metering

import (
	"encoding/json"
	"fmt"
)

type UsageClient struct {
	BaseClient
}

type AggregationType string

const (
	Sum AggregationType = "SUM"
	Min AggregationType = "MIN"
	Max AggregationType = "MAX"
)

type UsagePayload struct {
	MeterApiName         string              `json:"meterApiName"`
	Aggregation          AggregationType     `json:"aggregation"`
	TimeGroupingInterval AggregationInterval `json:"timeGroupingInterval"`
	GroupBy              []string            `json:"groupBy,omitempty"`
	TimeRange            *TimeRange          `json:"timeRange"`
	Take                 *Take               `json:"take,omitempty"`
	Filter               map[string][]string `json:"filter,omitempty"`
}

type MeterAggregationMetadata struct {
	MeterApiName         string              `json:"meterApiName"`
	Aggregation          AggregationType     `json:"aggregation"`
	TimeGroupingInterval AggregationInterval `json:"timeGroupingInterval"`
	GroupBy              []string            `json:"groupBy,omitempty"`
	TimeRange            *TimeRange          `json:"timeRange"`
	Take                 *Take               `json:"take,omitempty"`
	Filter               map[string][]string `json:"filter,omitempty"`
}

type GroupInfo struct {
	GroupInfo map[string]string `json:"groupInfo"`
}

type DetailedAggregationValue struct {
	PercentageFromPrevious float64 `json:"percentageFromPrevious"`
	Value                  float64 `json:"value"`
	SecondsSinceEpochUtc   int64   `json:"secondsSinceEpochUtc"`
}

type DetailedMeterAggregationGroup struct {
	GroupValue float64                    `json:"groupValue"`
	Group      *GroupInfo                 `json:"group"`
	Values     []DetailedAggregationValue `json:"values"`
}

type DetailedMeterAggregation struct {
	SecondsSinceEpochIntervals []int64                         `json:"secondsSinceEpochIntervals,omitempty"`
	Metadata                   *MeterAggregationMetadata       `json:"metadata,omitempty"`
	ClientMeters               []DetailedMeterAggregationGroup `json:"clientMeters,omitempty"`
}

func NewUsageClient(apiKey string, opts ...ClientOption) *UsageClient {
	bc := NewBaseClient(apiKey, opts...)
	u := &UsageClient{BaseClient: *bc}
	u.logf("Instantiating amberflo.io Usage Client")
	return u
}

func (u *UsageClient) GetUsageAsJson(payload *UsagePayload) (*string, error) {
	url := fmt.Sprintf("%s/usage", Endpoint)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %s", err)
	}

	u.logf("Usage Payload %s", string(b))
	apiName := "Usage"
	body, err := u.AmberfloHttpClient.sendHttpRequest(apiName, url, "POST", b)
	if err != nil {
		u.logf("API error: %s", err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	v := string(body)
	return &v, nil
}

func (u *UsageClient) GetUsage(payload *UsagePayload) (*DetailedMeterAggregation, error) {
	usageResult, err := u.GetUsageAsJson(payload)

	if err != nil {
		u.logf("Usage API error: %s", err)
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	var result DetailedMeterAggregation
	json.Unmarshal([]byte(*usageResult), &result)

	return &result, nil
}
