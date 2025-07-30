package gqfsl

import (
	"context"
	"database/sql"
	"fmt"
	p "path"
	"strconv"
	"sync"
	"time"

	"github.com/inkochetkov/exist"
	"github.com/inkochetkov/ujt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type messageDB struct {
	ID int64 `db:"id"`

	From     string `db:"from"`
	To       string `db:"to"`
	Subject  string `db:"subject"`
	TypeBody string `db:"type_body"`
	Body     string `db:"body"`

	Status []byte `db:"status"`
}

type sqLite struct {
	mu     sync.Mutex
	conn   *sqlx.DB
	config Config
}

const (
	getEmail = `
	SELECT
		*
	FROM
		email
	WHERE
		id = $1		
	`
	listEmail = `
	SELECT
		*
	FROM
		email
	WHERE
		count_try_send < $1
		AND time_send is null
	`
	deleteEmail = `
		Delete 
			email
		WHERE 
			id = $1`
	initEmailFirst = `
	INSERT INTO email (
		from, 
		to, 
		subject,
		body_type,
		body,
		time_registry, 
	) VALUES ($1, $2, $3, $4, $5, $6)`
)

func (s *sqLite) Update(m Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: implementation
	return nil
}

func (s *sqLite) Get(ID int64) (*Message, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Sql.Timeout)
	defer cancel()

	mes := &messageDB{}
	err := s.conn.QueryRowxContext(ctx, getEmail, ID).StructScan(mes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertFromDB(mes)
}

func (s *sqLite) List() ([]*Message, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Sql.Timeout)
	defer cancel()

	rows, err := s.conn.QueryxContext(ctx, listEmail, s.config.Cron.CountTry)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var messages []*Message

	for rows.Next() {
		mes := &messageDB{}
		err := rows.StructScan(mes)
		if err != nil {
			return nil, err
		}
		m, err := convertFromDB(mes)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)

	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *sqLite) Delete(ID int64) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	statement, err := s.conn.Prepare(deleteEmail)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(ID)
	return err

}

func (s *sqLite) Add(m Message) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	statement, err := s.conn.Prepare(initEmailFirst)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(m.From, m.To, m.Subject, m.TypeBody, m.Body, time.Now().Unix())
	return err
}

func startSQL(config Config) (*sqLite, error) {

	// check this file base
	dataSourcePath, err := checkFileBD(config.Sql.Path, config.Sql.BaseName)
	if err != nil {
		return nil, err
	}
	// connection to base
	conn, err := connectSqLite(dataSourcePath)
	if err != nil {
		return nil, err
	}
	// check migration
	err = migration(conn)
	if err != nil {
		return nil, err
	}

	return &sqLite{conn: conn, config: config}, nil
}

func migration(conn *sqlx.DB) error {

	_, err := conn.Exec(`
CREATE TABLE IF NOT EXISTS email (

    id INTEGER PRIMARY KEY AUTOINCREMENT
        NOT NULL,
    from TEXT NOT NULL,
    to TEXT NOT NULL,
    subject TEXT NOT NULL,
    body_type TEXT NOT NULL,
    body TEXT NOT NULL,
    time_registry INTEGER NOT NULL,

	count_try_send INTEGER,
    time_send INTEGER,
    err TEXT

);`)
	if err != nil {
		return fmt.Errorf("fail create table, %w", err)
	}

	return nil
}

// checkFileBD - check dir and file exist
func checkFileBD(path, fileName string) (string, error) {

	url := p.Join(path, fileName)

	if ok := exist.CheckFile(url); ok {
		return url, nil
	}

	_, err := exist.InitDirFile(path, fileName)
	if err != nil {
		return "", fmt.Errorf("fail create path base: path - %s, fileName - %s, %w", path, fileName, err)
	}

	return url, nil
}

func connectSqLite(dataSourcePath string) (*sqlx.DB, error) {
	return sqlx.Open("sqlite3", dataSourcePath)
}

func convertFromDB(mes *messageDB) (*Message, error) {

	m := &Message{
		ID:       strconv.FormatInt(mes.ID, 10),
		From:     mes.From,
		To:       mes.To,
		Subject:  mes.Subject,
		TypeBody: mes.TypeBody,
		Body:     mes.TypeBody,
	}

	status, err := ujt.UnmarshaledJSONToMap(mes.Status)
	if err != nil {
		return nil, err
	}

	m.Status = status

	return m, nil
}
