package metering

import (
	"encoding/json"
	"fmt"
)

type PromotionClient struct {
	BaseClient
}

func NewPromotionClient(apiKey string, opts ...ClientOption) *PromotionClient {
	bc := NewBaseClient(apiKey, opts...)
	pc := &PromotionClient{BaseClient: *bc}
	pc.logf("Instantiating amberflo.io Promotion Client")
	return pc
}

type CustomerAppliedPromotion struct {
	CustomerId           string    `json:"customerId"`
	PromotionId          string    `json:"promotionId"`
	ProductId            string    `json:"productId"`
	AppliedTimeInSeconds string    `json:"appliedTimeInSeconds"`
	AddedTimeInSeconds   string    `json:"addedTimeInSeconds"`
	RemovedTimeInSeconds string    `json:"removedTimeInSeconds"`
	RelationId           string    `json:"relationId"`
	AppliedTimeRange     TimeRange `json:"appliedTimeRange"`
	Priority             string    `json:"priority"`
}

type ApplyPromotionRequest struct {
	CustomerId  string `json:"customerId"`
	PromotionId string `json:"promotionId"`
	ProductId   string `json:"productId"`
}

type RemovePromotionRequest struct {
	CustomerId  string `json:"customerId"`
	PromotionId string `json:"promotionId"`
	RelationId  string `json:"relationId"`
}

type Promotion struct {
	Id                     string `json:"id"`
	PromotionName          string `json:"promotionName"`
	Description            string `json:"description"`
	LockingStatus          string `json:"lockingStatus"`
	LastUpdateTimeInMillis string `json:"lastUpdateTimeInMillis"`
}

func (pc *PromotionClient) ApplyPromotion(request *ApplyPromotionRequest) (*CustomerAppliedPromotion, error) {
	signature := fmt.Sprintf("ApplyPromotion(%s): ", request)

	request.ProductId = "1"

	bytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s error marshalling payload: %s", signature, err)
	}

	pc.logf("%s json payload %s", signature, string(bytes))
	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-promotions", Endpoint)
	body, err := pc.AmberfloHttpClient.sendHttpRequest("Customer Promotions", url, "POST", bytes)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	var appliedPromotion CustomerAppliedPromotion

	err = json.Unmarshal(body, &appliedPromotion)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return &appliedPromotion, nil
}

func (pc *PromotionClient) ListAppliedPromotion(customerId string) (*[]CustomerAppliedPromotion, error) {
	signature := fmt.Sprintf("ListAppliedPromotions(%s): ", customerId)

	pc.logf("%s payload %s", signature)
	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-promotions/list?ProductId=1&CustomerId=%s", Endpoint, customerId)
	body, err := pc.AmberfloHttpClient.sendHttpRequest("Customer Promotions", url, "GET", nil)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	var appliedPromotions []CustomerAppliedPromotion

	err = json.Unmarshal(body, &appliedPromotions)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return &appliedPromotions, nil
}

func (pc *PromotionClient) RemovePromotion(request *RemovePromotionRequest) (*string, error) {
	signature := fmt.Sprintf("RemovePromotion(%s): ", request)

	bytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s error marshalling payload: %s", signature, err)
	}

	pc.logf("%s json payload %s", signature, string(bytes))
	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-promotions", Endpoint)
	body, err := pc.AmberfloHttpClient.sendHttpRequest("Customer Promotions", url, "DELETE", bytes)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	bodyString := string(body)

	return &bodyString, nil
}

func (pc *PromotionClient) ListPromotions() (*[]Promotion, error) {
	signature := "ListPromotions(): "

	pc.logf("%s payload %s", signature)
	url := fmt.Sprintf("%s/payments/pricing/amberflo/account-pricing/promotions/list", Endpoint)
	body, err := pc.AmberfloHttpClient.sendHttpRequest("Promotions", url, "GET", nil)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	var promotions []Promotion

	err = json.Unmarshal(body, &promotions)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return &promotions, nil
}

func (pc *PromotionClient) GetPromotionById(id string) (*Promotion, error) {
	signature := fmt.Sprintf("GetPromotionById(%s): ", id)

	pc.logf("%s payload %s", signature)
	url := fmt.Sprintf("%s/payments/pricing/amberflo/account-pricing/promotions?id=%s", Endpoint, id)
	body, err := pc.AmberfloHttpClient.sendHttpRequest("Promotions", url, "GET", nil)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	var promotion Promotion

	err = json.Unmarshal(body, &promotion)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return &promotion, nil
}
