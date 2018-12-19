package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/sergeiten/golearn"
)

// DBService mysql database struct.
type DBService struct {
	DB *sql.DB
}

// Open returns opened mysql connection object.
func Open(connection string) (*sql.DB, error) {
	db, err := sql.Open("mysql", connection)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Exists returns boolean if row exists in table.
func Exists(DB *sql.DB, sql string, args []interface{}) bool {
	var exists bool

	sql = fmt.Sprintf("SELECT exists (%s)", sql)

	err := DB.QueryRow(sql, args...).Scan(&exists)

	if err != nil {
		log.Fatal(err)
	}

	return exists
}

// RandomRow returns random row from database.
func (s *DBService) RandomRow() (golearn.Row, error) {
	q := golearn.Row{}

	s.DB.QueryRow("SELECT id, word, translate FROM words ORDER BY RAND() LIMIT 1").Scan(&q.ID, &q.Word, &q.Translate)

	return q, nil
}

// RandomAnswers returns list of random rows for building answer options.
func (s *DBService) RandomAnswers(q golearn.Row, limit int) ([]golearn.Row, error) {
	fmt.Printf("%+v", q)
	answers := []golearn.Row{}

	rows, err := s.DB.Query(`
	(SELECT id, word, translate FROM words WHERE id <> ? ORDER BY RAND() LIMIT ?)
	UNION ALL
	(SELECT id, word, translate FROM words WHERE id = ?)
	ORDER BY RAND()
	`, q.ID, limit-1, q.ID)

	if err != nil {
		return answers, err
	}

	defer rows.Close()

	for rows.Next() {
		a := golearn.Row{}
		err := rows.Scan(&a.ID, &a.Word, &a.Translate)
		if err != nil {
			return answers, err
		}
		answers = append(answers, a)
	}

	return answers, nil
}

// SetState sets state.
func (s *DBService) SetState(state golearn.State) error {
	return nil
}

// GetState returns state by passed user id.
func (s *DBService) GetState(userKey string) (golearn.State, error) {
	return golearn.State{}, nil
}

// ResetState resets state by passed user id.
func (s *DBService) ResetState(userKey string) error {
	if userKey == "" {
		return errors.New("user key is empty")
	}

	_, err := s.DB.Exec("UPDATE users SET last_word_id=0, last_word_set='' WHERE key=?", userKey)
	return err
}
