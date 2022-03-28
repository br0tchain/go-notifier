package mocks

import (
	"github.com/br0tchain/go-notifier/lib"
	"github.com/stretchr/testify/mock"
	"net/http"
)

type LibClient struct {
	mock.Mock
}

func (l *LibClient) Notify(req *http.Request) <-chan *lib.NotificationResult {
	args := l.Called(req)
	resp, ok := args.Get(0).(<-chan *lib.NotificationResult)
	if !ok {
		resp = nil
	}
	return resp
}

func (l *LibClient) Prepare(url, body string) (request *http.Request, error error) {
	args := l.Called(url, body)
	resp, ok := args.Get(0).(*http.Request)
	if !ok {
		resp = nil
	}
	return resp, args.Error(1)
}

func (l *LibClient) GetNotificationResult(results <-chan *lib.NotificationResult) *lib.NotificationResult {
	args := l.Called(results)
	resp, ok := args.Get(0).(*lib.NotificationResult)
	if !ok {
		resp = nil
	}
	return resp
}
