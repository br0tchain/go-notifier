package usecase

import (
	"bytes"
	"fmt"
	"github.com/br0tchain/go-notifier/internal/mocks"
	"github.com/br0tchain/go-notifier/lib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_notifier_parseInput(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	res := make(chan *lib.NotificationResult, 1)
	path := "http://localhost:8080/notify"
	body := "content to be sent"
	request, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	l := &lib.NotificationResult{IsError: false}

	mockClient.On("Prepare", path, body).Return(request, nil)
	mockClient.On("Notify", mock.Anything).Return(res).Once()
	mockClient.On("GetNotificationResult", mock.Anything).Return(l)
	err := n.parseInput(fmt.Sprintf(`notify --url %s -m "%s"`, path, body))
	time.Sleep(100 * time.Millisecond)
	require.Nil(t, err)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

//testing than in 5 seconds, we call 6 times (T0, T1s, T2s, T3s, T4s, T5s)
func Test_notifier_parseInput_checkInterval(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	res := make(chan *lib.NotificationResult, 1)
	path := "http://localhost:8080/notify"
	body := "content to be sent"
	request, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	l := &lib.NotificationResult{IsError: false}

	mockClient.On("Prepare", path, body).Return(request, nil)
	mockClient.On("Notify", mock.Anything).Return(res).Times(6)
	mockClient.On("GetNotificationResult", mock.Anything).Return(l)
	err := n.parseInput(fmt.Sprintf(`notify --url %s -m "%s" -i 1s`, path, body))
	time.Sleep(5*time.Second + 50*time.Millisecond)
	require.Nil(t, err)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_notifier_parseInput_need_help(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	err := n.parseInput("notify --help")
	require.Nil(t, err)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_notifier_parseInput_error_input(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	err := n.parseInput("random input")
	require.NotNil(t, err)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_notifier_parseInput_error_missing_params(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	err := n.parseInput("notify --url=http://www.az.yty  ")
	require.NotNil(t, err)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_notifier_parseInput_error_prepare(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	path := "error"
	body := "content to be sent"
	mockClient.On("Prepare", path, body).Return(nil, fmt.Errorf("error on url"))
	err := n.parseInput(fmt.Sprintf(`notify --url %s -m "%s"`, path, body))
	require.NotNil(t, err)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_parseFlags_nominal_message(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	flags, err := n.parseFlags(`-m "content to be sent"`)
	require.Nil(t, err)
	require.NotNil(t, flags)
	require.Equal(t, flags.Interval.String(), defaultDuration)
	require.Equal(t, flags.Body, "content to be sent")
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_parseFlags_nominal_message_interval1(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	flags, err := n.parseFlags(`-m "content to be sent" -i 10s`)
	require.Nil(t, err)
	require.NotNil(t, flags)
	require.Equal(t, flags.Interval.String(), "10s")
	require.Equal(t, flags.Body, "content to be sent")
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_parseFlags_nominal_message_interval2(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	flags, err := n.parseFlags(`-i 3ms -m "content to be sent"`)
	require.Nil(t, err)
	require.NotNil(t, flags)
	require.Equal(t, flags.Interval.String(), "3ms")
	require.Equal(t, flags.Body, "content to be sent")
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_parseFlags_error_load_file(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	fileName := "resources/sample_message.txt"
	flags, err := n.parseFlags(fmt.Sprintf(`-f %s`, fileName))
	require.NotNil(t, err)
	require.Nil(t, flags)
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf(errTemplateNoFileFound, fileName)))
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_parseFlags_error_no_content(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	flags, err := n.parseFlags(`-i 2m`)
	require.NotNil(t, err)
	require.Nil(t, flags)
	require.Equal(t, err.Error(), errNoContent)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}

func Test_parseFlags_error_wrong_interval(t *testing.T) {
	mockClient := new(mocks.LibClient)
	n := notifier{
		Client: mockClient,
	}
	flags, err := n.parseFlags(`-i azert`)
	require.NotNil(t, err)
	require.Nil(t, flags)
	require.True(t, mock.AssertExpectationsForObjects(t, mockClient))
}
