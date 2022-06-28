package metering

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

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

type UsageClient struct {
	ApiKey    string
	Client    http.Client
	Logger    Logger
	UsageBase UsageBase
}

func NewUsageClient(apiKey string, opts ...UsageOption) *UsageClient {
	u := &UsageClient{
		ApiKey: apiKey,
		Client: *http.DefaultClient,
	}

	u.Logger = u.UsageBase.GetLoggerInstance(opts...)
	u.logf("instantiated the logger of type: %s", reflect.TypeOf(u.Logger))
	u.logf("Instantiating amberflo.io Usage client")

	return u
}

func (u *UsageClient) GetUsageAsJson(payload *UsagePayload) (*string, error) {
	url := fmt.Sprintf("%s/usage", Endpoint)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %s", err)
	}

	u.logf("Usage Payload %s", string(b))

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-KEY", u.ApiKey)

	res, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	//finally
	defer res.Body.Close()

	u.logf("Usage API response: %s", res.Status)

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		u.logf("Usage API error: %s %s", res.Status, err)
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	if res.StatusCode >= 400 {
		u.logf("Usage API response not OK: %d %s", res.StatusCode, string(body))
		return nil, fmt.Errorf("response %s: %d – %s", res.Status, res.StatusCode, string(body))
	}

	//In case we need to return map
	// var jsonBody interface{}
	// err = json.Unmarshal(body, &jsonBody)
	// if err != nil {
	// 	u.log("Usage API error: Error decoding the response body")
	// 	return nil, fmt.Errorf("error decoding the response body")
	// }

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

func (uc *UsageClient) logf(msg string, args ...interface{}) {
	uc.Logger.Logf(msg, args...)
}
