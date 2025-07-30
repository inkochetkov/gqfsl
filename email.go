package gqfsl

import "gopkg.in/gomail.v2"

type Message struct {
	ID string

	From     string
	To       string
	Subject  string
	TypeBody string // "text/plain" or "text/html", default "text/plain"
	Body     string

	Status map[string]any
}

type emailServer struct {
	conn *gomail.Dialer
}

func connEmailServer(config Config) (*emailServer, error) {
	return &emailServer{conn: &gomail.Dialer{
		Host:     config.Email.Host,
		Port:     config.Email.Port,
		Username: config.Email.Username,
		Password: config.Email.Password,
		SSL:      config.Email.Port == 465,
	}}, nil
}

func (e *emailServer) Send(message Message) error {

	m := gomail.NewMessage()
	m.SetHeader("From", message.From)
	m.SetHeader("To", message.To)
	m.SetHeader("Subject", message.Subject)
	m.SetBody(message.TypeBody, message.Body)

	return e.conn.DialAndSend(m)
}
