package gqfsl

type Queue struct {
	config Config

	emailServer *emailServer
	sql         *sqLite
	cron        *cron
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

	cron, err := startCron(config)
	if err != nil {
		return nil, err
	}

	return &Queue{config, emailServer, sql, cron}, nil
}

// Add delayed dispatch
func (q *Queue) Add(message Message) error {
	return q.sql.Add(message)
}

// Send immediate dispatch
func (q *Queue) Send(message Message) error {

	err := q.emailServer.Send(message)
	if err == nil {
		return nil
	}

	return q.sql.Add(message)
}

// Get  get a specific message
func (q *Queue) Get(ID int64) (Message, error) {
	// TODO: implementation
	return Message{}, nil
}

// List get list of messages
func (q *Queue) List() ([]Message, error) {
	// TODO: implementation
	return nil, nil
}

// Delete delete a specific message
func (q *Queue) Delete(ID int64) error {
	return q.sql.Delete(ID)
}
