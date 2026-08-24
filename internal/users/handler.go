package users

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	usernameLen := len(data.Username)

	if usernameLen < 3 {
		http.Error(w, "Username too short", http.StatusBadRequest)
		return
	}

	if usernameLen > 20 {
		http.Error(w, "Username too long", http.StatusBadRequest)
		return
	}

	passwordLen := len(data.Password)

	if passwordLen < 3 {
		http.Error(w, "Password too short", http.StatusBadRequest)
		return
	}

	if passwordLen > 20 {
		http.Error(w, "Password too long", http.StatusBadRequest)
		return
	}

	passwordHash, err := HashPassword(data.Password)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	result, err := h.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		data.Username,
		passwordHash,
	)

	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Could not get user ID", http.StatusInternalServerError)
		return
	}

	user := User{
		ID:       int(id),
		Username: data.Username,
	}

	response := UserResponse{
		Message: "User created",
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var (
		userID       int
		username     string
		passwordHash string
	)

	err = h.db.QueryRow(
		"SELECT id, username, password_hash FROM users WHERE username = ?",
		data.Username,
	).Scan(&userID, &username, &passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	valid, err := VerifyPassword(data.Password, passwordHash)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	sessionID, err := createSession(h.db, userID)
	if err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})

	response := UserResponse{
		Message: "Login successful",
		User: User{
			ID:       userID,
			Username: username,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = deleteSession(h.db, cookie.Value)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int

		err = h.db.QueryRow(
			`SELECT user_id
			 FROM sessions
			 WHERE id = ?
			 AND expires_at > datetime('now')`,
			cookie.Value,
		).Scan(&userID)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := setUserIDContext(r.Context(), userID)

		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User

	err := h.db.QueryRow(
		"SELECT id, username FROM users WHERE id = ?",
		userID,
	).Scan(
		&user.ID,
		&user.Username,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(user); err != nil {
		return
	}
}
