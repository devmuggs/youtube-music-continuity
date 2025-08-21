package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/devmuggs/youtube-music-continuity/backend/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

// NewDB creates a new database connection and initializes tables
func NewDB() (*DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./continuity.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbWrapper := &DB{db}
	if err := dbWrapper.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return dbWrapper, nil
}

// createTables initializes the database schema
func (db *DB) createTables() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	sessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		video_id TEXT NOT NULL,
		playback_time REAL NOT NULL DEFAULT 0,
		is_playing BOOLEAN NOT NULL DEFAULT FALSE,
		updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if _, err := db.Exec(sessionsTable); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	return nil
}

// CreateUser creates a new user
func (db *DB) CreateUser(email, passwordHash string) (*models.User, error) {
	query := `INSERT INTO users (email, password) VALUES (?, ?) RETURNING id, created`
	var user models.User
	user.Email = email
	user.Password = passwordHash

	err := db.QueryRow(query, email, passwordHash).Scan(&user.ID, &user.Created)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password, created FROM users WHERE email = ?`
	var user models.User

	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, email, password, created FROM users WHERE id = ?`
	var user models.User

	err := db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Password, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpsertSession creates or updates a user's session
func (db *DB) UpsertSession(userID int, videoID string, playbackTime float64, isPlaying bool) (*models.Session, error) {
	query := `
	INSERT INTO sessions (user_id, video_id, playback_time, is_playing, updated)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(user_id) DO UPDATE SET
		video_id = excluded.video_id,
		playback_time = excluded.playback_time,
		is_playing = excluded.is_playing,
		updated = CURRENT_TIMESTAMP
	RETURNING id, user_id, video_id, playback_time, is_playing, updated`

	var session models.Session
	err := db.QueryRow(query, userID, videoID, playbackTime, isPlaying).Scan(
		&session.ID, &session.UserID, &session.VideoID, 
		&session.PlaybackTime, &session.IsPlaying, &session.Updated,
	)

	if err != nil {
		// Fallback for SQLite which doesn't support ON CONFLICT in this way
		// First try to update existing session
		updateQuery := `UPDATE sessions SET video_id = ?, playback_time = ?, is_playing = ?, updated = CURRENT_TIMESTAMP WHERE user_id = ?`
		result, updateErr := db.Exec(updateQuery, videoID, playbackTime, isPlaying, userID)
		
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update session: %w", updateErr)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			// No existing session, insert new one
			insertQuery := `INSERT INTO sessions (user_id, video_id, playback_time, is_playing) VALUES (?, ?, ?, ?)`
			result, insertErr := db.Exec(insertQuery, userID, videoID, playbackTime, isPlaying)
			if insertErr != nil {
				return nil, fmt.Errorf("failed to insert session: %w", insertErr)
			}
			sessionID, _ := result.LastInsertId()
			session.ID = int(sessionID)
		}

		// Get the updated/inserted session
		getQuery := `SELECT id, user_id, video_id, playback_time, is_playing, updated FROM sessions WHERE user_id = ?`
		err = db.QueryRow(getQuery, userID).Scan(
			&session.ID, &session.UserID, &session.VideoID,
			&session.PlaybackTime, &session.IsPlaying, &session.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get session after upsert: %w", err)
		}
	}

	return &session, nil
}

// GetSession retrieves a user's current session
func (db *DB) GetSession(userID int) (*models.Session, error) {
	query := `SELECT id, user_id, video_id, playback_time, is_playing, updated FROM sessions WHERE user_id = ?`
	var session models.Session

	err := db.QueryRow(query, userID).Scan(
		&session.ID, &session.UserID, &session.VideoID,
		&session.PlaybackTime, &session.IsPlaying, &session.Updated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}