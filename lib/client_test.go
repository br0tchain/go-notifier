package lib

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockDoer struct {
	Response *http.Response
	Error    error
}

func (m mockDoer) Do(req *http.Request) (*http.Response, error) {
	return m.Response, m.Error
}

func TestNew(t *testing.T) {
	client := New()
	require.NotNil(t, client)
}

func TestPrepare(t *testing.T) {
	client := New()

	path := "http://localhost:8080/notify"
	body := "content to be sent"
	request := newRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))

	prepare, err := client.Prepare(path, body)
	require.Nil(t, err)
	require.NotNil(t, prepare)
	require.Equal(t, prepare.URL, request.URL)
	require.Equal(t, prepare.Body, request.Body)
	require.Equal(t, prepare.Method, request.Method)
}

func TestPrepare_error(t *testing.T) {
	client := New()

	path := "xyz"
	body := "content to be sent"

	prepare, err := client.Prepare(path, body)
	require.NotNil(t, err)
	require.Nil(t, prepare)
}

func Test_customClient_sendNotification(t *testing.T) {
	res := make(chan *NotificationResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/200", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
	})
	server := httptest.NewServer(mux)

	c := &customClient{
		client:    server.Client(),
		isVerbose: true,
	}
	c.sendNotification(newRequest(http.MethodPost, server.URL+"/200", strings.NewReader("toto")), res)
	notif := c.GetNotificationResult(res)
	require.NotNil(t, notif)
	require.False(t, notif.IsError)
	require.Nil(t, notif.ErrorDetails)
}

func Test_customClient_sendNotification_error(t *testing.T) {
	res := make(chan *NotificationResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/500", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(500)
	})
	server := httptest.NewServer(mux)

	c := &customClient{
		client:    server.Client(),
		isVerbose: true,
	}
	c.sendNotification(newRequest(http.MethodPost, server.URL+"/500", strings.NewReader("toto")), res)
	notif := c.GetNotificationResult(res)
	require.NotNil(t, notif)
	require.True(t, notif.IsError)
	require.NotNil(t, notif.ErrorDetails)
}

func newRequest(method, url string, body io.Reader) *http.Request {
	request, _ := http.NewRequest(method, url, body)
	return request
}
