package gqfsl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue_Add(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	msg := Message{ID: "1"}
	storeMock.On("Add", msg).Return(nil)

	err := q.Add(msg)
	assert.NoError(t, err)
	storeMock.AssertExpectations(t)
}

func TestQueue_Add_Error(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	msg := Message{ID: "1"}
	expectedErr := errors.New("storage error")
	storeMock.On("Add", msg).Return(expectedErr)

	err := q.Add(msg)
	assert.EqualError(t, err, expectedErr.Error())
	storeMock.AssertExpectations(t)
}

func TestQueue_Send_Success(t *testing.T) {
	emailMock := &MockEmailSender{}
	q := &Queue{
		emailSender: emailMock,
	}

	msg := Message{ID: "1"}
	emailMock.On("Send", msg).Return(nil)

	err := q.Send(msg)
	assert.NoError(t, err)
	emailMock.AssertExpectations(t)
}

func TestQueue_Send_Failure(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}
	q := &Queue{
		emailSender: emailMock,
		store:       storeMock,
	}

	msg := Message{ID: "1"}
	sendErr := errors.New("smtp error")
	emailMock.On("Send", msg).Return(sendErr)
	storeMock.On("Add", msg).Return(nil)

	err := q.Send(msg)
	assert.NoError(t, err) // Ошибка отправки, но успешное сохранение
	emailMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestQueue_Send_FailureWithStorageError(t *testing.T) {
	emailMock := &MockEmailSender{}
	storeMock := &MockMessageStore{}
	q := &Queue{
		emailSender: emailMock,
		store:       storeMock,
	}

	msg := Message{ID: "1"}
	sendErr := errors.New("smtp error")
	storageErr := errors.New("storage error")
	emailMock.On("Send", msg).Return(sendErr)
	storeMock.On("Add", msg).Return(storageErr)

	err := q.Send(msg)
	assert.EqualError(t, err, storageErr.Error())
	emailMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestQueue_Get(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	id := int64(1)
	expectedMsg := &Message{ID: "1"}
	storeMock.On("Get", id).Return(expectedMsg, nil)

	msg, err := q.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg, msg)
	storeMock.AssertExpectations(t)
}

func TestQueue_Get_Error(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	id := int64(1)
	expectedErr := errors.New("not found")
	storeMock.On("Get", id).Return((*Message)(nil), expectedErr)

	msg, err := q.Get(id)
	assert.Nil(t, msg)
	assert.EqualError(t, err, "failed to get message: "+expectedErr.Error())
	storeMock.AssertExpectations(t)
}

func TestQueue_List(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	expectedMsgs := []*Message{
		{ID: "1"},
		{ID: "2"},
	}
	storeMock.On("List").Return(expectedMsgs, nil)

	msgs, err := q.List()
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgs, msgs)
	storeMock.AssertExpectations(t)
}

func TestQueue_List_Error(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	expectedErr := errors.New("storage error")
	storeMock.On("List").Return([]*Message(nil), expectedErr)

	msgs, err := q.List()
	assert.Nil(t, msgs)
	assert.EqualError(t, err, "failed to list messages: "+expectedErr.Error())
	storeMock.AssertExpectations(t)
}

func TestQueue_Delete(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	id := int64(1)
	storeMock.On("Delete", id).Return(nil)

	err := q.Delete(id)
	assert.NoError(t, err)
	storeMock.AssertExpectations(t)
}

func TestQueue_Delete_Error(t *testing.T) {
	storeMock := &MockMessageStore{}
	q := &Queue{
		store: storeMock,
	}

	id := int64(1)
	expectedErr := errors.New("delete error")
	storeMock.On("Delete", id).Return(expectedErr)

	err := q.Delete(id)
	assert.EqualError(t, err, expectedErr.Error())
	storeMock.AssertExpectations(t)
}

func TestQueue_Stop(t *testing.T) {
	cronMock := &MockCronService{}
	q := &Queue{
		cron: cronMock,
	}

	cronMock.On("Stop")

	q.Stop()
	cronMock.AssertExpectations(t)
}
