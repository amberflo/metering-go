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
	dimensions["tenant_type"] = "Tech"

	for i := 0; i < 50; i++ {
		utcMillis := time.Now().UnixNano() / int64(time.Millisecond)
		//Queue meter messages for ingestion.
		//Queue will be flushed asyncrhonously when Metering.BatchSize is exceeded
		//or periodically at Metering.IntervalSeconds
		meteringError := Metering.Meter(&metering.MeterMessage{
			MeterName:     "ApiCalls-From-Go",
			CustomerId:    "1234",
			CustomerName:  "Dell",
			MeterValue:    int64(i) + 234,
			UtcTimeMillis: utcMillis,
			Dimensions:    dimensions,
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

	//OPTION 1: get usage by meter name. Synchronous call.
	//This will filter usage data by meter name and customer
	usageResult, err := UsageClient.GetUsage(&metering.UsagePayload{
		MeterName:    "ApiCalls-From-Go",
		CustomerName: "Dell",
	})
	fmt.Println("Usage by meter name")
	printUsageData(usageResult, err)

	//OPTION 2: get usage data by meter id. Synchronous call.
	//This will filter the usage by meter id and customer
	usageResult, err = UsageClient.GetUsage(&metering.UsagePayload{
		MeterId:      "my-meter-id",
		CustomerName: "Dell",
	})
	fmt.Println("Usage by meter id")
	printUsageData(usageResult, err)
}

func printUsageData(usageResult []metering.UsageResult, err error) {
	if err != nil {
		fmt.Println("Usage error: ", err)
		return
	}
	if len(usageResult) < 1 {
		fmt.Println("No usage data found")
		return
	}
	for _, usage := range usageResult {
		fmt.Printf("%+v\n", usage)
	}
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

	//setup customer details
	//Traits are optional. Traits can be used as filters or aggregation buckets.
	traits := make(map[string]string)
	traits["region"] = "Midwest"
	traits["tenant_type"] = "Tech"

	customerDetails := &metering.CustomerDetails{
		CustomerId:   "1234",
		CustomerName: "Dell",
		Traits:       traits,
	}
	err := Metering.AddorUpdateCustomerDetails(customerDetails)
	if err != nil {
		fmt.Println("Error creating customer details: ", err)
	}
}
```
