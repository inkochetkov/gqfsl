package gqfsl

type Queue struct {
	config Config

	emailServ *emailServer
	sql       *sqLite
	cron      *cron
}

func New(config Config) (*Queue, error) {

	emailServer, err := connEmailServer(config)
	if err != nil {
		return nil, err
	}

	sql, err := startSQL(config)
	if err != nil {
		return nil, err
	}

	cron, err := startCron(config)
	if err != nil {
		return nil, err
	}

	return &Queue{config, emailServer, sql, cron}, nil
}

// Add delayed dispatch
func (q *Queue) Add(message Message) error {
	return nil
}

// Send immediate dispatch
func (q *Queue) Send(message Message) error {
	return nil
}

// Get  get a specific message
func (q *Queue) Get(ID int64) (Message, error) {
	return Message{}, nil
}

// List get list of messages
func (q *Queue) List() ([]Message, error) {
	return nil, nil
}

// Delete delete a specific message
func (q *Queue) Delete(ID int64) error {
	return nil
}
