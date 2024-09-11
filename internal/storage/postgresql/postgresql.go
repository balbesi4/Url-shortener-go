package postgresql

import (
	"Url-shortener-go/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	postgresDriverName        = "postgres"
	uniqueConstraintErrorName = "unique_violation"
	urlNotFoundErrName        = "no_data"
)

type Storage struct {
	db *sql.DB
}

func NewStorage() (*Storage, error) {
	const f = "storage.postgresql.NewStorage"

	port, _ := strconv.Atoi(os.Getenv("DATABASE_PORT"))
	psqlConn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DATABASE_HOST"),
		port,
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
	)

	db, err := sql.Open(postgresDriverName, psqlConn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", f, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", f, err)
	}

	err = execCreateQuery(db, `
		CREATE TABLE IF NOT EXISTS urls (
		    id SERIAL PRIMARY KEY,
		    alias TEXT UNIQUE NOT NULL,
		    full_url TEXT NOT NULL
		);
	`)

	if err != nil {
		return nil, err
	}

	err = execCreateQuery(db, `
		CREATE INDEX IF NOT EXISTS alias_idx ON urls (alias);
	`)

	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func execCreateQuery(db *sql.DB, query string) error {
	const f = "storage.postgresql.execCreateQuery"

	stmt, err := db.Prepare(query)

	if err != nil {
		return fmt.Errorf("%s: %w", f, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", f, err)
	}

	return nil
}

func (s *Storage) SaveURL(fullURL string, alias string) (int64, error) {
	const f = "storage.postgresql.saveURL"

	row := s.db.QueryRow(`INSERT INTO urls (alias, full_url) VALUES ($1, $2) RETURNING id`, alias, fullURL)

	var id int64
	err := row.Scan(&id)
	if err != nil {
		var psqlErr *pq.Error
		if errors.As(err, &psqlErr) && psqlErr.Code.Name() == uniqueConstraintErrorName {
			return 0, fmt.Errorf("%s: %w", f, storage.ErrUrlExists)
		}
		return 0, fmt.Errorf("%s: %w", f, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const f = "storage.postgresql.getURL"

	row := s.db.QueryRow(`SELECT full_url FROM urls WHERE alias = $1`, alias)

	var fullURL string
	err := row.Scan(&fullURL)
	if err != nil {
		var psqlErr *pq.Error
		if errors.As(err, &psqlErr) && psqlErr.Code.Name() == urlNotFoundErrName {
			return "", storage.ErrUrlNotFound
		}
		return "", fmt.Errorf("%s: %w", f, err)
	}

	return fullURL, nil
}

func (s *Storage) DeleteURL(alias string) (int64, error) {
	const f = "storage.postgresql.deleteURL"

	row := s.db.QueryRow(`DELETE FROM urls WHERE alias = $1 RETURNING id`, alias)

	var id int64
	err := row.Scan(&id)

	if err != nil {
		var psqlErr *pq.Error
		if errors.As(err, &psqlErr) && psqlErr.Code.Name() == urlNotFoundErrName {
			return 0, storage.ErrUrlNotFound
		}
		return 0, fmt.Errorf("%s: %w", f, err)
	}

	return id, nil
}
