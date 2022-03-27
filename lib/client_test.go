package lib

import (
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
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

func TestCompleteNotify(t *testing.T) {
	client := New()
	prepare, err := client.Prepare("http://localhost:8080/xyz", `{"name":"John"}`)
	if err == nil {
		t.Errorf("prepare error = %+v", err)
		return
	}
	results := client.Notify(prepare)
	result := client.GetNotificationResult(results)
	log.Print(result)
}
