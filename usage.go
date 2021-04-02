package metering

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type UsagePayload struct {
	MeterId      string `json:"meter_id,omitempty"`
	MeterName    string `json:"meter_name,omitempty"`
	CustomerName string `json:"tenant,omitempty"`
}

type UsageResult struct {
	CustomerName   string `json:"tenant"`
	MeterName      string `json:"measure_name"`
	Date           string `json:"date"`
	OperationValue string `json:"operation_value"`
}

type UsageClient struct {
	ApiKey string
	Client http.Client
	Logger *log.Logger
}

func NewUsageClient(apiKey string) *UsageClient {
	u := &UsageClient{
		ApiKey: apiKey,
		Client: *http.DefaultClient,
		Logger: log.New(os.Stderr, "amberflo.io ", log.LstdFlags),
	}

	u.log("Instantiating amberflo.io Usage client")

	return u
}

func (u *UsageClient) GetUsage(payload *UsagePayload) ([]UsageResult, error) {
	url := fmt.Sprintf("%s/usage-endpoint", Endpoint)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %s", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-KEY", u.ApiKey)

	res, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	//finally
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		u.log("Usage API error: %s %s", res.Status, err)
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	if res.StatusCode >= 400 {
		u.log("Usage API response not OK: %d %s", res.StatusCode, string(body))
		return nil, fmt.Errorf("response %s: %d – %s", res.Status, res.StatusCode, string(body))
	}

	var result [][]UsageResult
	if body != nil {
		err = json.Unmarshal(body, &result)
		if err != nil {
			u.log("Usage API error reading from JSON: %s %s", err, string(body))
			return nil, fmt.Errorf("Error reading JSON body: %s", err)
		}
	}

	return result[0], nil
}

func (u *UsageClient) log(msg string, args ...interface{}) {
	u.Logger.Printf(msg, args...)
}
