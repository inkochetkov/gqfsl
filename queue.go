package gqfsl

import (
	"fmt"
)

type Queue struct {
	config Config

	emailSender emailSender
	store       messageStore
	cron        cronService
}

func New(cgf Config) (*Queue, error) {

	config := checkDefaultConfig(cgf)

	emailServer, err := connEmailServer(config)
	if err != nil {
		return nil, err
	}

	sql, err := startSQL(config)
	if err != nil {
		return nil, err
	}

	cron, err := startCron(config, emailServer, sql)
	if err != nil {
		return nil, err
	}
	return &Queue{config, emailServer, sql, cron}, nil
}

// Add delayed dispatch
func (q *Queue) Add(message Message) error {
	return q.store.Add(message)
}

// Send immediate dispatch
func (q *Queue) Send(message Message) error {

	err := q.emailSender.Send(message)
	if err == nil {
		return nil
	}

	return q.store.Add(message)
}

// Get  get a specific message
func (q *Queue) Get(ID int64) (*Message, error) {

	msgPtr, err := q.store.Get(ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return msgPtr, nil
}

// List get list of messages
func (q *Queue) List() ([]*Message, error) {

	msgPtrs, err := q.store.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	return msgPtrs, nil
}

// Delete delete a specific message
func (q *Queue) Delete(ID int64) error {
	return q.store.Delete(ID)
}

func (q *Queue) Stop() {
	if q.cron != nil {
		q.cron.Stop()
	}
}
