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
	CustomerId            string                 `json:"customerId"`
	PromotionId           string                 `json:"promotionId"`
	ProductId             string                 `json:"productId"`
	AppliedTimeInSeconds  int64                  `json:"appliedTimeInSeconds"`
	AddedTimeInSeconds    int64                  `json:"addedTimeInSeconds"`
	RemovedTimeInSeconds  int64                  `json:"removedTimeInSeconds"`
	RelationId            string                 `json:"relationId"`
	Priority              float64                `json:"priority"`
	EntityProductInvoices []EntityProductInvoice `json:"entityProductInvoices"`
	AmountLeftInCycle     float64                `json:"amountLeftInCycle"`
	TotalAmountLeft       float64                `json:"totalAmountLeft"`
}

type EntityProductInvoice struct {
    InvoiceUri           string  `json:"invoiceUri"`
    CreatedTimeInSeconds int64   `json:"createdTimeInSeconds"`
    Amount               float64 `json:"amount"`
    InvoiceAmount        float64 `json:"invoiceAmount"`
}

type ApplyPromotionRequest struct {
	CustomerId  string `json:"customerId"`
	PromotionId string `json:"promotionId"`
	ProductId   string `json:"productId"`
	AppliedTimeRange *TimeRange `json:"appliedTimeRange,omitempty"`
}

type RemovePromotionRequest struct {
	CustomerId  string `json:"customerId"`
	PromotionId string `json:"promotionId"`
	RelationId  string `json:"relationId"`
}

type PromotionRecurrenceInterval struct {
	IntervalLength      string `json:"intervalLength,omitempty"`
	IntervalLengthCount int32  `json:"intervalLengthCount,omitempty"`
	MaxRecurrence       int32  `json:"maxRecurrence,omitempty"`
}

type HistoryRequirements struct {
	Cycles int32 `json:"cycles,omitempty"`
	Months int32 `json:"months,omitempty"`
}

type Measure struct {
	Type string `json:"type,omitempty"`
}

type Promotion struct {
	Id                           string                       `json:"id"`
	Type                         string                       `json:"type"`
	PromotionName                string                       `json:"promotionName"`
	Description                  string                       `json:"description"`
	LockingStatus                string                       `json:"lockingStatus"`
	LastUpdateTimeInMillis       int64                        `json:"lastUpdateTimeInMillis"`
	DimensionConstraintMap       *map[string]string           `json:"dimensionConstraintMap,omitempty"`
	CustomerTags                 *map[string]string           `json:"customerTags,omitempty"`
	IsHidden                     *bool                        `json:"isHidden,omitempty"`
	Discount                     *float64                     `json:"discount,omitempty"`
	TotalMaxDiscount             *float64                     `json:"totalMaxDiscount,omitempty"`
	TargetProductId              *string                      `json:"targetProductId,omitempty"`
	TargetProductItemId          *string                      `json:"targetProductItemId,omitempty"`
	AcceptablePromotionTimeRange *TimeRange                   `json:"acceptablePromotionTimeRange,omitempty"`
	RecurrenceInterval           *PromotionRecurrenceInterval `json:"recurrenceInterval,omitempty"`
	TargetItemId                 *string                      `json:"targetItemId,omitempty"`
	PromotionType                *string                      `json:"promotionType,omitempty"`
	PartnerTag                   *string                      `json:"partnerTag,omitempty"`
	DiscountRatio                *float64                     `json:"discountRatio,omitempty"`
	CycleMaxDiscount             *float64                     `json:"cycleMaxDiscount,omitempty"`
	DiscountMap                  *map[string]float64          `json:"discountMap,omitempty"`
	Measure                      *Measure                     `json:"measure,omitempty"`
	PriceToDiscountMap           *map[string]float64          `json:"priceToDiscountMap,omitempty"`
	DiscountCalculationStrategy  *string                      `json:"discountCalculationStrategy,omitempty"`
	PromotionTimeLimit           *HistoryRequirements         `json:"promotionTimeLimit,omitempty"`
	AcrossBillingPeriods         *bool                        `json:"acrossBillingPeriods,omitempty"`
	TargetPlanId                 *string                      `json:"targetPlanId,omitempty"`
	// TODO condition *PromotionCondition
	// TODO promotionModel *PromotionModel
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

func (pc *PromotionClient) RemovePromotion(request *RemovePromotionRequest) error {
	signature := fmt.Sprintf("RemovePromotion(%s): ", request)

	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-promotions?CustomerId=%s&PromotionId=%s", Endpoint, request.CustomerId, request.PromotionId)
	_, err := pc.AmberfloHttpClient.sendHttpRequest("Customer Promotions", url, "DELETE", nil)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return fmt.Errorf("API error: %s", err)
	}

	return nil
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

	pc.logf("%s", signature)
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
