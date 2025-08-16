package gqfsl

type emailSender interface {
	Send(message Message) error
}

type messageStore interface {
	Add(message Message) error
	Get(id int64) (*Message, error)
	List() ([]*Message, error)
	Update(message Message) error
	Delete(id int64) error
	Close() error
}

type cronService interface {
	Stop()
}
