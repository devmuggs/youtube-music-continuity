package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID       int       `json:"id" db:"id"`
	Email    string    `json:"email" db:"email"`
	Password string    `json:"-" db:"password"` // Password hash, never sent in JSON
	Created  time.Time `json:"created" db:"created"`
}

// Session represents a user's current playback session
type Session struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	VideoID      string    `json:"video_id" db:"video_id"`
	PlaybackTime float64   `json:"playback_time" db:"playback_time"` // In seconds
	IsPlaying    bool      `json:"is_playing" db:"is_playing"`
	Updated      time.Time `json:"updated" db:"updated"`
}

// AuthRequest represents login/register request payload
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents login response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// SessionUpdateRequest represents session update request
type SessionUpdateRequest struct {
	VideoID      string  `json:"video_id"`
	PlaybackTime float64 `json:"playback_time"`
	IsPlaying    bool    `json:"is_playing"`
}

// WebSocketMessage represents messages sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocket message types
const (
	MessageTypeSessionUpdate = "session_update"
	MessageTypeError         = "error"
	MessageTypeAuth          = "auth"
)