package metering

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jehiah/go-strftime"
	"github.com/xtgo/uuid"
	"go.uber.org/zap"

	"github.com/amberflo/metering-go/v2/internal/zaplog"
)

const (
	Endpoint               = "https://app.amberflo.io"
	IngestEndpoint         = "https://ingest.amberflo.io"
	RetryCount             = 6
	BatchSize              = 100
	AwsMarketPlaceTraitKey = "awsm.customerIdentifier"
	StripeTraitKey         = "stripeId"
	CancelMeter            = "aflo.cancel_previous_resource_event"
)

// Message interface.
type message interface {
	setMessageId(string)
	setTimestamp(string)
}

type Message struct {
	MessageId string `json:"-"`
	Timestamp string `json:"-"`
	SentAt    string `json:"-"`
}

// Metering message
type MeterMessage struct {
	UniqueId          string            `json:"uniqueId"`
	MeterApiName      string            `json:"meterApiName"`
	CustomerId        string            `json:"customerId"`
	MeterValue        float64           `json:"meterValue"`
	MeterTimeInMillis int64             `json:"meterTimeInMillis"`
	Dimensions        map[string]string `json:"dimensions,omitempty"`
	Message
}

type MeteringOption func(*Metering)

func WithDebug(debug bool) MeteringOption {
	return func(m *Metering) {
		m.Debug = debug
	}
}

func WithIntervalSeconds(intervalSeconds time.Duration) MeteringOption {
	return func(m *Metering) {
		m.IntervalSeconds = intervalSeconds
	}
}

func WithBatchSize(batchSize int) MeteringOption {
	return func(m *Metering) {
		m.BatchSize = batchSize
	}
}

func WithLogger(logger Logger) MeteringOption {
	return func(m *Metering) {
		m.Logger = logger
	}
}

func NewZapLogger(l *zap.SugaredLogger) *zaplog.ZapLogger {
	return zaplog.NewZapLogger(l)
}

// Amberflo.io metering client batches messages and flushes periodically at IntervalSeconds or
// when the BatchSize limit is exceeded.
type Metering struct {
	Endpoint string
	// IntervalSeconds is the frequency at which messages are flushed.
	IntervalSeconds    time.Duration
	BatchSize          int
	Logger             Logger
	Debug              bool
	Client             http.Client
	ApiKey             string
	AmberfloHttpClient AmberfloHttpClient

	// channels
	msgs     chan interface{}
	quit     chan struct{}
	shutdown chan struct{}

	// helper functions
	uid func() string
	now func() time.Time

	// Synch primitives to control number of concurrent calls to API
	once    sync.Once
	wg      sync.WaitGroup
	mutex   sync.Mutex
	upcond  sync.Cond
	counter int
}

// Create a new instance with a custom logger
func NewMeteringClient(apiKey string, opts ...MeteringOption) *Metering {
	m := &Metering{
		Endpoint:        IngestEndpoint,
		IntervalSeconds: 1 * time.Second,
		BatchSize:       BatchSize,
		Debug:           false,
		Client:          *http.DefaultClient,
		ApiKey:          apiKey,
		msgs:            make(chan interface{}, BatchSize),
		quit:            make(chan struct{}),
		shutdown:        make(chan struct{}),
		now:             time.Now,
		uid:             uid,
	}

	// iterate through each option
	for _, opt := range opts {
		opt(m)
	}

	if m.Logger == nil {
		m.Logger = NewAmberfloDefaultLogger()
		m.log("instantiated the default logger")
	}

	amberfloHttpClient := NewAmberfloHttpClient(apiKey, m.Logger, m.Client)
	m.AmberfloHttpClient = *amberfloHttpClient

	m.log("instantiating amberflo.io metering client")
	m.upcond.L = &m.mutex
	return m
}

// Queue a metering message to send to Ingest API. Messages are flushes periodically at IntervalSeconds or when the BatchSize limit is exceeded.
func (m *Metering) Meter(msg *MeterMessage) error {
	if msg.MeterApiName == "" {
		return errors.New("'MeterName' is required field")
	}
	if msg.MeterTimeInMillis < 1 {
		return errors.New("invalid UtcTimeMillis: should be milliseconds in UTC")
	}

	if strings.Trim(msg.UniqueId, " ") == "" {
		msg.UniqueId = m.uid()
	}
	m.logf("Queuing meter message: %+v", msg)
	m.queue(msg)
	return nil
}

// Start goroutine for concurrent execution to monitor channels
func (m *Metering) startLoop() {
	go m.loop()
}

// Queue the metering message
func (m *Metering) queue(msg message) {
	m.once.Do(m.startLoop)
	msg.setMessageId(m.uid())
	msg.setTimestamp(timestamp(m.now()))
	// send message to channel
	m.msgs <- msg
}

// Flush all messages in the queue, stop the timer, close all channels, shutdown the client
func (m *Metering) Shutdown() error {
	m.log("Running shutdown....")
	m.once.Do(m.startLoop)
	// start shutdown by sending message to quit channel
	m.quit <- struct{}{}
	// close the ingest meter messages channel
	close(m.msgs)
	// receive shutdown message, blocking call
	<-m.shutdown
	m.log("Shutdown completed")
	return nil
}

// Sends batch to API asynchonrously and limits the number of concurrrent calls to API
func (m *Metering) sendAsync(msgs []interface{}) {
	m.mutex.Lock()

	// support 1000 asyncrhonus calls
	for m.counter == 1000 {
		// sleep until signal
		m.upcond.Wait()
	}
	m.counter++
	m.mutex.Unlock()
	m.wg.Add(1)

	// spin new thread to call API with retry
	go func() {
		err := m.send(msgs)
		if err != nil {
			m.log(err.Error())
		}
		m.mutex.Lock()
		m.counter--
		// signal the waiting blocked wait
		m.upcond.Signal()
		m.mutex.Unlock()
		m.wg.Done()
	}()
}

// Send the batch request with retry
func (m *Metering) send(msgs []interface{}) error {
	if len(msgs) == 0 {
		return nil
	}

	// serialize to json
	b, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("error marshalling msgs: %s", err)
	}

	// retry attempts to call Ingest API
	for i := 0; i <= RetryCount; i++ {
		if i > 0 {
			m.logf("Ingest Api call retry attempt: %d", i)
		}
		if err = m.ingestToApi(b); err == nil {
			return nil
		}
		m.logf("Retry attempt: %d error: %s", i, err.Error())
		time.Sleep(backoffDelay(i))
	}

	return err
}

var retryDelays = []float64{2, 6, 12, 20, 40, 80}

const oneSecond = float64(time.Second)

func backoffDelay(retryNumber int) time.Duration {
	duration := 80.0
	if retryNumber < 6 {
		duration = retryDelays[retryNumber]
	}
	return time.Duration(duration * rand.Float64() * oneSecond)
}

// Ingest Api Client code
func (m *Metering) ingestToApi(b []byte) error {
	m.logf("Ingest API Payload %s", string(b))
	url := m.Endpoint + "/ingest"
	_, err := m.AmberfloHttpClient.sendHttpRequest("Ingest Api", url, "POST", b)
	if err != nil {
		return fmt.Errorf("ingestToApi()=>Error calling ingest API: %s", err)
	}

	return nil
}

// Run the listener loop in a separate thread to monitor all channels
func (m *Metering) loop() {
	var msgs []interface{}
	tick := time.NewTicker(m.IntervalSeconds)
	m.log("Listener thread and timer have started")
	m.logf("loop() ==> Effective batch size %d interval in seconds %d retry attempts %d", m.BatchSize, m.IntervalSeconds, RetryCount)

	for {
		// select to wait on multiple communication operations
		// blocks until one of cases can run
		select {

		// process new meter message
		case msg := <-m.msgs:
			m.debugf("buffer (%d/%d) %v", len(msgs), m.BatchSize, msg)
			msgs = append(msgs, msg)
			if len(msgs) >= m.BatchSize {
				m.debugf("exceeded %d messages – flushing", m.BatchSize)
				m.sendAsync(msgs)
				msgs = make([]interface{}, 0, m.BatchSize)
			}

		// timer event
		case <-tick.C:
			if len(msgs) > 0 {
				m.debugf("interval reached - flushing %d", len(msgs))
				m.sendAsync(msgs)
				msgs = make([]interface{}, 0, m.BatchSize)
			} else {
				m.debug("interval reached – nothing to send")
			}

		// process shutdown
		case <-m.quit:
			// stop the timer
			tick.Stop()
			// flush the queue
			for msg := range m.msgs {
				m.debugf("queue: (%d/%d) %v", len(msgs), m.BatchSize, msg)
				msgs = append(msgs, msg)
			}
			m.debugf("Flushing %d messages", len(msgs))
			m.sendAsync(msgs)
			// wait for all messages to be sent to the API
			m.wg.Wait()
			m.log("Queue flushed")
			// let caller know shutdown is compelete
			m.shutdown <- struct{}{}
			return
		}
	}
}

func (m *Metering) debug(args ...interface{}) {
	if m.Debug {
		m.Logger.Log(args...)
	}
}

func (m *Metering) debugf(format string, args ...interface{}) {
	if m.Debug {
		m.Logger.Logf(format, args...)
	}
}

func (m *Metering) log(args ...interface{}) {
	m.Logger.Log(args...)
}

func (m *Metering) logf(format string, args ...interface{}) {
	m.Logger.Logf(format, args...)
}

func (m *Message) setTimestamp(s string) {
	if m.Timestamp == "" {
		m.Timestamp = s
	}
}

func (m *Message) setMessageId(s string) {
	if m.MessageId == "" {
		m.MessageId = s
	}
}

func timestamp(t time.Time) string {
	return strftime.Format("%Y-%m-%dT%H:%M:%S%z", t)
}

func uid() string {
	return uuid.NewRandom().String()
}
