package analytics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper type used to implement the io.Reader interface on function values.
type readFunc func([]byte) (int, error)

func (f readFunc) Read(b []byte) (int, error) { return f(b) }

// Helper type used to implement the http.RoundTripper interface on function
// values.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func (f roundTripperFunc) CancelRequest(r *http.Request) {}

// Instances of this type are used to mock the client callbacks in unit tests.
type testCallback struct {
	success func(Message)
	failure func(Message, error)
}

func (c testCallback) Success(m Message) {
	if c.success != nil {
		c.success(m)
	}
}

func (c testCallback) Failure(m Message, e error) {
	if c.failure != nil {
		c.failure(m, e)
	}
}

// Instances of this type are used to mock the client logger in unit tests.
type testLogger struct {
	logf   func(string, ...interface{})
	errorf func(string, ...interface{})
}

func (l testLogger) Logf(format string, args ...interface{}) {
	if l.logf != nil {
		l.logf(format, args...)
	}
}

func (l testLogger) Errorf(format string, args ...interface{}) {
	if l.errorf != nil {
		l.errorf(format, args...)
	}
}

var _ Message = (*testErrorMessage)(nil)

// Instances of this type are used to force message validation errors in unit
// tests.
type testErrorMessage struct{}

func (m testErrorMessage) internal() {
}

func (m testErrorMessage) Validate() error { return testError }

var (
	// A control error returned by mock functions to emulate a failure.
	testError = errors.New("test error")

	// HTTP transport that always succeeds.
	testTransportOK = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Proto:      r.Proto,
			ProtoMajor: r.ProtoMajor,
			ProtoMinor: r.ProtoMinor,
			Body:       ioutil.NopCloser(strings.NewReader("")),
			Request:    r,
		}, nil
	})

	// HTTP transport that sleeps for a little while and eventually succeeds.
	testTransportDelayed = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		time.Sleep(10 * time.Millisecond)
		return testTransportOK.RoundTrip(r)
	})

	// HTTP transport that always returns a 400.
	testTransportBadRequest = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(http.StatusBadRequest),
			StatusCode: http.StatusBadRequest,
			Proto:      r.Proto,
			ProtoMajor: r.ProtoMajor,
			ProtoMinor: r.ProtoMinor,
			Body:       ioutil.NopCloser(strings.NewReader("")),
			Request:    r,
		}, nil
	})

	// HTTP transport that always returns a 400 with an erroring body reader.
	testTransportBodyError = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(http.StatusBadRequest),
			StatusCode: http.StatusBadRequest,
			Proto:      r.Proto,
			ProtoMajor: r.ProtoMajor,
			ProtoMinor: r.ProtoMinor,
			Body:       ioutil.NopCloser(readFunc(func(b []byte) (int, error) { return 0, testError })),
			Request:    r,
		}, nil
	})

	// HTTP transport that always return an error.
	testTransportError = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, testError
	})
)

func fixture(name string) string {
	f, err := os.Open(filepath.Join("fixtures", name))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func mockId() string { return "I'm unique" }

func mockTime() time.Time {
	// time.Unix(0, 0) fails on Circle
	return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
}

func mockServer() (chan []byte, *httptest.Server) {
	done := make(chan []byte, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, r.Body)

		var v interface{}
		err := json.Unmarshal(buf.Bytes(), &v)
		if err != nil {
			panic(err)
		}

		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(err)
		}

		done <- b
	}))

	return done, server
}

func ExampleTrack() {
	body, server := mockServer()
	defer server.Close()

	client, _ := NewWithConfig("h97jamjwbh", Config{
		Endpoint:  server.URL,
		BatchSize: 1,
		now:       mockTime,
		uid:       mockId,
	})
	defer client.Close()

	client.Enqueue(Track{
		Type:         2,
		InstanceId:   "123456",
		DeploymentId: "qwerty",
	})

	fmt.Printf("%s\n", <-body)
	// Output:
	// {
	//   "batch": [
	//     {
	//       "alloc": 0,
	//       "did": "qwerty",
	//       "frees": 0,
	//       "heapAlloc": 0,
	//       "heapIdle": 0,
	//       "heapInuse": 0,
	//       "heapObjects": 0,
	//       "heapReleased": 0,
	//       "heapSys": 0,
	//       "iid": "123456",
	//       "lookups": 0,
	//       "mallocs": 0,
	//       "mid": "I'm unique",
	//       "numGC": 0,
	//       "p": "",
	//       "sys": 0,
	//       "t": 2,
	//       "totalAlloc": 0,
	//       "ts": "2009-11-10T23:00:00Z",
	//       "v": 1
	//     }
	//   ],
	//   "messageId": "I'm unique",
	//   "sentAt": "2009-11-10T23:00:00Z"
	// }
}

func TestEnqueue(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		fixture string
	}{
		{"identify", Identify{InstanceId: "A", DeploymentId: "B"}, "test-enqueue-identify.json"},
		{"*identify", &Identify{InstanceId: "A", DeploymentId: "B"}, "test-enqueue-identify.json"},
		{"track", Track{InstanceId: "A", DeploymentId: "B"}, "test-enqueue-track.json"},
		{"*track", &Track{InstanceId: "A", DeploymentId: "B"}, "test-enqueue-track.json"},
		{"page", Page{InstanceId: "A", DeploymentId: "B"}, "test-enqueue-page.json"},
		{"*page", &Page{InstanceId: "A", DeploymentId: "B"}, "test-enqueue-page.json"},
		{"alias", Alias{PreviousId: "A", UserId: "B"}, ""},
		{"*alias", &Alias{PreviousId: "A", UserId: "B"}, ""},
		{"group", Group{GroupId: "A", UserId: "B"}, ""},
		{"*group", &Group{GroupId: "A", UserId: "B"}, ""},
		{"screen", Screen{Name: "A", UserId: "B"}, ""},
		{"*screen", &Screen{Name: "A", UserId: "B"}, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, server := mockServer()
			defer server.Close()
			client, err := NewWithConfig("local-fixture-key", Config{
				Endpoint: server.URL, Transport: server.Client().Transport, BatchSize: 1, now: mockTime, uid: mockId,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			err = client.Enqueue(test.msg)
			if test.fixture == "" {
				want := fmt.Sprintf("messages with custom types cannot be enqueued: %T", test.msg)
				if err == nil || err.Error() != want {
					t.Fatalf("unsupported message: got %v, want %s", err, want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			select {
			case res := <-body:
				if want := fixture(test.fixture); string(res) != want {
					t.Errorf("compact batch: got %s, want %s", res, want)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("batch was not delivered")
			}
		})
	}
}

var _ Message = (*customMessage)(nil)

type customMessage struct {
}

func (c *customMessage) internal() {
}

func (c *customMessage) Validate() error {
	return nil
}

func TestEnqueuingCustomTypeFails(t *testing.T) {
	client := New("0123456789")
	err := client.Enqueue(&customMessage{})

	if err.Error() != "messages with custom types cannot be enqueued: *analytics.customMessage" {
		t.Errorf("invalid/missing error when queuing unsupported message: %v", err)
	}
}

// trackBatchFixture varies only the fields relevant to each batching test,
// keeping the complete wire-format expectation in one independent JSON fixture.
func trackBatchFixture(t *testing.T, count int, messageID string) string {
	t.Helper()
	var want map[string]interface{}
	if err := json.Unmarshal([]byte(fixture("test-enqueue-track.json")), &want); err != nil {
		t.Fatal(err)
	}
	event := want["batch"].([]interface{})[0].(map[string]interface{})
	if messageID != "" {
		event["mid"] = messageID
	}
	events := make([]interface{}, count)
	for i := range events {
		events[i] = event
	}
	want["batch"] = events
	encoded, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func TestTrackBatching(t *testing.T) {
	for _, test := range []struct {
		name      string
		interval  time.Duration
		batchSize int
		count     int
		messageID string
		timestamp time.Time
	}{
		{name: "interval", interval: 100 * time.Millisecond, batchSize: 100, count: 1},
		{name: "timestamp uses enqueue time", batchSize: 1, count: 1, timestamp: time.Date(2015, time.July, 10, 23, 0, 0, 0, time.UTC)},
		{name: "explicit message ID", batchSize: 1, count: 1, messageID: "abc"},
		{name: "batch size", batchSize: 3, count: 5},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, server := mockServer()
			defer server.Close()
			started := time.Now()
			client, err := NewWithConfig("local-fixture-key", Config{
				Endpoint: server.URL, Transport: server.Client().Transport,
				Interval: test.interval, BatchSize: test.batchSize, now: mockTime, uid: mockId,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			for i := 0; i < test.count; i++ {
				if err := client.Enqueue(Track{
					InstanceId: "A", DeploymentId: "B", MessageId: test.messageID, Timestamp: test.timestamp,
				}); err != nil {
					t.Fatal(err)
				}
			}
			count := test.count
			if count > test.batchSize {
				count = test.batchSize
			}
			select {
			case got := <-body:
				if want := trackBatchFixture(t, count, test.messageID); string(got) != want {
					t.Errorf("compact batch: got %s, want %s", got, want)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("batch was not delivered")
			}
			if time.Since(started) < test.interval {
				t.Error("batch flushed before the configured interval")
			}
		})
	}
}

func TestClientCloseTwice(t *testing.T) {
	client := New("0123456789")

	if err := client.Close(); err != nil {
		t.Error("closing a client should not a return an error")
	}

	if err := client.Close(); err != ErrClosed {
		t.Error("closing a client a second time should return ErrClosed:", err)
	}

	if err := client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"}); err != ErrClosed {
		t.Error("using a client after it was closed should return ErrClosed:", err)
	}
}

func TestClientConfigError(t *testing.T) {
	client, err := NewWithConfig("0123456789", Config{
		Interval: -1 * time.Second,
	})

	if err == nil {
		t.Error("no error returned when creating a client with an invalid config")
	}

	if _, ok := err.(ConfigError); !ok {
		t.Errorf("invalid error type returned when creating a client with an invalid config: %T", err)
	}

	if client != nil {
		t.Error("invalid non-nil client object returned when creating a client with and invalid config:", client)
		client.Close()
	}
}

func TestClientEnqueueError(t *testing.T) {
	client := New("0123456789")
	defer client.Close()

	if err := client.Enqueue(testErrorMessage{}); err != testError {
		t.Error("invlaid error returned when queueing an invalid message:", err)
	}
}

func TestClientCallback(t *testing.T) {
	reschan := make(chan bool, 1)
	errchan := make(chan error, 1)

	client, _ := NewWithConfig("0123456789", Config{
		Logger: testLogger{t.Logf, t.Logf},
		Callback: testCallback{
			func(m Message) { reschan <- true },
			func(m Message, e error) { errchan <- e },
		},
		Transport: testTransportOK,
	})

	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})
	client.Close()

	select {
	case <-reschan:
	case err := <-errchan:
		t.Error("failure callback triggered:", err)
	}
}

func TestClientNewRequestError(t *testing.T) {
	errchan := make(chan error, 1)

	client, _ := NewWithConfig("0123456789", Config{
		Endpoint: "://localhost:80", // Malformed endpoint URL.
		Logger:   testLogger{t.Logf, t.Logf},
		Callback: testCallback{
			nil,
			func(m Message, e error) { errchan <- e },
		},
		Transport: testTransportOK,
	})

	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})
	client.Close()

	if err := <-errchan; err == nil {
		t.Error("failure callback not triggered for an invalid request")
	}
}

func TestClientRoundTripperError(t *testing.T) {
	errchan := make(chan error, 1)

	client, _ := NewWithConfig("0123456789", Config{
		Logger: testLogger{t.Logf, t.Logf},
		Callback: testCallback{
			nil,
			func(m Message, e error) { errchan <- e },
		},
		Transport: testTransportError,
	})

	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})
	client.Close()

	if err := <-errchan; err == nil {
		t.Error("failure callback not triggered for an invalid request")

	} else if e, ok := err.(*url.Error); !ok {
		t.Errorf("invalid error returned by round tripper: %T: %s", err, err)

	} else if e.Err != testError {
		t.Errorf("invalid error returned by round tripper: %T: %s", e.Err, e.Err)
	}
}

func TestClientRetryError(t *testing.T) {
	errchan := make(chan error, 1)

	client, _ := NewWithConfig("0123456789", Config{
		Logger: testLogger{t.Logf, t.Logf},
		Callback: testCallback{
			nil,
			func(m Message, e error) { errchan <- e },
		},
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return nil, testError
		}),
		BatchSize:  1,
		RetryAfter: func(i int) time.Duration { return time.Millisecond },
	})

	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})

	// Each retry should happen ~1 millisecond, this should give enough time to
	// the test to trigger the failure callback.
	time.Sleep(50 * time.Millisecond)

	if err := <-errchan; err == nil {
		t.Error("failure callback not triggered for a retry falure")

	} else if e, ok := err.(*url.Error); !ok {
		t.Errorf("invalid error returned by round tripper: %T: %s", err, err)

	} else if e.Err != testError {
		t.Errorf("invalid error returned by round tripper: %T: %s", e.Err, e.Err)
	}

	client.Close()
}

func TestClientResponse400(t *testing.T) {
	errchan := make(chan error, 1)

	client, _ := NewWithConfig("0123456789", Config{
		Logger: testLogger{t.Logf, t.Logf},
		Callback: testCallback{
			nil,
			func(m Message, e error) { errchan <- e },
		},
		// This HTTP transport always return 400's.
		Transport: testTransportBadRequest,
	})

	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})
	client.Close()

	if err := <-errchan; err == nil {
		t.Error("failure callback not triggered for a 400 response")
	}
}

func TestClientResponseBodyError(t *testing.T) {
	errchan := make(chan error, 1)

	client, _ := NewWithConfig("0123456789", Config{
		Logger: testLogger{t.Logf, t.Logf},
		Callback: testCallback{
			nil,
			func(m Message, e error) { errchan <- e },
		},
		// This HTTP transport always return 400's with an erroring body.
		Transport: testTransportBodyError,
	})

	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})
	client.Close()

	if err := <-errchan; err == nil {
		t.Error("failure callback not triggered for a 400 response")

	} else if err != testError {
		t.Errorf("invalid error returned by erroring response body: %T: %s", err, err)
	}
}

func TestClientMaxConcurrentRequests(t *testing.T) {
	reschan := make(chan bool, 1)
	errchan := make(chan error, 1)

	client, _ := NewWithConfig("0123456789", Config{
		Logger: testLogger{t.Logf, t.Logf},
		Callback: testCallback{
			func(m Message) { reschan <- true },
			func(m Message, e error) { errchan <- e },
		},
		Transport: testTransportDelayed,
		// Only one concurreny request can be submitted, because the transport
		// introduces a short delay one of the uploads should fail.
		BatchSize:             1,
		maxConcurrentRequests: 1,
	})

	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})
	client.Enqueue(Track{Type: 2, InstanceId: "A", DeploymentId: "B"})
	client.Close()

	if _, ok := <-reschan; !ok {
		t.Error("one of the requests should have succeeded but the result channel was empty")
	}

	if err := <-errchan; err == nil {
		t.Error("failure callback not triggered after reaching the request limit")

	} else if err != ErrTooManyRequests {
		t.Errorf("invalid error returned by erroring response body: %T: %s", err, err)
	}
}
