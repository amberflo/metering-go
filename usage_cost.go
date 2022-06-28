package metering

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

type UsageCostsKey struct {
	ProductId            string              `json:"productId"`
	TimeRange            *TimeRange          `json:"timeRange"`
	TimeGroupingInterval AggregationInterval `json:"timeGroupingInterval"`
	Filters              map[string][]string `json:"filters,omitempty"`
	GroupBy              []string            `json:"groupBy,omitempty"`
	Take                 *Take               `json:"take,omitempty"`
}

type UsageGroupCostValue struct {
	StartTimeInSeconds       int64   `json:"startTimeInSeconds"`
	MeteredUnits             float64 `json:"meteredUnits"`
	Price                    float64 `json:"price"`
	PricePercentageDiff      float64 `json:"pricePercentageDiff"`
	MeterUnitsPercentageDiff float64 `json:"meterUnitsPercentageDiff"`
}

type UsageGroupCosts struct {
	GroupInfos   map[string]string     `json:"groupInfos"`
	MeteredUnits float64               `json:"meteredUnits"`
	Price        float64               `json:"price"`
	Costs        []UsageGroupCostValue `json:"costs"`
}

type UsageCosts struct {
	Key                        *UsageCostsKey    `json:"key,omitempty"`
	SecondsSinceEpochIntervals []int64           `json:"secondsSinceEpochIntervals,omitempty"`
	CostList                   []UsageGroupCosts `json:"costList,omitempty"`
}

type UsageCostClient struct {
	ApiKey    string
	Client    http.Client
	Logger    Logger
	UsageBase UsageBase
}

func NewUsageCostClient(apiKey string, opts ...UsageOption) *UsageCostClient {
	uc := &UsageCostClient{
		ApiKey: apiKey,
		Client: *http.DefaultClient,
	}

	uc.Logger = uc.UsageBase.GetLoggerInstance(opts...)
	uc.logf("instantiated the logger of type: %s", reflect.TypeOf(uc.Logger))
	uc.logf("Instantiating amberflo.io Usage Cost client")

	return uc
}

func (uc *UsageCostClient) GetUsageCostAsJson(payload *UsageCostsKey) (*string, error) {
	url := fmt.Sprintf("%s/payments/cost/usage-cost", Endpoint)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %s", err)
	}

	uc.logf("Usage cost payload %s", string(b))

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-KEY", uc.ApiKey)

	res, err := uc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	//finally
	defer res.Body.Close()

	uc.logf("Usage Cost API response: %s", res.Status)

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		uc.logf("Usage API error: %s %s", res.Status, err)
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	if res.StatusCode >= 400 {
		uc.logf("Usage Cost API response not OK: %d %s", res.StatusCode, string(body))
		return nil, fmt.Errorf("response %s: %d – %s", res.Status, res.StatusCode, string(body))
	}

	v := string(body)
	return &v, nil
}

func (uc *UsageCostClient) GetUsageCost(payload *UsageCostsKey) (*UsageCosts, error) {
	if payload.ProductId == "" {
		payload.ProductId = "1"
	}
	usageCostResult, err := uc.GetUsageCostAsJson(payload)

	if err != nil {
		uc.logf("Usage Cost API error: %s", err)
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	var result UsageCosts
	json.Unmarshal([]byte(*usageCostResult), &result)
	return &result, nil
}

func (uc *UsageCostClient) logf(msg string, args ...interface{}) {
	uc.Logger.Logf(msg, args...)
}
