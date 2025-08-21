package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/devmuggs/youtube-music-continuity/backend/internal/database"
	"github.com/devmuggs/youtube-music-continuity/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db *database.DB
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}
	return secret
}

// NewHandler creates a new auth handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, err := h.db.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.db.CreateUser(req.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.db.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSession retrieves user's current session
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session, err := h.db.GetSession(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session == nil {
		// Return empty session
		session = &models.Session{
			UserID:       userID,
			VideoID:      "",
			PlaybackTime: 0,
			IsPlaying:    false,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// UpdateSession updates user's current session
func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.SessionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update session
	session, err := h.db.UpsertSession(userID, req.VideoID, req.PlaybackTime, req.IsPlaying)
	if err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// generateToken creates a JWT token for the user
func generateToken(userID int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "userID", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	}
}

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(ctx context.Context) int {
	if userID, ok := ctx.Value("userID").(int); ok {
		return userID
	}
	return 0
}

// ValidateToken validates a JWT token and returns the user ID
func ValidateToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token")
}

// ParseUserIDFromToken parses user ID from token string without full validation
func ParseUserIDFromToken(tokenString string) (int, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token claims")
}