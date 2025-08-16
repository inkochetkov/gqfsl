package gqfsl

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	p "path"
	"strconv"
	"sync"
	"time"

	"github.com/inkochetkov/exist"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type messageDB struct {
	ID int64 `db:"id"`

	From     string `db:"from"`
	To       string `db:"to"`
	Subject  string `db:"subject"`
	BodyType string `db:"body_type"`
	Body     string `db:"body"`

	Status []byte `db:"status"`
}

func migration(conn *sqlx.DB, baseName string) error {

	_, err := conn.Exec(`
CREATE TABLE IF NOT EXISTS ` + baseName + ` (

    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,

    "from" TEXT NOT NULL,
    "to" TEXT NOT NULL,
    "subject" TEXT NOT NULL,
    "body_type" TEXT NOT NULL,
    "body" TEXT NOT NULL,

	"status" BLOB
);`)
	if err != nil {
		return fmt.Errorf("fail create table, %w", err)
	}

	return nil
}

type sqLite struct {
	mu     sync.Mutex
	conn   *sqlx.DB
	config Config
}

func (s *sqLite) Update(m Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mesDB, err := convertToDB(&m)
	if err != nil {
		return err
	}

	statement, err := s.conn.Prepare(`
	UPDATE 
		` + s.config.Sql.BaseName + `
	SET 
		"from" = ?,
		"to" = ?,
		subject = ?,
		body_type = ?,
		body = ?,
		status = ?
	WHERE 
		id = ?`)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(
		mesDB.From,
		mesDB.To,
		mesDB.Subject,
		mesDB.BodyType,
		mesDB.Body,
		mesDB.Status,
		mesDB.ID,
	)
	return err
}

func (s *sqLite) Get(ID int64) (*Message, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Sql.Timeout)
	defer cancel()

	mes := &messageDB{}
	err := s.conn.QueryRowxContext(ctx, `
	 	SELECT
			*
		FROM
			`+s.config.Sql.BaseName+`
		WHERE
			id = ?`, ID).StructScan(mes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return convertFromDB(mes)
}

func (s *sqLite) List() ([]*Message, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Sql.Timeout)
	defer cancel()

	rows, err := s.conn.QueryxContext(ctx, "SELECT * FROM "+s.config.Sql.BaseName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	var messages []*Message

	for rows.Next() {
		mes := &messageDB{}
		err := rows.StructScan(mes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		m, err := convertFromDB(mes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return messages, nil
}

func (s *sqLite) Delete(ID int64) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	statement, err := s.conn.Prepare(`
		Delete FROM
			` + s.config.Sql.BaseName + `
		WHERE 
			id = ?`)
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

	// Добавляем время регистрации
	if m.Status == nil {
		m.Status = make(map[string]any)
	}
	m.Status["time_registry"] = time.Now().Unix()
	m.Status["count_try_send"] = 0 // Инициализируем счетчик попыток

	mesDB, err := convertToDB(&m)
	if err != nil {
		return err
	}

	statement, err := s.conn.Prepare(`
	INSERT INTO ` + s.config.Sql.BaseName + ` (
		"from", 
		"to", 
		subject,
		body_type,
		body,
		status
	) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(
		mesDB.From,
		mesDB.To,
		mesDB.Subject,
		mesDB.BodyType,
		mesDB.Body,
		mesDB.Status,
	)
	return err
}

func startSQL(config Config) (*sqLite, error) {

	// check this file base
	dataSourcePath, err := checkFileBD(config.Sql.Path, config.Sql.BaseName+".sqlite")
	if err != nil {
		return nil, err
	}
	// connection to base
	conn, err := connectSqLite(dataSourcePath)
	if err != nil {
		return nil, err
	}
	// check migration
	err = migration(conn, config.Sql.BaseName)
	if err != nil {
		return nil, err
	}

	return &sqLite{conn: conn, config: config}, nil
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
	return sqlx.Open("sqlite", dataSourcePath)
}

func convertFromDB(mes *messageDB) (*Message, error) {
	m := &Message{
		ID:       strconv.FormatInt(mes.ID, 10),
		From:     mes.From,
		To:       mes.To,
		Subject:  mes.Subject,
		BodyType: mes.BodyType,
		Body:     mes.Body,
	}

	// Десериализация статуса из JSON
	if len(mes.Status) > 0 {
		status := make(map[string]any)
		if err := json.Unmarshal(mes.Status, &status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal status: %w", err)
		}
		m.Status = status
	} else {
		m.Status = make(map[string]any)
	}

	return m, nil
}

func convertToDB(mes *Message) (*messageDB, error) {

	result := &messageDB{
		From:     mes.From,
		To:       mes.To,
		Subject:  mes.Subject,
		BodyType: mes.BodyType,
		Body:     mes.Body,
	}

	if mes.ID != "" {
		var err error
		result.ID, err = strconv.ParseInt(mes.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID format: %w", err)
		}

	}

	// Добавляем обязательные поля в статус
	if mes.Status == nil {
		mes.Status = make(map[string]any)
	}

	// Сериализация статуса в JSON
	statusJSON, err := json.Marshal(mes.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status: %w", err)
	}

	if len(statusJSON) > 0 {
		result.Status = statusJSON
	}

	return result, nil
}
func (s *sqLite) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
}
