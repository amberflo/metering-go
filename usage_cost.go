package metering

import (
	"encoding/json"
	"fmt"
)

type UsageCostClient struct {
	BaseClient
}

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

func NewUsageCostClient(apiKey string, opts ...ClientOption) *UsageCostClient {
	bc := NewBaseClient(apiKey, opts...)
	uc := &UsageCostClient{BaseClient: *bc}
	uc.logf("Instantiating amberflo.io Usage Cost Client")
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
