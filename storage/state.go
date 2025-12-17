package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// Store provides persistent state management using SQLite.
// Tracks connection requests, accepted connections, sent messages, and daily counters.
// Enables resumption after interruptions by persisting all automation state.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store instance with the provided database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) IncrementDailyCounter(key string) (int, error) {
	todayKey := fmt.Sprintf("%s:%s", key, time.Now().Format("2006-01-02"))
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var val int
	row := tx.QueryRow("SELECT value FROM counters WHERE key = ?", todayKey)
	switch err := row.Scan(&val); err {
	case sql.ErrNoRows:
		val = 1
		_, err = tx.Exec("INSERT INTO counters(key,value) VALUES(?,?)", todayKey, val)
	default:
		if err != nil {
			return 0, err
		}
		val++
		_, err = tx.Exec("UPDATE counters SET value=? WHERE key=?", val, todayKey)
	}
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return val, nil
}

func (s *Store) GetDailyCounter(key string) (int, error) {
	todayKey := fmt.Sprintf("%s:%s", key, time.Now().Format("2006-01-02"))
	var val int
	err := s.db.QueryRow("SELECT value FROM counters WHERE key=?", todayKey).Scan(&val)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return val, err
}

func (s *Store) SaveConnection(profileURL, name, company, status string) error {
	_, err := s.db.Exec(
		`INSERT INTO connections(profile_url,name,company,status,contacted_at) VALUES(?,?,?,?,?)
		 ON CONFLICT(profile_url) DO UPDATE SET status=excluded.status`,
		profileURL, name, company, status, time.Now(),
	)
	return err
}

func (s *Store) MarkAccepted(profileURL string) error {
	_, err := s.db.Exec(`UPDATE connections SET status='accepted', accepted_at=? WHERE profile_url=?`, time.Now(), profileURL)
	return err
}

func (s *Store) SaveMessage(profileURL, body string) error {
	_, err := s.db.Exec(`INSERT INTO messages(profile_url,body,sent_at) VALUES(?,?,?)`, profileURL, body, time.Now())
	if err != nil {
		return err
	}
	_, _ = s.db.Exec(`UPDATE connections SET last_message_at=? WHERE profile_url=?`, time.Now(), profileURL)
	return nil
}

func (s *Store) PendingFollowups() ([]string, error) {
	rows, err := s.db.Query(`SELECT profile_url FROM connections WHERE status='accepted'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}

// RequestedConnections returns connections not yet accepted.
func (s *Store) RequestedConnections() ([]string, error) {
	rows, err := s.db.Query(`SELECT profile_url FROM connections WHERE status='requested'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}

