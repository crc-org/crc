package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Version of the client.
const Version = "3.4.0"

// This interface is the main API exposed by the analytics package.
// Values that satsify this interface are returned by the client constructors
// provided by the package and provide a way to send messages via the HTTP API.
type Client interface {
	io.Closer

	// Queues a message to be sent by the client when the conditions for a batch
	// upload are met.
	// This is the main method you'll be using, a typical flow would look like
	// this:
	//
	//	client := analytics.New(writeKey)
	//	...
	//	client.Enqueue(analytics.Track{ ... })
	//	...
	//	client.Close()
	//
	// The method returns an error if the message queue not be queued, which
	// happens if the client was already closed at the time the method was
	// called or if the message was malformed.
	Enqueue(Message) error
}

type client struct {
	Config
	key string

	// This channel is where the `Enqueue` method writes messages so they can be
	// picked up and pushed by the backend goroutine taking care of applying the
	// batching rules.
	msgs chan Message

	// These two channels are used to synchronize the client shutting down when
	// `Close` is called.
	// The first channel is closed to signal the backend goroutine that it has
	// to stop, then the second one is closed by the backend goroutine to signal
	// that it has finished flushing all queued messages.
	quit     chan struct{}
	shutdown chan struct{}

	// This HTTP client is used to send requests to the backend, it uses the
	// HTTP transport provided in the configuration.
	http http.Client
}

// Instantiate a new client that uses the write key passed as first argument to
// send messages to the backend.
// The client is created with the default configuration.
func New(writeKey string) Client {
	// Here we can ignore the error because the default config is always valid.
	c, _ := NewWithConfig(writeKey, Config{})
	return c
}

// Instantiate a new client that uses the write key and configuration passed as
// arguments to send messages to the backend.
// The function will return an error if the configuration contained impossible
// values (like a negative flush interval for example).
// When the function returns an error the returned client will always be nil.
func NewWithConfig(writeKey string, config Config) (cli Client, err error) {
	if err = config.validate(); err != nil {
		return
	}

	c := &client{
		Config:   makeConfig(config),
		key:      writeKey,
		msgs:     make(chan Message, 100),
		quit:     make(chan struct{}),
		shutdown: make(chan struct{}),
		http:     makeHttpClient(config.Transport),
	}

	go c.loop()

	cli = c
	return
}

func makeHttpClient(transport http.RoundTripper) http.Client {
	httpClient := http.Client{
		Transport: transport,
	}
	if supportsTimeout(transport) {
		httpClient.Timeout = 10 * time.Second
	}
	return httpClient
}

func dereferenceMessage(msg Message) Message {
	switch m := msg.(type) {
	case *Alias:
		if m == nil {
			return nil
		}
		return *m
	case *Group:
		if m == nil {
			return nil
		}
		return *m
	case *Identify:
		if m == nil {
			return nil
		}
		return *m
	case *Page:
		if m == nil {
			return nil
		}
		return *m
	case *Screen:
		if m == nil {
			return nil
		}
		return *m
	case *Track:
		if m == nil {
			return nil
		}
		return *m
	}

	return msg
}

func (c *client) Enqueue(msg Message) (err error) {
	msg = dereferenceMessage(msg)
	if err = msg.Validate(); err != nil {
		return
	}

	var id = c.uid()
	var ts = c.now()

	switch m := msg.(type) {
	case Alias:
		m.Type = "alias"
		m.MessageId = makeMessageId(m.MessageId, id)
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		msg = m

	case Group:
		m.Type = "group"
		m.MessageId = makeMessageId(m.MessageId, id)
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		msg = m

	case Identify:
		m.Type = "identify"
		m.MessageId = makeMessageId(m.MessageId, id)
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		msg = m

	case Page:
		m.Type = "page"
		m.MessageId = makeMessageId(m.MessageId, id)
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		msg = m

	case Screen:
		m.Type = "screen"
		m.MessageId = makeMessageId(m.MessageId, id)
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		msg = m

	case Track:
		m.Type = "track"
		m.MessageId = makeMessageId(m.MessageId, id)
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		msg = m

	default:
		err = fmt.Errorf("messages with custom types cannot be enqueued: %T", msg)
		return
	}

	defer func() {
		// When the `msgs` channel is closed writing to it will trigger a panic.
		// To avoid letting the panic propagate to the caller we recover from it
		// and instead report that the client has been closed and shouldn't be
		// used anymore.
		if recover() != nil {
			err = ErrClosed
		}
	}()

	c.msgs <- msg
	return
}

// Close and flush metrics.
func (c *client) Close() (err error) {
	defer func() {
		// Always recover, a panic could be raised if `c`.quit was closed which
		// means the method was called more than once.
		if recover() != nil {
			err = ErrClosed
		}
	}()
	close(c.quit)
	<-c.shutdown
	return
}

// Asychronously send a batched requests.
func (c *client) sendAsync(msgs []message, wg *sync.WaitGroup, ex *executor) {
	wg.Add(1)

	if !ex.do(func() {
		defer wg.Done()
		defer func() {
			// In case a bug is introduced in the send function that triggers
			// a panic, we don't want this to ever crash the application so we
			// catch it here and log it instead.
			if err := recover(); err != nil {
				c.errorf("panic - %s", err)
			}
		}()
		c.send(msgs)
	}) {
		wg.Done()
		c.errorf("sending messages failed - %s", ErrTooManyRequests)
		c.notifyFailure(msgs, ErrTooManyRequests)
	}
}

// httpError is returned by report() for non-2xx/3xx responses.
type httpError struct {
	StatusCode int
	Retryable  bool
	RetryAfter int64 // seconds from Retry-After header; 0 if absent
	Body       string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, e.Body)
}

// Send batch request.
func (c *client) send(msgs []message) {
	b, err := json.Marshal(batch{
		MessageId: c.uid(),
		SentAt:    c.now(),
		Messages:  msgs,
		Context:   c.DefaultContext,
	})

	if err != nil {
		c.errorf("marshalling messages - %s", err)
		c.notifyFailure(msgs, err)
		return
	}

	retry := retryState{client: c, msgs: msgs}
	var shutdownDeadline time.Time
	for {
		retry.totalAttempts++

		uploadErr := c.upload(b, retry.totalAttempts, shutdownDeadline)
		if uploadErr == nil {
			c.notifySuccess(msgs)
			return
		}

		action := retry.classify(uploadErr)
		var delay time.Duration
		switch action {
		case retryActionDrop:
			return
		case retryActionRateLimit, retryActionBackoff:
			delay = retry.delay
		}

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-c.quit:
			// Stopped explicitly: time.After would leave the timer live until it
			// fired, and this loop can run for hours with delays up to the
			// Retry-After cap.
			timer.Stop()
			// Closing: finish the retry schedule so shutdown does not discard a
			// batch the server asked us to resend, bounded by ShutdownTimeout
			// rather than the much longer MaxRateLimitDuration.
			if shutdownDeadline.IsZero() {
				shutdownDeadline = time.Now().Add(c.ShutdownTimeout)
			}
			remaining := time.Until(shutdownDeadline)
			if remaining <= 0 {
				c.errorf("%d messages dropped because they failed to be sent and the client was closed", len(msgs))
				c.notifyFailure(msgs, uploadErr)
				return
			}
			if delay > remaining {
				delay = remaining
			}
			time.Sleep(delay)
		}
	}
}

type retryAction int

const (
	retryActionBackoff   retryAction = iota
	retryActionRateLimit retryAction = iota
	retryActionDrop      retryAction = iota
)

// retryState tracks state across attempts within a single send call.
type retryState struct {
	client             *client
	msgs               []message
	totalAttempts      int
	backoffAttempts    int
	firstFailureTime   time.Time
	rateLimitStartTime time.Time
	delay              time.Duration
}

// classify determines what to do after a failed upload. It updates internal
// counters, logs/notifies on terminal failures, and returns the action the
// caller should take.
func (r *retryState) classify(uploadErr error) retryAction {
	c := r.client

	httpErr, ok := uploadErr.(*httpError)
	if !ok {
		httpErr = &httpError{Retryable: true}
	}

	if !httpErr.Retryable {
		c.errorf("messages dropped due to non-retryable error - %s", uploadErr)
		c.notifyFailure(r.msgs, uploadErr)
		return retryActionDrop
	}

	if httpErr.RetryAfter > 0 {
		return r.handleRateLimit(httpErr)
	}

	return r.handleBackoff(uploadErr)
}

func (r *retryState) handleRateLimit(httpErr *httpError) retryAction {
	c := r.client

	now := c.now()
	if r.rateLimitStartTime.IsZero() {
		r.rateLimitStartTime = now
	}

	// One reading serves both the budget test and the clamp below. Reading the clock
	// twice lets the budget expire between them and yields a negative remaining,
	// which time.NewTimer treats as "fire immediately" rather than rejecting — so it
	// would be harmless here, but only by accident of that behaviour.
	remaining := c.MaxRateLimitDuration - now.Sub(r.rateLimitStartTime)
	if remaining <= 0 {
		c.errorf("messages dropped - %s", ErrRateLimitBudgetExceeded)
		c.notifyFailure(r.msgs, ErrRateLimitBudgetExceeded)
		return retryActionDrop
	}

	// A wait that will not fit ends the episode. Shortening it would send the next
	// request inside the window the server asked us to wait out -- one it has
	// already said it will not serve -- and since the budget is spent by then it
	// would be the last attempt either way. Giving up here loses the same batch and
	// sends one fewer request at something already rate-limiting us.
	delay := time.Duration(httpErr.RetryAfter) * time.Second
	if delay > remaining {
		c.errorf("messages dropped - %s", ErrRateLimitBudgetExceeded)
		c.notifyFailure(r.msgs, ErrRateLimitBudgetExceeded)
		return retryActionDrop
	}
	r.delay = delay
	return retryActionRateLimit
}

func (r *retryState) handleBackoff(lastErr error) retryAction {
	c := r.client

	// One reading serves the budget test and the clamp below, for the same reason
	// as handleRateLimit.
	now := c.now()
	if r.firstFailureTime.IsZero() {
		r.firstFailureTime = now
	}

	remaining := c.MaxTotalBackoffDuration - now.Sub(r.firstFailureTime)
	if remaining <= 0 {
		c.errorf("messages dropped - %s", ErrBackoffBudgetExceeded)
		c.notifyFailure(r.msgs, ErrBackoffBudgetExceeded)
		return retryActionDrop
	}

	r.backoffAttempts++
	if r.backoffAttempts > c.MaxRetries {
		c.errorf("%d messages dropped after %d attempts", len(r.msgs), r.totalAttempts)
		c.notifyFailure(r.msgs, lastErr)
		return retryActionDrop
	}

	// Same rule as the rate-limit path, for the simpler reason that a backoff
	// outlasting the budget is a wait whose attempt can never happen.
	delay := c.RetryAfter(r.backoffAttempts - 1)
	if delay > remaining {
		c.errorf("messages dropped - %s", ErrBackoffBudgetExceeded)
		c.notifyFailure(r.msgs, ErrBackoffBudgetExceeded)
		return retryActionDrop
	}
	if delay < 0 {
		// RetryAfter is caller-supplied; a negative reaches time.NewTimer, which
		// fires immediately and spins this loop at whatever rate the server answers.
		delay = 0
	}
	r.delay = delay

	return retryActionBackoff
}

// Upload serialized batch message. attempt is 1-based (1 = first attempt).
// upload sends one attempt. A non-zero deadline bounds the request itself, so
// Close cannot overrun ShutdownTimeout by a whole HTTP round trip while waiting
// on a final attempt.
func (c *client) upload(b []byte, attempt int, deadline time.Time) error {
	url := c.Endpoint + "/v1/batch"
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		c.errorf("creating request - %s", err)
		return err
	}

	req.Header.Add("User-Agent", "analytics-go (version: "+Version+")")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(b)))
	req.SetBasicAuth(c.key, "")

	// Omitted on the first attempt so the server can tell a retry from a first try.
	if attempt > 1 {
		req.Header.Add("X-Retry-Count", strconv.Itoa(attempt-1))
	}

	if !deadline.IsZero() {
		ctx, cancel := context.WithDeadline(req.Context(), deadline)
		defer cancel()
		req = req.WithContext(ctx)
	}

	res, err := c.http.Do(req)
	if err != nil {
		c.errorf("sending request - %s", err)
		return err
	}

	defer res.Body.Close()
	return c.report(res)
}

// Report on response body.
func (c *client) report(res *http.Response) error {
	if isSuccess(res.StatusCode) {
		c.debugf("response %s", res.Status)
		return nil
	}

	// Read before the body: a mid-read I/O error must not lose the server's
	// Retry-After and route this attempt onto the counted-backoff budget instead
	// of the rate-limit one.
	retryable := retryableStatus(res.StatusCode)
	var retryAfterSecs int64
	if retryable {
		retryAfterSecs = parseRetryAfter(res.Header.Get("Retry-After"), maxRetryAfterSeconds)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.errorf("response %d %s - %s", res.StatusCode, res.Status, err)
		return &httpError{
			StatusCode: res.StatusCode,
			Retryable:  retryable,
			RetryAfter: retryAfterSecs,
			Body:       err.Error(),
		}
	}

	c.logf("response %d %s – %s", res.StatusCode, res.Status, string(body))

	return &httpError{
		StatusCode: res.StatusCode,
		Retryable:  retryable,
		RetryAfter: retryAfterSecs,
		Body:       string(body),
	}
}

// Batch loop.
func (c *client) loop() {
	defer close(c.shutdown)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	tick := time.NewTicker(c.Interval)
	defer tick.Stop()

	ex := newExecutor(c.maxConcurrentRequests)
	defer ex.close()

	mq := messageQueue{
		maxBatchSize:  c.BatchSize,
		maxBatchBytes: c.maxBatchBytes(),
	}

	for {
		select {
		case msg := <-c.msgs:
			c.push(&mq, msg, wg, ex)

		case <-tick.C:
			c.flush(&mq, wg, ex)

		case <-c.quit:
			c.debugf("exit requested – draining messages")

			// Drain the msg channel, we have to close it first so no more
			// messages can be pushed and otherwise the loop would never end.
			close(c.msgs)
			for msg := range c.msgs {
				c.push(&mq, msg, wg, ex)
			}

			c.flush(&mq, wg, ex)
			c.debugf("exit")
			return
		}
	}
}

func (c *client) push(q *messageQueue, m Message, wg *sync.WaitGroup, ex *executor) {
	var msg message
	var err error

	if msg, err = makeMessage(m, maxMessageBytes); err != nil {
		c.errorf("%s - %v", err, m)
		c.notifyFailure([]message{{m, nil}}, err)
		return
	}

	c.debugf("buffer (%d/%d) %v", len(q.pending), c.BatchSize, m)

	if msgs := q.push(msg); msgs != nil {
		c.debugf("exceeded messages batch limit with batch of %d messages – flushing", len(msgs))
		c.sendAsync(msgs, wg, ex)
	}
}

func (c *client) flush(q *messageQueue, wg *sync.WaitGroup, ex *executor) {
	if msgs := q.flush(); msgs != nil {
		c.debugf("flushing %d messages", len(msgs))
		c.sendAsync(msgs, wg, ex)
	}
}

func (c *client) debugf(format string, args ...interface{}) {
	if c.Verbose {
		c.logf(format, args...)
	}
}

func (c *client) logf(format string, args ...interface{}) {
	c.Logger.Logf(format, args...)
}

func (c *client) errorf(format string, args ...interface{}) {
	c.Logger.Errorf(format, args...)
}

func (c *client) maxBatchBytes() int {
	b, _ := json.Marshal(batch{
		MessageId: c.uid(),
		SentAt:    c.now(),
		Context:   c.DefaultContext,
	})
	return maxBatchBytes - len(b)
}

func (c *client) notifySuccess(msgs []message) {
	if c.Callback != nil {
		for _, m := range msgs {
			c.Callback.Success(m.msg)
		}
	}
}

func (c *client) notifyFailure(msgs []message, err error) {
	if c.Callback != nil {
		for _, m := range msgs {
			c.Callback.Failure(m.msg, err)
		}
	}
}
