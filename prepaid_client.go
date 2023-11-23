package metering

import (
	"encoding/json"
	"errors"
	"fmt"
)

var apiName = "Customer Prepaid"

type PrepaidClient struct {
	BaseClient
}

func NewPrepaidClient(apiKey string, opts ...ClientOption) *PrepaidClient {
	bc := NewBaseClient(apiKey, opts...)
	pc := &PrepaidClient{BaseClient: *bc}
	pc.logf("Instantiating amberflo.io Prepaid Client")
	return pc
}

type CustomerPrepaid struct {
	Id                          string           `json:"id"`
	CustomerId                  string           `json:"customerId"`
	StartTimeInSeconds          int64            `json:"startTimeInSeconds"`
	EndTimeInSeconds            int64            `json:"endTimeInSeconds,omitempty"`
	FirstCardEndTimeSeconds     int64            `json:"firstCardEndTimeSeconds"`
	ProductId                   string           `json:"productId"`
	PrepaidOfferVersion         int64            `json:"prepaidOfferVersion"`
	PrepaidPrice                float64          `json:"prepaidPrice"`
	OriginalWorth               float64          `json:"originalWorth"`
	RecurrenceFrequency         *BillingPeriod   `json:"recurrenceFrequency"`
	ExternalPayment             bool             `json:"externalPayment"`
	InternalStatus              string           `json:"internalStatus"`
	PrepaidPaymentTimeInSeconds int64            `json:"prepaidPaymentTimeInSeconds"`
	PaymentStatus               PaymentStatus    `json:"paymentStatus"`
	FirstInvoiceUri             string           `json:"firstInvoiceUri"`
	OriginalCurrency            string           `json:"originalCurrency,omitempty"`
	Label                       string           `json:"label,omitempty"`
	CustomPriority              float64          `json:"customPriority,omitempty"`
	PrepaidPriority             *PrepaidPriority `json:"prepaidPriority"`
	PaymentId                   string           `json:"paymentId,omitempty"`
	CreateTimeInSeconds         int64            `json:"createTimeInSeconds,omitempty"`
	ModifiedTimeInSeconds       int64            `json:"modifiedTimeInSeconds"`
}

type ExternalPrepaidPaymentStatus struct {
	PrepaidUri           string        `json:"prepaidUri"`
	PaymentStatus        PaymentStatus `json:"paymentStatus"`
	SystemName           string        `json:"systemName"`
	PaymentId            string        `json:"paymentId"`
	PaymentTimeInSeconds int64         `json:"paymentTimeInSeconds"`
}

func (pc *PrepaidClient) CreatePrepaidOrder(customerPrepaidOrder *CustomerPrepaid) (*CustomerPrepaid, error) {
	signature := fmt.Sprintf("CreatePrepaidOrder(%v): ", customerPrepaidOrder)

	if customerPrepaidOrder.ProductId == "" {
		customerPrepaidOrder.ProductId = "1"
	}

	bytes, err := json.Marshal(customerPrepaidOrder)
	if err != nil {
		return nil, fmt.Errorf("%s error marshalling payload: %s", signature, err)
	}

	pc.logf("%s json payload %s", signature, string(bytes))
	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-prepaid", Endpoint)
	body, err := pc.AmberfloHttpClient.sendHttpRequest(apiName, url, "POST", bytes)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	err = json.Unmarshal(body, &customerPrepaidOrder)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return customerPrepaidOrder, nil
}

func (pc *PrepaidClient) UpdateExternalPrepaidStatus(externalPrepaidPaymentStatus *ExternalPrepaidPaymentStatus) (*ExternalPrepaidPaymentStatus, error) {
	signature := fmt.Sprintf("UpdateExternalPrepaidStatus(%v): ", externalPrepaidPaymentStatus)

	paymentStatus := externalPrepaidPaymentStatus.PaymentStatus
	if paymentStatus != SETTLED && paymentStatus != FAILED && paymentStatus != PENDING {
		return nil, fmt.Errorf("%s: %s", signature, errors.New("only (SETTLED, FAILED, PENDING) allowed as 'paymentStatus'"))
	}

	bytes, err := json.Marshal(externalPrepaidPaymentStatus)
	if err != nil {
		return nil, fmt.Errorf("%s error marshalling payload: %s", signature, err)
	}

	pc.logf("%s json payload %s", signature, string(bytes))
	url := fmt.Sprintf("%s/payments/external/prepaid-payment-status", Endpoint)
	body, err := pc.AmberfloHttpClient.sendHttpRequest(apiName, url, "POST", bytes)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	err = json.Unmarshal(body, &externalPrepaidPaymentStatus)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return externalPrepaidPaymentStatus, nil
}

func (pc *PrepaidClient) GetActivePrepaidOrders(customerId string) ([]CustomerPrepaid, error) {
	signature := fmt.Sprintf("GetActivePrepaidOrders(%s): ", customerId)

	if customerId == "" {
		return nil, fmt.Errorf("%s: %s", signature, errors.New("'customerId' is required"))
	}

	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-prepaid/list?CustomerId=%s", Endpoint, customerId)
	pc.logf("%s calling API %s", signature, url)
	body, err := pc.AmberfloHttpClient.sendHttpRequest(apiName, url, "GET", nil)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	var customerPrepaidOrders []CustomerPrepaid
	err = json.Unmarshal(body, &customerPrepaidOrders)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return customerPrepaidOrders, nil
}

func (pc *PrepaidClient) DeletePrepaidOrder(id string, customerId string) error {
	signature := fmt.Sprintf("DeletePrepaidOrder(%s, %s): ", id, customerId)

	if id == "" || customerId == "" {
		return fmt.Errorf("%s: %s", signature, errors.New("'id' and 'customerId' are required"))
	}

	url := fmt.Sprintf("%s/payments/pricing/amberflo/customer-prepaid?Id=%s&CustomerId=%s", Endpoint, id, customerId)
	pc.logf("%s calling API %s", signature, url)
	_, err := pc.AmberfloHttpClient.sendHttpRequest(apiName, url, "DELETE", nil)
	if err != nil {
		pc.logf("%s API error: %s", signature, err)
		return fmt.Errorf("API error: %s", err)
	}

	return nil
}
