# Amberflo Metering GO SDK

## Install SDK

In your GO code project directory, download package

```sh
go get github.com/amberflo/metering-go/v2@v2.0.1
```

## Ingesting meters
[See API Reference](https://docs.amberflo.io/reference/post_ingest)  
[Guide](https://docs.amberflo.io/docs/cloud-metering-service)
<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"fmt"
	"time"

	"github.com/amberflo/metering-go/v2"
	"github.com/xtgo/uuid"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {
	//Optional ingest options
	//Frequency at which queued data will be sent to API. Default is 1 second.
	intervalSeconds := 30 * time.Second
	//Number of messages posted to the API. Default is 100.
	batchSize := 5
	//Debug mode logging. Default is false.
	debug := true

	//Instantiate a new metering client
	meteringClient := metering.NewMeteringClient(
		apiKey,
		metering.WithBatchSize(batchSize),
		metering.WithIntervalSeconds(intervalSeconds),
		metering.WithDebug(debug),
	)

	customerId := "dell-10"

	//Define dimesions for your meters. Dimensions can be used as filters.
	dimensions := make(map[string]string)
	dimensions["region"] = "Midwest"
	dimensions["customerType"] = "Tech"

	for i := 0; i < 50; i++ {
		utcMillis := time.Now().UnixNano() / int64(time.Millisecond)
		//Queue meter messages for ingestion.
		//Queue will be flushed asyncrhonously when Metering.BatchSize is exceeded
		//or periodically at Metering.IntervalSeconds

		//unique ID is optional, but setting it
		//helps with de-dupe and revoking an ingested meter
		uniqueId := uuid.NewRandom().String()
		meteringError := meteringClient.Meter(&metering.MeterMessage{
			UniqueId:          uniqueId,
			MeterApiName:      "ApiCalls-From-Go",
			CustomerId:        customerId,
			MeterValue:        float64(i) + 234.0,
			MeterTimeInMillis: utcMillis,
			Dimensions:        dimensions,
		})
		if meteringError != nil {
			fmt.Println("Metering error: ", meteringError)
		}
		time.Sleep(500 * time.Millisecond)
	}

	//Perform graceful shutdown
	//Flush all messages in the queue, stop the timer,
	//close all channels, and shutdown the client
	meteringClient.Shutdown()
}
```
</details>

### Cancel an ingested meter
A meter can be cancelled by resending the same ingestion event and setting `metering.CancelMeter` dimension to "true".

<details>
<summary>
Sample Code
</summary>

```go
	dimensions[metering.CancelMeter] = "true"

	//cancel an ingested meter
	meteringError := meteringClient.Meter(&metering.MeterMessage{
		UniqueId:          uniqueId,
		MeterApiName:      "ApiCalls-From-Go",
		CustomerId:        customerId,
		MeterValue:        meterValue,
		MeterTimeInMillis: utcMillis,
		Dimensions:        dimensions,
	})
```
</details>

## Query usage
[See API Reference](https://docs.amberflo.io/reference/post_usage)
<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/amberflo/metering-go/v2"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {
	customerId := "dell-8"

	//initialize the usage client
	usageClient := metering.NewUsageClient(
		apiKey,
		// metering.WithCustomLogger(customerLogger),
	)

	//set the start time of the time range in Epoch seconds
	startTimeInSeconds := (time.Now().UnixNano() / int64(time.Second)) - (24 * 60 * 60)
	timeRange := &metering.TimeRange{
		StartTimeInSeconds: startTimeInSeconds,
	}

	//specify the limit and sort order
	take := &metering.Take{
		Limit:       10,
		IsAscending: true,
	}

	// Example 1: group by customers for a specific meter and all customers
	// setup usage query params
	// visit following link for description of payload:
	// https://amberflo.readme.io/reference#usage
	usageResult, err := usageClient.GetUsage(&metering.UsagePayload{
		MeterApiName:         "ApiCalls-From-Go",
		Aggregation:          metering.Sum,
		TimeGroupingInterval: metering.Day,
		GroupBy:              []string{"customerId"},
		TimeRange:            timeRange,
	})
	fmt.Println("Usage by meterApiName in json format")
	printUsageData(*usageResult, err)

	//Example 2: filter for a meter for specific customer
	//setup usage query params
	filter := make(map[string][]string)
	filter["customerId"] = []string{customerId}

	usageResult, err = usageClient.GetUsage(&metering.UsagePayload{
		MeterApiName:         "ApiCalls-From-Go",
		Aggregation:          metering.Sum,
		TimeGroupingInterval: metering.Day,
		GroupBy:              []string{"customerId"},
		TimeRange:            timeRange,
		Filter:               filter,
	})
	fmt.Println("Usage for meter for specific customer in json format")
	printUsageData(*usageResult, err)
}

func printUsageData(usageResult metering.DetailedMeterAggregation, err error) {
	if err != nil {
		fmt.Println("Usage error: ", err)
		return
	}

	jsonString, err := json.MarshalIndent(usageResult, "", "  ")
	if err != nil {
		fmt.Println("Usage error: ", err)
		return
	}

	fmt.Println(string(jsonString))
}
```
</details>

## Manage customers
[See API Reference](https://docs.amberflo.io/reference/post_customers)
<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"fmt"

	"github.com/amberflo/metering-go/v2"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {
	customerId := "dell-8"
	//Automatically create customer in Stripe
  	//and add stripeId to traits
	createCustomerInStripe := true

	//initialize the customer client
	customerClient := metering.NewCustomerClient(
		apiKey,
		//metering.WithCustomLogger(customerLogger),
	)

	//check if customer exists
	customer, err := customerClient.GetCustomer(customerId)
	if err != nil {
		fmt.Println("Error getting customer details: ", err)
	}

	//setup customer
	if customer != nil {
		//customer exists
		//update properties
		customer.CustomerName = "Dell 2"
	} else {
		//setup new customer
		//Traits are optional. Traits can be used as filters or aggregation buckets.
		traits := make(map[string]string)
		traits["region"] = "us-west"
		traits["customerType"] = "Tech"

		//In case createCustomerInStripe is false, set the trait for stripeId
		//traits[metering.StripeTraitKey] = "cus_LVxxpBQvyN3V49"

		//Set the AWS marketplace ID trait
		//traits[metering.AwsMarketPlaceTraitKey] = "aws_marketplace_id"

		customer = &metering.Customer{
			CustomerId:    customerId,
			CustomerName:  "Dell",
			CustomerEmail: "test@dell.com",
			Traits:        traits,
			Enabled:       true,
		}
	}

	customer, err = customerClient.AddorUpdateCustomer(customer, createCustomerInStripe)
	if err != nil {
		fmt.Println("Error creating customer details: ", err)
	}

	customerStatus := fmt.Sprintf("Stripe id for customer: %s", customer.Traits[metering.StripeTraitKey])
	fmt.Println(customerStatus)
}
```
</details>

## Custom logger using zerolog

By default, metering-go uses the default GO logger.

You can inject your own logger by implementing the following interface `metering.Logger`:

<details>
<summary>
Sample Code
</summary>

```go
type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
}
```

Define the custom logger:

```go
package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type CustomLogger struct {
	logger *zerolog.Logger
}

func NewCustomLogger() *CustomLogger {
	logLevel := zerolog.DebugLevel
	zerolog.SetGlobalLevel(logLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &CustomLogger{logger: &logger}
}

func (l *CustomLogger) Log(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	l.logger.Debug().Msg(msg)
}

func (l *CustomLogger) Logf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}
```
</details>

Instantiate metering clients with custom logger:

<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"github.com/amberflo/metering-go/v2"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {
	customerLogger := NewCustomLogger()

	//Instantiate a new metering client with custom logger
	meteringClient := metering.NewMeteringClient(
		apiKey,
		metering.WithLogger(customerLogger),
	)

	//initialize the usage client with custom logger
	usageClient := metering.NewUsageClient(
		apiKey,
		metering.WithCustomLogger(customerLogger),
	)

	//initialize the usage cost client with custom logger
	usageCostClient := metering.NewUsageCostClient(
		apiKey,
		metering.WithCustomLogger(customerLogger),
	)
}
```
</details>

## Query usage cost with paging
[See API Reference](https://docs.amberflo.io/reference/post_payments-cost-usage-cost)
<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/amberflo/metering-go/v2"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {
	customerId := "dell-8"

	//initialize the usage cost client
	usageCostClient := metering.NewUsageCostClient(
		apiKey,
		// metering.WithCustomLogger(customerLogger),
	)

	//set the start time of the time range in Epoch seconds
	startTimeInSeconds := (time.Now().UnixNano() / int64(time.Second)) - (24 * 60 * 60)
	timeRange := &metering.TimeRange{
		StartTimeInSeconds: startTimeInSeconds,
	}

	// paging
	pageIndex := int64(1)
	page := &metering.Page{
		Number: pageIndex,
		Size:   2,
	}

	// Example 1: group by product plan
	// setup usage cost query params
	// visit following link for description of payload:
	// https://docs.amberflo.io/reference/post_payments-cost-usage-cost
	for pageIndex < 5 {
		usageCostResultForPage, err := usageCostClient.GetUsageCost(&metering.UsageCostsKey{
			TimeGroupingInterval: metering.Day,
			GroupBy:              []string{"product_plan_id"},
			TimeRange:            timeRange,
			Page:                 page,
		})

		fmt.Println("Usage Cost Result for page: ", pageIndex)
		printUsageCostData(*usageCostResultForPage, err)

		//increment the page number
		pageIndex = pageIndex + 1
		//obtain total pages from result and stop if limit reached
		if usageCostResultForPage.PageInfo.TotalPages < pageIndex {
			break
		}

		page.Number = pageIndex
		//a token from a previous query page result to track pages and improve performance
		pageToken := usageCostResultForPage.PageInfo.PageToken
		page.Token = pageToken
	}
}

func printUsageCostData(usageCostResult metering.UsageCosts, err error) {
	if err != nil {
		fmt.Println("Usage cost error: ", err)
		return
	}

	jsonString, err := json.MarshalIndent(usageCostResult, "", "  ")
	if err != nil {
		fmt.Println("Usage cost error: ", err)
		return
	}

	fmt.Println(string(jsonString))
}
```
</details>

## Pricing plans
[See API Reference](https://docs.amberflo.io/reference/post_payments-pricing-amberflo-customer-pricing)
<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/amberflo/metering-go/v2"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {	
	customerId := "dell-8"

	//Assign pricing plan to customer
	productPlanId := "8e880691-1ae8-493b-b0a7-12a71e5dfcca"
	customerPricingClient := metering.NewCustomerPricingPlanClient(apiKey)
	customerPricingPlan, err := customerPricingClient.AddOrUpdate(&metering.CustomerProductPlan{
		ProductPlanId:      productPlanId,
		CustomerId:         customerId,
		StartTimeInSeconds: (time.Now().UnixNano() / int64(time.Second)),
	})

	if err != nil {
		fmt.Println("Error assigning customer plan: ", err)
	}
	pricingStatus := fmt.Sprintf("Customer pricing plan %s assigned to customer %s", customerPricingPlan.ProductPlanId, customerId)
	fmt.Println(pricingStatus)
}
```
</details>

## Prepaid
[See API Reference](https://docs.amberflo.io/reference/post_payments-pricing-amberflo-customer-prepaid)
<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"fmt"
	"time"

	"github.com/amberflo/metering-go/v2"
	"github.com/xtgo/uuid"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {
	customerId := "dell-1800"

	startTimeInSeconds := (time.Now().UnixNano() / int64(time.Second)) - (1 * 24 * 60 * 60)

	//*****************************************
	//*****************************************
	//Prepaid SDK
	//initialize the prepaidClient 
	prepaidClient := metering.NewPrepaidClient(
		apiKey, 
		//use a custom logger
		//metering.WithCustomLogger(customerLogger), for custom logger
	)

	recurrenceFrequency := &metering.BillingPeriod{
		Interval:       metering.DAY,
		IntervalsCount: 1,
	}

	prepaidOrder := &metering.CustomerPrepaid{
		Id:                  uuid.NewRandom().String(),
		CustomerId:          customerId,
		ExternalPayment:     true,
		StartTimeInSeconds:  startTimeInSeconds,
		PrepaidPrice:        123,
		PrepaidOfferVersion: -1,
		RecurrenceFrequency: recurrenceFrequency,
	}

	// Create a prepaid order
	prepaidOrder, err := prepaidClient.CreatePrepaidOrder(prepaidOrder)
	if err != nil {
		fmt.Println("Prepaid API error: ", err)
		return
	}

	// Get a list of all active prepaid orders
	prepaidOrders, err := prepaidClient.GetActivePrepaidOrders(customerId)
	if err != nil {
		fmt.Println("Prepaid API error: ", err)
		return
	}

	// Update the external payment status of a prepaid order
	externalPrepaidPaymentStatus := &metering.ExternalPrepaidPaymentStatus{
		PaymentStatus:        metering.SETTLED,
		SystemName:           "Stripe",
		PaymentId:            "payment-id-1",
		PaymentTimeInSeconds: (time.Now().UnixNano() / int64(time.Second)),
		PrepaidUri:           prepaidOrder.FirstInvoiceUri,
	}
	externalPrepaidPaymentStatus, err = prepaidClient.UpdateExternalPrepaidStatus(externalPrepaidPaymentStatus)

	// Delete a prepaid order
	err = prepaidClient.DeletePrepaidOrder(prepaidOrder.Id, customerId)
	if err != nil {
		fmt.Println("Prepaid API error: ", err)
		return
	}

}
```
</details>

## Signals
[See API Reference](https://docs.amberflo.io/reference/post_notifications)  
[Guide](https://docs.amberflo.io/docs/create-real-time-alerts)
<details>
<summary>
Sample Code
</summary>

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/amberflo/metering-go/v2"
)

//obtain your Amberflo API Key
var apiKey = "my-api-key"

func main() {
	signalsClient := metering.NewSignalsClient(
		apiKey,
		//use a custom logger
		//metering.WithCustomLogger(customLogger), 
	)

	invoiceAlert := &metering.Notification{
		Name:               "invoice-tracker-alert",
		NotificationType:   metering.Invoice,
		Email:              []string{"amberflo.tester@gmail.com"},
		ThresholdValue:     "200",
		CustomerFilterMode: metering.PerCustomer,
		Enabled:            true,
	}

	//Create a new signal
	invoiceAlert, err := signalsClient.CreateSignal(invoiceAlert)
	if err != nil {
		fmt.Println("API error: ", err)
		return
	}

	//update an existing signal
	invoiceAlert.ThresholdValue = "150"
	invoiceAlert.Enabled = false //disable a signal without deleting
	invoiceAlert, err = signalsClient.UpdateSignal(invoiceAlert)
	if err != nil {
		fmt.Println("API error: ", err)
		return
	}

	//get an existing signal
	invoiceAlert, err = signalsClient.GetSignal(invoiceAlert.Id)
	if err != nil {
		fmt.Println("API error: ", err)
		return
	}

	//delete a signal
	invoiceAlert, err = signalsClient.DeleteSignal(invoiceAlert.Id)
	if err != nil {
		fmt.Println("API error: ", err)
		return
	}
	fmt.Println("signal with following id deleted: ", invoiceAlert.Id)
}
```
</details>