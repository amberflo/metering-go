package metering

import (
	"encoding/json"
	"errors"
	"fmt"
)

type CustomerPricingPlanClient struct {
	BaseClient
}

type CustomerProductPlan struct {
	ProductId          string `json:"productId"`
	ProductPlanId      string `json:"productPlanId"`
	CustomerId         string `json:"customerId"`
	StartTimeInSeconds int64  `json:"startTimeInSeconds"`
	EndTimeInSeconds   int64  `json:"endTimeInSeconds,omitempty"`
}

func NewCustomerPricingPlanClient(apiKey string, opts ...ClientOption) *CustomerPricingPlanClient {
	bc := NewBaseClient(apiKey, opts...)
	cpc := &CustomerPricingPlanClient{BaseClient: *bc}
	cpc.logf("Instantiating amberflo.io Customer Pricing Plan Client")
	return cpc
}

func (cpc *CustomerPricingPlanClient) AddOrUpdate(payload *CustomerProductPlan) (*CustomerProductPlan, error) {
	signature := fmt.Sprintf("AddOrUpdate(%v)", payload)
	if payload.CustomerId == "" || payload.ProductPlanId == "" {
		return nil, errors.New("'CustomerId' and 'ProductPlanId' are required fields")
	}

	if payload.ProductId == "" {
		payload.ProductId = "1"
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s error marshalling payload: %s", signature, err)
	}

	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-pricing", Endpoint)
	apiName := "Customer Pricing"
	cpc.logf("Customer pricing client payload %s", string(b))
	body, err := cpc.AmberfloHttpClient.sendHttpRequest(apiName, url, "POST", b)
	if err != nil {
		cpc.logf("API error: %s", err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	v := string(body)
	var result CustomerProductPlan
	json.Unmarshal([]byte(v), &result)
	return &result, nil
}
