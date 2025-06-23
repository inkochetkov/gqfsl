package gqfsl

import (
	"database/sql"
	"fmt"
	p "path"
	"sync"
	"time"

	"github.com/inkochetkov/exist"
	_ "github.com/mattn/go-sqlite3"
)

type sqLite struct {
	mu   sync.Mutex
	conn *sql.DB
}

const (
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

	return &sqLite{conn: conn}, nil
}

func migration(conn *sql.DB) error {

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

func connectSqLite(dataSourcePath string) (*sql.DB, error) {
	return sql.Open("sqlite3", dataSourcePath)
}
