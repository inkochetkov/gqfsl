package gqfsl

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEmailSender implements the interface emailSender
type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) Send(msg Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

// MockMessageStore implements the interface messageStore
type MockMessageStore struct {
	mock.Mock
}

func (m *MockMessageStore) List() ([]*Message, error) {
	args := m.Called()
	return args.Get(0).([]*Message), args.Error(1)
}

func (m *MockMessageStore) Update(msg Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockMessageStore) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCron_ProcessMessages_Sequential(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}

	config := Config{
		Cron: CronConf{
			DurationRetry:       1 * time.Minute,
			KeepingTheOrder:     true,
			DurationSaveSuccess: 1 * time.Hour,
		},
	}

	msgs := []*Message{
		{ID: "1", Status: make(map[string]any)},
		{ID: "2", Status: make(map[string]any)},
	}

	storeMock.On("List").Return(msgs, nil)
	emailMock.On("Send", *msgs[0]).Return(nil)
	emailMock.On("Send", *msgs[1]).Return(nil)
	storeMock.On("Update", mock.Anything).Return(nil).Twice()

	c := &cron{
		config:      config,
		emailSender: emailMock,
		store:       storeMock,
		stopChan:    make(chan struct{}),
	}

	c.processMessages()

	storeMock.AssertExpectations(t)
	emailMock.AssertExpectations(t)
	assert.Equal(t, 2, len(emailMock.Calls))
}

func TestCron_ProcessMessages_Parallel(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}

	config := Config{
		Cron: CronConf{
			DurationRetry:       1 * time.Minute,
			KeepingTheOrder:     false,
			DurationSaveSuccess: 1 * time.Hour,
		},
	}

	msgs := []*Message{
		{ID: "1", Status: make(map[string]any)},
		{ID: "2", Status: make(map[string]any)},
	}

	storeMock.On("List").Return(msgs, nil)
	emailMock.On("Send", mock.Anything).Return(nil).Twice()
	storeMock.On("Update", mock.Anything).Return(nil).Twice()

	c := &cron{
		config:      config,
		emailSender: emailMock,
		store:       storeMock,
		stopChan:    make(chan struct{}),
	}

	c.processMessages()

	storeMock.AssertExpectations(t)
	emailMock.AssertExpectations(t)
	assert.Equal(t, 2, len(emailMock.Calls))
}

func TestCron_ProcessMessage_Success(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}

	config := Config{
		Cron: CronConf{
			DurationSaveSuccess: 1 * time.Hour,
		},
	}

	msg := &Message{
		ID:     "1",
		Status: make(map[string]any),
	}

	emailMock.On("Send", *msg).Return(nil)
	storeMock.On("Update", mock.Anything).Return(nil)

	c := &cron{
		config:      config,
		emailSender: emailMock,
		store:       storeMock,
	}

	c.processMessage(msg)

	assert.Equal(t, "sent", msg.Status["status"])
	assert.NotNil(t, msg.Status["time_send"])
	assert.Equal(t, 1.0, msg.Status["count_try_send"])

	emailMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestCron_ProcessMessage_Failure(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}

	msg := &Message{
		ID:     "1",
		Status: make(map[string]any),
	}

	expectedErr := errors.New("smtp error")
	emailMock.On("Send", *msg).Return(expectedErr)
	storeMock.On("Update", mock.Anything).Return(nil)

	c := &cron{
		config:      Config{},
		emailSender: emailMock,
		store:       storeMock,
	}

	c.processMessage(msg)

	assert.Equal(t, expectedErr.Error(), msg.Status["error"])
	assert.NotNil(t, msg.Status["last_try"])
	assert.Equal(t, 1.0, msg.Status["count_try_send"])

	emailMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestCron_ProcessMessage_DeleteAfterSuccess(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}

	config := Config{
		Cron: CronConf{
			DurationSaveSuccess: 1 * time.Hour,
		},
	}

	// Set the send time in the past (greater than DurationSaveSuccess)
	oldTime := time.Now().Add(-2 * time.Hour).Unix()
	msg := &Message{
		ID: "1",
		Status: map[string]any{
			"status":    "sent",
			"time_send": oldTime,
		},
	}

	// Setting up mocks
	storeMock.On("Delete", int64(1)).Return(nil)

	c := &cron{
		config:      config,
		emailSender: emailMock,
		store:       storeMock,
	}

	// Processing the message
	c.processMessage(msg)

	// Check that Send was not called
	emailMock.AssertNotCalled(t, "Send")

	// Check that Delete was called
	storeMock.AssertCalled(t, "Delete", int64(1))

	// Check that Update was not called (message removed)
	storeMock.AssertNotCalled(t, "Update")
}

func TestCron_Stop(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}

	config := Config{
		Cron: CronConf{
			DurationRetry: 100 * time.Millisecond,
		},
	}

	// Setting up a wait for List()
	storeMock.On("List").Return([]*Message{}, nil)

	c := &cron{
		config:      config,
		emailSender: emailMock,
		store:       storeMock,
		stopChan:    make(chan struct{}),
		wg:          sync.WaitGroup{},
	}

	c.wg.Add(1)
	go c.run()

	time.Sleep(200 * time.Millisecond)

	c.Stop()

	c.wg.Wait()

	// Check that List was called
	storeMock.AssertCalled(t, "List")
}
func TestCron_ProcessMessage_SkipSent(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}

	config := Config{
		Cron: CronConf{
			DurationSaveSuccess: 1 * time.Hour,
		},
	}

	msg := &Message{
		ID: "1",
		Status: map[string]any{
			"status":    "sent",
			"time_send": time.Now().Unix(),
		},
	}

	// Wait only for Delete call (if time expired) or nothing
	storeMock.On("Delete", int64(1)).Return(nil)

	c := &cron{
		config:      config,
		emailSender: emailMock,
		store:       storeMock,
	}

	c.processMessage(msg)

	// Check that Send was not called
	emailMock.AssertNotCalled(t, "Send")

	// We check that the attempt counter has not changed.
	assert.Equal(t, nil, msg.Status["count_try_send"])
}
