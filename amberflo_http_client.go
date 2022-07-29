package metering

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpParams struct {
	ApiName    string
	Url        string
	HttpMethod string
	Payload    []byte
}

type AmberfloHttpClient struct {
	ApiKey string
	Logger Logger
	Client http.Client
}

func NewAmberfloHttpClient(apiKey string, logger Logger, httpClient http.Client) *AmberfloHttpClient {
	client := &AmberfloHttpClient{
		ApiKey: apiKey,
		Client: httpClient,
		Logger: logger,
	}
	return client
}

//http client to make REST call
func (client *AmberfloHttpClient) sendHttpRequest(apiName string, url string, httpMethod string, payload []byte) ([]byte, error) {
	signature := fmt.Sprintf("sendHttpRequest(%s, %s, %s): ", apiName, httpMethod, url)

	client.logf("%s sending http request", signature)
	if httpMethod != "GET" {
		client.logf("%s API Payload %s", signature, string(payload))
	}
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%s error creating request: %s", signature, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-KEY", client.ApiKey)

	res, err := client.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	//finally
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode < 400 {
		client.logf("%s API response: %s", signature, res.Status)
		return body, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	return nil, fmt.Errorf("response %s: %d – %s", res.Status, res.StatusCode, string(body))
}

func (client *AmberfloHttpClient) logf(msg string, args ...interface{}) {
	client.Logger.Logf(msg, args...)
}
