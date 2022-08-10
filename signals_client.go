package metering

import (
	"encoding/json"
	"errors"
	"fmt"
)

type SignalsClient struct {
	BaseClient
}

type NotificationType string

const (
	Usage            NotificationType = "usage"
	Cost             NotificationType = "cost"
	Invoice          NotificationType = "invoice"
	Prepaid          NotificationType = "prepaid"
	ProductItemUnits NotificationType = "product-item-units"
	ProductItemPrice NotificationType = "product-item-price"
)

type ThresholdCondition string

const (
	LessThan      ThresholdCondition = "less-than"
	GreaterThan   ThresholdCondition = "greater-than"
	PercentChange ThresholdCondition = "percent-change"
)

type CustomerFilterMode string

const (
	AllCustomers     CustomerFilterMode = "*"
	PerCustomer      CustomerFilterMode = "per-customer"
	SpecificCustomer CustomerFilterMode = "specific-customer"
)

func NewSignalsClient(apiKey string, opts ...ClientOption) *SignalsClient {
	bc := NewBaseClient(apiKey, opts...)
	sc := &SignalsClient{BaseClient: *bc}
	sc.logf("Instantiating amberflo.io Signals Client")
	return sc
}

type Notification struct {
	Id                 string              `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description,omitempty"`
	NotificationType   NotificationType    `json:"notificationType"`
	MeterId            string              `json:"meterId,omitempty"`
	ThresholdCondition ThresholdCondition  `json:"thresholdCondition,omitempty"`
	ThresholdValue     string              `json:"thresholdValue"`
	Range              AggregationInterval `json:"range,omitempty"`
	Cron               string              `json:"cron,omitempty"`
	Email              []string            `json:"email,omitempty"`
	WebhookUrl         string              `json:"webhookUrl,omitempty"`
	WebhookHeaders     string              `json:"webhookHeaders,omitempty"`
	WebhookPayload     string              `json:"webhookPayload,omitempty"`
	CustomerFilterMode CustomerFilterMode  `json:"customerFilterMode"`
	CustomerId         string              `json:"customerId,omitempty"`
	ProductPlanId      string              `json:"productPlanId,omitempty"`
	Enabled            bool                `json:"enabled"`
	CreateTime         int64               `json:"createTime,omitempty"`
	UpdateTime         int64               `json:"updateTime,omitempty"`
}

func (sc *SignalsClient) CreateSignal(notification *Notification) (*Notification, error) {
	signature := fmt.Sprintf("CreateSignal(%v): ", notification)
	url := fmt.Sprintf("%s/notification", Endpoint)
	return sc.wrapSignalRequest(signature, url, "POST", notification)
}

func (sc *SignalsClient) UpdateSignal(notification *Notification) (*Notification, error) {
	signature := fmt.Sprintf("UpdateSignal(%v): ", notification)

	if notification.Id == "" {
		return nil, fmt.Errorf("%s: %s", signature, errors.New("'Id' is required"))
	}

	url := fmt.Sprintf("%s/notification", Endpoint)
	return sc.wrapSignalRequest(signature, url, "PUT", notification)
}

func (sc *SignalsClient) GetSignal(notificationId string) (*Notification, error) {
	signature := fmt.Sprintf("GetSignal(%s): ", notificationId)

	if notificationId == "" {
		return nil, fmt.Errorf("%s: %s", signature, errors.New("'notificationId' is required"))
	}

	url := fmt.Sprintf("%s/notification/%s", Endpoint, notificationId)
	return sc.wrapSignalRequest(signature, url, "GET", nil)
}

func (sc *SignalsClient) DeleteSignal(notificationId string) (*Notification, error) {
	signature := fmt.Sprintf("DeleteSignal(%s): ", notificationId)

	if notificationId == "" {
		return nil, fmt.Errorf("%s: %s", signature, errors.New("'notificationId' is required"))
	}

	url := fmt.Sprintf("%s/notification/%s", Endpoint, notificationId)
	return sc.wrapSignalRequest(signature, url, "DELETE", nil)
}

func (sc *SignalsClient) wrapSignalRequest(signature string, url string, httpMethod string, notification *Notification) (*Notification, error) {
	var bytes []byte
	var err error

	//serialize payload
	if notification != nil {
		bytes, err = json.Marshal(notification)
		if err != nil {
			return nil, fmt.Errorf("%s error marshalling payload: %s", signature, err)
		}
	}

	//call API
	body, err := sc.AmberfloHttpClient.sendHttpRequest("Signals", url, httpMethod, bytes)
	if err != nil {
		sc.logf("%s API error: %s", signature, err)
		return nil, fmt.Errorf("API error: %s", err)
	}

	//deserialize API result
	err = json.Unmarshal(body, &notification)
	if err != nil {
		return nil, fmt.Errorf("%s Error reading JSON body: %s", signature, err)
	}

	return notification, nil
}
