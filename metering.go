package metering

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jehiah/go-strftime"
	"github.com/segmentio/backo-go"
	"github.com/xtgo/uuid"
)

const Endpoint = "https://app.amberflo.io"
const RETRY_COUNT = 5
const BATCH_SIZE = 100

var Backo = backo.DefaultBacko()

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

//Metering message
type MeterMessage struct {
	UniqueId          string            `json:"uniqueId"`
	MeterApiName      string            `json:"meterApiName"`
	CustomerId        string            `json:"customerId"`
	MeterValue        float64           `json:"meterValue"`
	MeterTimeInMillis int64             `json:"meterTimeInMillis"`
	Dimensions        map[string]string `json:"dimensions,omitempty"`
	Message
}

type Customer struct {
	CustomerId    string            `json:"customerId"`
	CustomerName  string            `json:"customerName"`
	CustomerEmail string            `json:"customerEmail"`
	Traits        map[string]string `json:"traits,omitempty"`
	Enabled       bool              `json:"enabled,omitempty"`
	UpdateTime    int64             `json:"updateTime,omitempty"`
	CreateTime    int64             `json:"createTime,omitempty"`
}

// Amberflo.io metering client batches messages and flushes periodically at IntervalSeconds or
// when the BatchSize limit is exceeded.
type Metering struct {
	Endpoint string
	// IntervalSeconds is the frequency at which messages are flushed.
	IntervalSeconds time.Duration
	BatchSize       int
	Logger          *log.Logger
	Debug           bool
	Client          http.Client
	ApiKey          string

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

//Create a new instance of the metering client
func NewMeteringClient(apiKey string) *Metering {
	m := &Metering{
		Endpoint:        Endpoint,
		IntervalSeconds: 1 * time.Second,
		BatchSize:       BATCH_SIZE,
		Logger:          log.New(os.Stderr, "amberflo.io ", log.LstdFlags),
		Debug:           false,
		Client:          *http.DefaultClient,
		ApiKey:          apiKey,
		msgs:            make(chan interface{}, BATCH_SIZE),
		quit:            make(chan struct{}),
		shutdown:        make(chan struct{}),
		now:             time.Now,
		uid:             uid,
	}

	m.log("Instantiating amberflo.io metering client")

	m.upcond.L = &m.mutex
	return m
}

//AddorUpdateCustomer
func (m *Metering) AddorUpdateCustomer(customer *Customer) error {
	if customer.CustomerId == "" || customer.CustomerName == "" {
		return errors.New("customer info 'CustomerId' and 'CustomerName' are required fields")
	}

	err := m.sendCustomerToApi(customer)
	if err != nil {
		m.log("Error adding or updating customer details: %s", err)
	}
	return err
}

func (m *Metering) sendCustomerToApi(payload *Customer) error {
	signature := fmt.Sprintf("sendCustomerToApi(%v)", payload)

	m.debug("Checking if customer deatils exist %s", payload.CustomerId)
	urlGet := fmt.Sprintf("%s/customers/?customerId=%s", m.Endpoint, payload.CustomerId)
	data, err := m.sendHttpRequest("Customers", urlGet, "GET", nil)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}

	var customer *Customer
	if data != nil && string(data) != "{}" {
		err = json.Unmarshal(data, &customer)
		if err != nil {
			return fmt.Errorf("%s Error reading JSON body: %s", signature, err)
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s error marshalling payload: %s", signature, err)
	}

	url := fmt.Sprintf("%s/customers", m.Endpoint)
	httpMethod := "POST"
	if customer != nil && customer.CustomerId == payload.CustomerId {
		httpMethod = "PUT"
	}
	_, err = m.sendHttpRequest("customers", url, httpMethod, b)
	if err != nil {
		return fmt.Errorf("%s error making %s http call: %s", signature, httpMethod, err)
	}

	return nil
}

//Queue a metering message to send to Ingest API. Messages are flushes periodically at IntervalSeconds or when the BatchSize limit is exceeded.
func (m *Metering) Meter(msg *MeterMessage) error {
	currentMillis := time.Now().UnixNano()/int64(time.Millisecond) + (5 * 60 * 1000)

	if msg.MeterApiName == "" {
		return errors.New("'MeterName' is required field.")
	}
	if msg.MeterTimeInMillis < 1 {
		return errors.New("Invalid UtcTimeMillis. It should be milliseconds in UTC")
	}
	if msg.MeterTimeInMillis > currentMillis {
		return errors.New("'UtcTimeMillis' is invalid, future date not allowed")
	}

	msg.UniqueId = m.uid()
	m.log("Queuing meter message: %+v", msg)
	m.queue(msg)
	return nil
}

//Start goroutine for concurrent execution to monitor channels
func (m *Metering) startLoop() {
	go m.loop()
}

//Queue the metering message
func (m *Metering) queue(msg message) {
	m.once.Do(m.startLoop)
	msg.setMessageId(m.uid())
	msg.setTimestamp(timestamp(m.now()))
	//send message to channel
	m.msgs <- msg
}

//Flush all messages in the queue, stop the timer, close all channels, shutdown the client
func (m *Metering) Shutdown() error {
	m.log("Running shutdown....")
	m.once.Do(m.startLoop)
	//start shutdown by sending message to quit channel
	m.quit <- struct{}{}
	//close the ingest meter messages channel
	close(m.msgs)
	//receive shutdown message, blocking call
	<-m.shutdown
	m.log("Shutdown completed")
	return nil
}

//Sends batch to API asynchonrously and limits the number of concurrrent calls to API
func (m *Metering) sendAsync(msgs []interface{}) {
	m.mutex.Lock()

	//support 1000 asyncrhonus calls
	for m.counter == 1000 {
		//sleep until signal
		m.upcond.Wait()
	}
	m.counter++
	m.mutex.Unlock()
	m.wg.Add(1)

	//spin new thread to call API with retry
	go func() {
		err := m.send(msgs)
		if err != nil {
			m.log(err.Error())
		}
		m.mutex.Lock()
		m.counter--
		//signal the waiting blocked wait
		m.upcond.Signal()
		m.mutex.Unlock()
		m.wg.Done()
	}()
}

//Send the batch request with retry
func (m *Metering) send(msgs []interface{}) error {
	if len(msgs) == 0 {
		return nil
	}

	//serialize to json
	b, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("error marshalling msgs: %s", err)
	}

	//retry attempts to call Ingest API
	for i := 0; i < RETRY_COUNT; i++ {
		if i > 0 {
			m.log("Ingest Api call retry attempt: %d", i)
		}
		if err = m.ingestToApi(b); err == nil {
			return nil
		}
		m.log("Retry attempt: %d error: %s ", i, err.Error())
		Backo.Sleep(i)
	}

	return err
}

//Ingest Api Client code
func (m *Metering) ingestToApi(b []byte) error {
	m.log("Ingest API Payload %s", string(b))
	url := m.Endpoint + "/ingest"
	_, err := m.sendHttpRequest("Ingest Api", url, "POST", b)

	if err != nil {
		return fmt.Errorf("ingestToApi()=>Error calling ingest API: %s", err)
	}

	return nil
}

//Run the listener loop in a separate thread to monitor all channels
func (m *Metering) loop() {
	var msgs []interface{}
	tick := time.NewTicker(m.IntervalSeconds)
	m.log("Listener thread and timer have started")
	m.log("loop() ==> Effective batch size %d interval in seconds %d retry attempts %d", m.BatchSize, m.IntervalSeconds, RETRY_COUNT)

	for {
		//select to wait on multiple communication operations
		//blocks until one of cases can run
		select {

		//process new meter message
		case msg := <-m.msgs:
			m.debug("buffer (%d/%d) %v", len(msgs), m.BatchSize, msg)
			msgs = append(msgs, msg)
			if len(msgs) >= m.BatchSize {
				m.debug("exceeded %d messages – flushing", m.BatchSize)
				m.sendAsync(msgs)
				msgs = make([]interface{}, 0, m.BatchSize)
			}

		//timer event
		case <-tick.C:
			if len(msgs) > 0 {
				m.debug("interval reached - flushing %d", len(msgs))
				m.sendAsync(msgs)
				msgs = make([]interface{}, 0, m.BatchSize)
			} else {
				m.debug("interval reached – nothing to send")
			}

		//process shutdown
		case <-m.quit:
			//stop the timer
			tick.Stop()
			//flush the queue
			for msg := range m.msgs {
				m.debug("queue: (%d/%d) %v", len(msgs), m.BatchSize, msg)
				msgs = append(msgs, msg)
			}
			m.debug("Flushing %d messages", len(msgs))
			m.sendAsync(msgs)
			//wait for all messages to be sent to the API
			m.wg.Wait()
			m.log("Queue flushed")
			//let caller know shutdown is compelete
			m.shutdown <- struct{}{}
			return
		}
	}
}

//http client to make REST call
func (m *Metering) sendHttpRequest(apiName string, url string, httpMethod string, payload []byte) ([]byte, error) {
	signature := fmt.Sprintf("sendHttpRequest(%s, %s)", apiName, httpMethod)

	if httpMethod != "GET" {
		m.log("%s API Payload %s", signature, string(payload))
	}
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%s error creating request: %s", signature, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-KEY", m.ApiKey)

	res, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	//finally
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode < 400 {
		m.debug("%s API response: %s %s", signature, res.Status, string(body))
		return body, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	return nil, fmt.Errorf("response %s: %d – %s", res.Status, res.StatusCode, string(body))
}

func (m *Metering) debug(msg string, args ...interface{}) {
	if m.Debug {
		m.Logger.Printf(msg, args...)
	}
}

func (m *Metering) log(msg string, args ...interface{}) {
	m.Logger.Printf(msg, args...)
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
