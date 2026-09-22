package analytics

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestEditionBatchProtocol(t *testing.T) {
	for _, edition := range []string{"", "oss", "oel", "network", "future-edition"} {
		for _, compression := range []int{0, 6} {
			t.Run(fmt.Sprintf("edition=%s/gzip=%d", edition, compression), func(t *testing.T) {
				bodies := make(chan []byte, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					user, password, ok := r.BasicAuth()
					if !ok || user != "local-fixture-key" || password != "" || r.Method != "POST" || r.URL.Path != "/v1/batch" {
						t.Error("unexpected authentication or batch endpoint")
					}
					var reader io.Reader = r.Body
					if compression != 0 {
						if r.Header.Get("Content-Encoding") != "gzip" {
							t.Error("missing gzip header")
						}
						gz, err := gzip.NewReader(r.Body)
						if err != nil {
							t.Error(err)
							w.WriteHeader(400)
							return
						}
						defer gz.Close()
						reader = gz
					} else if r.Header.Get("Content-Encoding") != "" {
						t.Error("unexpected compression")
					}
					body, err := ioutil.ReadAll(reader)
					if err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					bodies <- body
					w.WriteHeader(http.StatusOK)
				}))
				defer server.Close()
				stamp := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
				client, err := NewWithConfig("local-fixture-key", Config{
					Endpoint: server.URL, Transport: server.Client().Transport,
					GzipCompressionLevel: compression, Interval: time.Hour, BatchSize: 100,
					now: func() time.Time { return stamp }, uid: func() string { return "fixture-batch" },
				})
				if err != nil {
					t.Fatal(err)
				}
				for _, msg := range []Message{
					Identify{Edition: edition, Project: "talos", MessageId: "identify-id", InstanceId: "instance-id", DeploymentId: "deployment-id", DatabaseDialect: "sqlite", ProductVersion: "v1.2.3", Startup: true},
					&Track{Edition: edition, Project: "talos", MessageId: "track-id", InstanceId: "instance-id", DeploymentId: "deployment-id", OsName: "linux", OsArchitecture: "amd64", CPU: 2},
					Page{Edition: edition, Project: "talos", MessageId: "page-id", InstanceId: "instance-id", DeploymentId: "deployment-id", UrlHost: "example.com", UrlPath: "/api", RequestCode: 201, RequestLatency: 12},
				} {
					if err := client.Enqueue(msg); err != nil {
						t.Fatal(err)
					}
				}
				if err := client.Close(); err != nil {
					t.Fatal(err)
				}
				var actual map[string]interface{}
				select {
				case body := <-bodies:
					if err := json.Unmarshal(body, &actual); err != nil {
						t.Fatal(err)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("batch was not delivered")
				}
				for _, event := range actual["batch"].([]interface{}) {
					fields := event.(map[string]interface{})
					got, exists := fields["ed"]
					if edition == "" && exists {
						t.Error("unset edition must be omitted")
					}
					if edition != "" && got != edition {
						t.Errorf("edition: got %v, want %s", got, edition)
					}
					delete(fields, "ed")
				}
				var expected map[string]interface{}
				if err := json.Unmarshal([]byte(fixture("compact-batch.json")), &expected); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("compact protocol changed:\ngot %v\nwant %v", actual, expected)
				}
			})
		}
	}
}

func TestEditionBatchRetryPreservesPayload(t *testing.T) {
	requests := make(chan []byte, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		requests <- body
		if len(requests) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	delivered := make(chan struct{}, 1)
	client, err := NewWithConfig("local-fixture-key", Config{
		Endpoint: server.URL, Transport: server.Client().Transport, GzipCompressionLevel: 6,
		BatchSize: 1, RetryAfter: func(int) time.Duration { return time.Millisecond },
		Callback: testCallback{success: func(Message) { delivered <- struct{}{} }},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.Enqueue(Identify{Edition: "network", Project: "talos", InstanceId: "i", DeploymentId: "d", Startup: true, IsOptOut: true}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-delivered:
	case <-time.After(5 * time.Second):
		t.Fatal("retry did not deliver")
	}
	if first, second := <-requests, <-requests; !bytes.Equal(first, second) {
		t.Fatal("retry changed the compressed batch, timestamps or message IDs")
	}
}
