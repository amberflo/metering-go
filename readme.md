# Metering GO Client

## Download Client
In your GO code project directory, download package 
```sh
go get github.com/amberflo/metering-go
```
## Sample Ingestion Code 
```go
package main

import (
	"fmt"
	"time"

	"github.com/amberflo/metering-go"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"
	//Instantiate a new metering client
	Metering := metering.NewMeteringClient(apiKey)
	//Optional ingest options
	//Frequency at which queued data will be sent to API. Default is 1 second.
	Metering.IntervalSeconds = 30 * time.Second
	//Number of messages posted to the API. Default is 100.
	Metering.BatchSize = 10
	//Debug mode logging. Default is false.
	Metering.Debug = true

	//Define dimesions for your meters. Dimensions can be used as filters.
	dimensions := make(map[string]string)
	dimensions["region"] = "Midwest"
	dimensions["customerType"] = "Tech"

	for i := 0; i < 50; i++ {
		utcMillis := time.Now().UnixNano() / int64(time.Millisecond)
		//Queue meter messages for ingestion.
		//Queue will be flushed asyncrhonously when Metering.BatchSize is exceeded
		//or periodically at Metering.IntervalSeconds
		meteringError := Metering.Meter(&metering.MeterMessage{
			MeterApiName:      "ApiCalls-From-Go",
			CustomerId:        "1234",
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
	//Flush all messages in the queue, stop the timer, close all channels, and shutdown the client
	Metering.Shutdown()
}
```

## Sample Usage SDK Code 
```go
package main

import (
	"fmt"

	"github.com/amberflo/metering-go"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"

	//initialize the usage client
	UsageClient := metering.NewUsageClient(apiKey)

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
	filter := make(map[string]string)
	filter["customerId"] = "1234"

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

func printUsageData(usageResult string, err error) {
	if err != nil {
		fmt.Println("Usage error: ", err)
		return
	}

	fmt.Println(usageResult)
}
```

## Sample to Setup Customer Details 
```go
package main

import (
	"fmt"

	"github.com/amberflo/metering-go"
)

func main() {
	//obtain your Amberflo API Key
	apiKey := "my-api-key"
	//Instantiate a new metering client
	Metering := metering.NewMeteringClient(apiKey)

	//setup customer
	//Traits are optional. Traits can be used as filters or aggregation buckets.
	traits := make(map[string]string)
	traits["stripeId"] = "cus_AJ6bY3VqcaLAEs"
	traits["customerType"] = "Tech"

	customer := &metering.Customer{
		CustomerId:    "1234",
		CustomerName:  "Dell",
		CustomerEmail: "test@dell.com",
		Traits:        traits,
		Enabled:       true,
	}
	err := Metering.AddorUpdateCustomer(customer)
	if err != nil {
		fmt.Println("Error creating customer details: ", err)
	}
}
```
