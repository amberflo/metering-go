# Metering GO Client

## Download Client

In your GO code project directory, download package

```sh
go get github.com/amberflo/metering-go
```

## Sample ingestion code

```go
package main

import (
	"fmt"
	"time"

	"github.com/amberflo/metering-go"
	"github.com/xtgo/uuid"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"

	//Optional ingest options
	//Frequency at which queued data will be sent to API. Default is 1 second.
	intervalSeconds := 30 * time.Second
	//Number of messages posted to the API. Default is 100.
	batchSize := 5
	//Debug mode logging. Default is false.
	debug := true

	//Instantiate a new metering client
	Metering := metering.NewMeteringClient(
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
		meteringError := Metering.Meter(&metering.MeterMessage{
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
	Metering.Shutdown()
}
```

### Cancel an ingested meter

A meter can be cancelled by resending the same ingestion event and setting `metering.CancelMeter` dimension to "true".

```go
	dimensions[metering.CancelMeter] = "true"

	//cancel an ingested meter
	meteringError := Metering.Meter(&metering.MeterMessage{
		UniqueId:          uniqueId,
		MeterApiName:      "ApiCalls-From-Go",
		CustomerId:        customerId,
		MeterValue:        meterValue,
		MeterTimeInMillis: utcMillis,
		Dimensions:        dimensions,
	})
```

## Sample Usage SDK code

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/amberflo/metering-go"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"

	customerId := "dell-8"

	//initialize the usage client
	UsageClient := metering.NewUsageClient(apiKey)

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
	usageResult, err := UsageClient.GetUsage(&metering.UsagePayload{
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

	usageResult, err = UsageClient.GetUsage(&metering.UsagePayload{
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

## Sample to setup a customer

```go
package main

import (
	"fmt"

	"github.com/amberflo/metering-go"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"
	customerId := "dell-8"
	//Automatically create customer in Stripe
  	//and add stripeId to traits
	createCustomerInStripe := true

	//Instantiate a new metering client
	Metering := metering.NewMeteringClient(apiKey)

	//check if customer exists
	customer, err := Metering.GetCustomer(customerId)
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

	customer, err = Metering.AddorUpdateCustomer(customer, createCustomerInStripe)
	if err != nil {
		fmt.Println("Error creating customer details: ", err)
	}

	customerStatus := fmt.Sprintf("Stripe id for customer: %s", customer.Traits[metering.StripeTraitKey])
	fmt.Println(customerStatus)
}
```

## Sample metering with Custom Logger using zerlog

By default, metering-go uses the default GO logger.

You can inject your own logger by implementing the following interface `metering.Logger`:

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

Instantiate metering client with custom logger:

```go
package main

import (
	"github.com/amberflo/metering-go"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"
	customerLogger := NewCustomLogger()

	//Instantiate a new metering client with custom logger
	Metering := metering.NewMeteringClient(
		apiKey,
		metering.WithLogger(customerLogger),
	)

	//initialize the usage client with custom logger
	UsageClient := metering.NewUsageClient(
		apiKey,
		metering.WithUsageLogger(customerLogger),
	)

	//initialize the usage cost client with custom logger
	usageCostClient := metering.NewUsageCostClient(
		apiKey,
		metering.WithUsageLogger(customerLogger),
	)
}
```

## Sample Usage Cost SDK code

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/amberflo/metering-go"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"

	customerId := "dell-8"

	//initialize the usage cost client
	usageCostClient := metering.NewUsageCostClient(apiKey)

	//set the start time of the time range in Epoch seconds
	startTimeInSeconds := (time.Now().UnixNano() / int64(time.Second)) - (24 * 60 * 60)
	timeRange := &metering.TimeRange{
		StartTimeInSeconds: startTimeInSeconds,
	}

	//specify the limit and sort order
	takeForCost := &metering.Take{
		Limit:       10,
		IsAscending: false,
	}

	// Example 1: group by customers
	// setup usage cost query params
	// visit following link for description of payload:
	// https://docs.amberflo.io/reference/post_payments-cost-usage-cost
	usageCostResult, err := usageCostClient.GetUsageCost(&metering.UsageCostsKey{
		TimeGroupingInterval: metering.Day,
		GroupBy:              []string{"product_plan_id"},
		TimeRange:            timeRange,
		Take:                 takeForCost,
	})
	fmt.Println("Usage Cost Result")
	printUsageCostData(*usageCostResult, err)

	//Example 2: filter for a cost for specific customer
	//setup usage query params
	filterForCost := make(map[string][]string)
	filterForCost["customerId"] = []string{customerId}

	usageCostResult, err = usageCostClient.GetUsageCost(&metering.UsageCostsKey{
		TimeGroupingInterval: metering.Day,
		GroupBy:              []string{"product_plan_id"},
		TimeRange:            timeRange,
		Take:                 takeForCost,
		Filters:              filterForCost,
	})
	fmt.Println("Usage cost for specific customer")
	printUsageCostData(*usageCostResult, err)
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
