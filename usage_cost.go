package metering

import (
	"encoding/json"
	"fmt"
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
	Page                 *Page               `json:"page,omitempty"`
}

type UsageGroupCostValue struct {
	StartTimeInSeconds       int64   `json:"startTimeInSeconds"`
	MeteredUnits             float64 `json:"meteredUnits"`
	Price                    float64 `json:"price"`
	PricePercentageDiff      float64 `json:"pricePercentageDiff"`
	MeterUnitsPercentageDiff float64 `json:"meterUnitsPercentageDiff"`
	PriceBeforeDiscounts     float64 `json:"priceBeforeDiscounts"`
	PrepaidUsed              float64 `json:"prepaidUsed"`
}

type UsageGroupCosts struct {
	GroupInfos           map[string]string     `json:"groupInfos"`
	MeteredUnits         float64               `json:"meteredUnits"`
	Price                float64               `json:"price"`
	PriceBeforeDiscounts float64               `json:"priceBeforeDiscounts"`
	PrepaidUsed          float64               `json:"prepaidUsed"`
	PriceMinusPrepaid    float64               `json:"priceMinusPrepaid"`
	Costs                []UsageGroupCostValue `json:"costs"`
}

type UsageCosts struct {
	Key                        *UsageCostsKey    `json:"key,omitempty"`
	SecondsSinceEpochIntervals []int64           `json:"secondsSinceEpochIntervals,omitempty"`
	CostList                   []UsageGroupCosts `json:"costList,omitempty"`
	PageInfo                   *PageInfo         `json:"pageInfo,omitempty"`
}

type UsageCostClient struct {
	ApiKey             string
	Client             http.Client
	Logger             Logger
	UsageBase          UsageBase
	AmberfloHttpClient AmberfloHttpClient
}

func NewUsageCostClient(apiKey string, opts ...UsageOption) *UsageCostClient {
	uc := &UsageCostClient{
		ApiKey: apiKey,
		Client: *http.DefaultClient,
	}

	uc.Logger = uc.UsageBase.GetLoggerInstance(opts...)
	uc.logf("instantiated the logger of type: %s", reflect.TypeOf(uc.Logger))
	uc.logf("Instantiating amberflo.io Usage Cost client")

	amberfloHttpClient := NewAmberfloHttpClient(apiKey, uc.Logger, uc.Client)
	uc.AmberfloHttpClient = *amberfloHttpClient

	return uc
}

func (uc *UsageCostClient) GetUsageCostAsJson(payload *UsageCostsKey) (*string, error) {
	url := fmt.Sprintf("%s/payments/cost/usage-cost", Endpoint)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %s", err)
	}

	uc.logf("Usage cost payload %s", string(b))
	apiName := "Usage Cost"
	body, err := uc.AmberfloHttpClient.sendHttpRequest(apiName, url, "POST", b)
	if err != nil {
		uc.logf("API error: %s", err)
		return nil, fmt.Errorf("API error: %s", err)
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
