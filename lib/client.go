package lib

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

//Client Interface
type Client interface {
	Notify(req *http.Request) <-chan *NotificationResult
	Prepare(url, body string) (request *http.Request, error error)
	GetNotificationResult(<-chan *NotificationResult) *NotificationResult
}

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type customClient struct {
	client    doer
	isVerbose bool
}

type NotificationResult struct {
	IsError      bool
	ErrorDetails error
	Response     *http.Response
}

// New is the constructor for the Client
func New(setters ...Option) Client {
	o := &Options{
		isVerbose: false,
	}
	for _, setter := range setters {
		setter(o)
	}
	return &customClient{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// TLS below 1.1 is deprecated
					MinVersion: tls.VersionTLS12,
				},
			},
		},
		isVerbose: o.isVerbose,
	}
}

// The Notify function will process the http request asynchronously
// and return a channel for the notification result to be sent into.
func (c *customClient) Notify(request *http.Request) <-chan *NotificationResult {
	res := make(chan *NotificationResult, 1)
	go func() {
		defer close(res)

		if c.isVerbose {
			dumpRequest, errLog := httputil.DumpRequestOut(request, true)
			if errLog == nil {
				requestDump := string(dumpRequest)
				log.Printf("request:\n %s\n", requestDump)
			}
		}
		buf := new(bytes.Buffer)
		buf.Reset()
		// send payload
		response, err := c.client.Do(request)
		// handle technical errors
		if err != nil {
			res <- &NotificationResult{
				IsError:      true,
				ErrorDetails: err,
			}
			return
		}
		// flush body at the end to avoid memory leak
		defer func() {
			_ = response.Body.Close()
		}()
		// handle business errors
		if inError(response.StatusCode) {
			body, errRead := ioutil.ReadAll(response.Body)
			if errRead != nil {
				res <- &NotificationResult{
					IsError:      true,
					ErrorDetails: errors.Wrap(errRead, "cannot decode response body"),
				}
				return
			}
			res <- &NotificationResult{
				IsError:      true,
				ErrorDetails: fmt.Errorf("error occurred: %s", body),
			}
			return
		}
		// success !!
		res <- &NotificationResult{
			IsError: false,
		}
		if c.isVerbose {
			dumpResponse, errLog := httputil.DumpResponse(response, true)
			if errLog == nil {
				responseDump := string(dumpResponse)
				log.Printf("response:\n %s\n", responseDump)
			}
		}

	}()
	return res
}

// The Prepare function will prepare the http.Request based on the url path and the body provided.
// The default method used is POST.
// An error will be thrown if the url does not have a correct format.
func (c *customClient) Prepare(path, body string) (request *http.Request, error error) {
	if _, err := url.ParseRequestURI(path); err != nil {
		return nil, err
	}
	return http.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
}

// The GetNotificationResult will retrieve the NotificationResult from the channel filled by Notify
func (c *customClient) GetNotificationResult(res <-chan *NotificationResult) *NotificationResult {
	return <-res
}

func inError(status int) bool {
	return !(status >= http.StatusOK && status < http.StatusMultipleChoices)
}
