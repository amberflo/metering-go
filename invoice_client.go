package metering

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-querystring/query"
)

type InvoiceClient struct {
	BaseClient
}

type CustomerProductItemInvoiceKey struct {
	AccountId      string `json:"accountId"`
	CustomerId     string `json:"customerId"`
	ProductId      string `json:"productId"`
	ProductPlanId  string `json:"productPlanId"`
	Year           int64  `json:"year"`
	Month          int64  `json:"month"`
	Day            int64  `json:"day"`
	ProductItemKey string `json:"productItemKey"`
}

type ItemVariantBill struct {
	PriceInCredits      float64            `json:"priceInCredits"`
	PriceInBaseCurrency float64            `json:"priceInBaseCurrency"`
	StartTimeInSeconds  int64              `json:"startTimeInSeconds"`
	EndTimeInSeconds    int64              `json:"endTimeInSeconds"`
	MeterUnits          float64            `json:"meterUnits"`
	Price               float64            `json:"price"`
	MeteredUnitsPerTier map[string]float64 `json:"meteredUnitsPerTier"`
}

type ProductItemVariantInvoice struct {
	Key               string            `json:"key"`
	ItemDimensions    map[string]string `json:"itemDimensions"`
	HourlyBillInfos   []ItemVariantBill `json:"hourlyBillInfos"`
	TotalBill         ItemVariantBill   `json:"totalBill"`
	LateArrivalMeters float64           `json:"lateArrivalMeters"`
}

type ProductItemInvoice struct {
	Key                 CustomerProductItemInvoiceKey `json:"key"`
	ProductItemId       string                        `json:"productItemId"`
	ProductItemName     string                        `json:"productItemName"`
	MeterApiName        string                        `json:"meterApiName"`
	ProductPlanName     string                        `json:"productPlanName"`
	ProductItemVariants []ProductItemVariantInvoice   `json:"productItemVariants"`
	TotalBill           ItemVariantBill               `json:"totalBill"`
}

type AppliedPromotion struct {
	PromotionId                   string  `json:"promotionId"`
	PromotionName                 string  `json:"promotionName"`
	PromotionType                 string  `json:"promotionType"`
	Discount                      float64 `json:"discount"`
	PromotionAppliedTimeInSeconds int64   `json:"promotionAppliedTimeInSeconds"`
	MaxDiscountPossible           float64 `json:"maxDiscountPossible"`
	CanBeUsedForPayAsYouGo        bool    `json:"canBeUsedForPayAsYouGo"`
	DiscountInCredits             float64 `json:"discountInCredits"`
}

type ProductPlanFee struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Cost         float64 `json:"cost"`
	IsOneTimeFee bool    `json:"isOneTimeFee"`
}

type ProductPlanBill struct {
	StartTimeInSeconds       int64   `json:"startTimeInSeconds"`
	EndTimeInSeconds         int64   `json:"endTimeInSeconds"`
	ItemPrice                float64 `json:"itemPrice"`
	FixPrice                 float64 `json:"fixPrice"`
	Prepaid                  float64 `json:"prepaid"`
	TotalDiscount            float64 `json:"totalDiscount"`
	TotalPriceBeforeDiscount float64 `json:"totalPriceBeforeDiscount"`
	TotalPriceBeforePrepaid  float64 `json:"totalPriceBeforePrepaid"`
	TotalPrice               float64 `json:"totalPrice"`
}

type CreditUnit struct {
	Id              string  `json:"id"`
	ShortName       string  `json:"shortName"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	ratioToCurrency float64 `json:"ratioToCurrency"`
}

type CustomerProductInvoice struct {
	InvoiceUri                        string                    `json:"invoiceUri"`
	InvoiceKey                        CustomerProductInvoiceKey `json:"invoiceKey"`
	PlanBillingPeriod                 BillingPeriod             `json:"planBillingPeriod"`
	PlanName                          string                    `json:"planName"`
	InvoiceStartTimeInSeconds         int64                     `json:"invoiceStartTimeInSeconds"`
	InvoiceEndTimeInSeconds           int64                     `json:"invoiceEndTimeInSeconds"`
	GracePeriodInHours                int64                     `json:"gracePeriodInHours"`
	ProductItemInvoices               []ProductItemInvoice      `json:"productItemInvoices"`
	AppliedPromotions                 []AppliedPromotion        `json:"appliedPromotions"`
	ProductPlanFees                   []ProductPlanFee          `json:"productPlanFees"`
	TotalBill                         ProductPlanBill           `json:"totalBill"`
	InvoicePriceStatus                string                    `json:"invoicePriceStatus"`
	CreditUnit                        CreditUnit                `json:"creditUnit"`
	PaymentStatus                     PaymentStatus             `json:"paymentStatus"`
	PaymentCreatedInSeconds           int64                     `json:"paymentCreatedInSeconds"`
	ExternalSystemStatus              string                    `json:"externalSystemStatus"`
	InvoiceBillInCredits              ProductPlanBill           `json:"invoiceBillInCredits"`
	AvailablePrepaidLeft              float64                   `json:"availablePrepaidLeft"`
	AvailablePrepaidLeftInCredits     float64                   `json:"availablePrepaidLeftInCredits"`
	AvailablePayAsYouGoMoney          float64                   `json:"availablePayAsYouGoMoney"`
	AvailablePayAsYouGoMoneyInCredits float64                   `json:"availablePayAsYouGoMoneyInCredits"`
}

type InvoiceKey struct {
	CustomerId         string `json:"customerId" validate:"required" url:"customerId"`
	ProductId          string `json:"productId" url:"productId"`
	FromCache          bool   `json:"fromCache" url:"fromCache"`
	WithPaymentStatus  bool   `json:"withPaymentStatus" url:"withPaymentStatus"`
}

type CustomerProductInvoiceKey struct {
	InvoiceKey
	ProductPlanId      string `json:"productPlanId" url:"productPlanId" validate:"required"`
	Year               int64  `json:"year" url:"year" validate:"required"`
	Month              int64  `json:"month" url:"month" validate:"required"`
	Day                int64  `json:"day" url:"day" validate:"required"`
}

func NewInvoiceClient(apiKey string, opts ...ClientOption) *InvoiceClient {
	bc := NewBaseClient(apiKey, opts...)
	ic := &InvoiceClient{BaseClient: *bc}
	ic.logf("Instantiating amberflo.io Customer Product Invoice Client")
	return ic
}

func (ic *InvoiceClient) GetLatestInvoice(invoiceKey *InvoiceKey) (*CustomerProductInvoice, error) {
	signature := fmt.Sprintf("GetLatestInvoice(%s): ", invoiceKey.CustomerId)
	if invoiceKey.ProductId == "" {
		invoiceKey.ProductId = "1"
	}

	if err := ic.validate(invoiceKey, signature); err != nil {
		return nil, err
	}

	queryParams, err := ic.getQueryParams(invoiceKey)
	if err != nil {
		return nil, err
	}
	queryParams = "latest=true&" + queryParams;

	body, err := ic.sendGetRequest("", queryParams, signature)
	if err != nil {
		return nil, err
	}

	var customerProductInvoice *CustomerProductInvoice;
	err = json.Unmarshal(body, &customerProductInvoice)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return customerProductInvoice, nil
}

func (ic *InvoiceClient) GetInvoice(customerProductInvoiceKey *CustomerProductInvoiceKey) (*CustomerProductInvoice, error) {
	signature := fmt.Sprintf("GetInvoice(%s): ", customerProductInvoiceKey.CustomerId)
	if customerProductInvoiceKey.ProductId == "" {
		customerProductInvoiceKey.ProductId = "1"
	}

	if err := ic.validate(customerProductInvoiceKey, signature); err != nil {
		return nil, err
	}

	queryParams, err := ic.getQueryParams(customerProductInvoiceKey)
	if err != nil {
		return nil, err
	}

	body, err := ic.sendGetRequest("", queryParams, signature)
	if err != nil {
		return nil, err
	}

	var customerProductInvoice *CustomerProductInvoice;
	err = json.Unmarshal(body, &customerProductInvoice)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return customerProductInvoice, nil
}

func (ic *InvoiceClient) ListInvoice(invoiceKey *InvoiceKey) (*[]CustomerProductInvoice, error) {
	signature := fmt.Sprintf("ListInvoice(%s): ", invoiceKey.CustomerId)
	if invoiceKey.ProductId == "" {
		invoiceKey.ProductId = "1"
	}

	if err := ic.validate(invoiceKey, signature); err != nil {
		return nil, err
	}

	queryParams, err := ic.getQueryParams(invoiceKey)
	if err != nil {
		return nil, err
	}

	body, err := ic.sendGetRequest("/all", queryParams, signature)
	if err != nil {
		return nil, err
	}

	var customerProductInvoice *[]CustomerProductInvoice;
	err = json.Unmarshal(body, &customerProductInvoice)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return customerProductInvoice, nil
}

func (ic *InvoiceClient) sendGetRequest(path string, queryParams string, signature string) ([]byte, error) {
	apiName := "Invoice"
	url := fmt.Sprintf("%s/payments/billing/customer-product-invoice%s?%s", Endpoint, path, queryParams)
	ic.logf("%s calling API %s", signature, url)
	body, err := ic.AmberfloHttpClient.sendHttpRequest(apiName, url, "GET", nil)
	if err != nil {
		ic.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}
	return body, err
}

func (ic *InvoiceClient) validate(invoiceKey interface{}, signature string) error {
	validate := validator.New()
	err := validate.Struct(invoiceKey)
	if err != nil {
		ic.logf("%s validation error: %s", signature, err)
		return errors.New("Missing required field in invoice key")
	}
	return nil
}

func (ic *InvoiceClient) getQueryParams(invoiceKey interface{}) (string, error) {
	params, err := query.Values(invoiceKey)
	if err != nil {
		ic.logf("Invoice API error: %s", err)
		return "", errors.New("Error parsing invoice key")
	}
	queryParams := params.Encode()
	return queryParams, nil
}
