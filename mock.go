package gqfsl

import "github.com/stretchr/testify/mock"

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

func (m *MockMessageStore) Add(message Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageStore) Get(id int64) (*Message, error) {
	args := m.Called(id)
	return args.Get(0).(*Message), args.Error(1)
}

func (m *MockMessageStore) List() ([]*Message, error) {
	args := m.Called()
	return args.Get(0).([]*Message), args.Error(1)
}

func (m *MockMessageStore) Update(message Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageStore) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMessageStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockCronService реализует интерфейс cronService
type MockCronService struct {
	mock.Mock
}

func (m *MockCronService) Stop() {
	m.Called()
}
