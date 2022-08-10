package metering

import (
	"net/http"
	"reflect"
)

type ClientOption func(*BaseClient)

func WithCustomLogger(logger Logger) ClientOption {
	return func(u *BaseClient) {
		u.Logger = logger
	}
}

type BaseClient struct {
	ApiKey             string
	Client             http.Client
	Logger             Logger
	AmberfloHttpClient AmberfloHttpClient
}

func NewBaseClient(apiKey string, opts ...ClientOption) *BaseClient {
	bc := &BaseClient{
		ApiKey: apiKey,
		Client: *http.DefaultClient,
	}

	//iterate through each option
	for _, opt := range opts {
		opt(bc)
	}

	if bc.Logger == nil {
		bc.Logger = NewAmberfloDefaultLogger()
	}

	bc.logf("instantiated the logger of type for BaseClient: %s", reflect.TypeOf(bc.Logger))
	amberfloHttpClient := NewAmberfloHttpClient(apiKey, bc.Logger, bc.Client)
	bc.AmberfloHttpClient = *amberfloHttpClient

	return bc
}

func (bc *BaseClient) logf(msg string, args ...interface{}) {
	bc.Logger.Logf(msg, args...)
}
