package users

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ServerBroadcaster interface {
	BroadcastServer(serverID int, data []byte)
}

type Handler struct {
	repo *Repository
	hub  ServerBroadcaster
}

func NewHandler(
	repo *Repository,
	hub ServerBroadcaster,
) *Handler {
	return &Handler{
		repo: repo,
		hub:  hub,
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

	id, err := h.repo.Create(data.Username, passwordHash)
	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	userID := int(id)

	sessionID, err := createSession(h.repo.db, userID)
	if err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
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

	user := User{
		ID:       userID,
		Username: data.Username,
	}

	response := UserResponse{
		Message: "User created",
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
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

	userID, username, passwordHash, err := h.repo.GetByUsername(data.Username)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
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

	sessionID, err := h.repo.CreateSession(userID)
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

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
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

	err = h.repo.DeleteSession(cookie.Value)
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

		userID, err := h.repo.GetUserIDFromSession(cookie.Value)
		if err == sql.ErrNoRows {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err != nil {
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

	user, err := h.repo.GetByID(userID)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(user); err != nil {
		return
	}
}

func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	log.Println("UploadAvatar called")

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	oldAvatarURL, err := h.repo.GetAvatarURL(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "File too large or invalid upload", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Avatar is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Println("uploading file: ", fileHeader.Filename)

	contentType, header, err := detectImageType(file)
	if err != nil {
		http.Error(w, "Could not read avatar", http.StatusBadRequest)
		return
	}

	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		http.Error(
			w,
			"Avatar must be a JPEG, PNG, or WebP image",
			http.StatusBadRequest,
		)
		return
	}

	avatarURL, err := saveAvatar(file, header, contentType)
	if err != nil {
		http.Error(w, "Could not save avatar", http.StatusInternalServerError)
		return
	}

	if err := h.repo.UpdateAvatar(userID, avatarURL); err != nil {
		deleteAvatarFile(avatarURL)

		http.Error(w, "Could not update avatar", http.StatusInternalServerError)
		return
	}

	if oldAvatarURL != "" {
		deleteAvatarFile(oldAvatarURL)
	}

	user, err := h.repo.GetByID(userID)
	if err != nil {
		http.Error(w, "Could not get user", http.StatusInternalServerError)
		return
	}

	eventData, err := json.Marshal(struct {
		Type string `json:"type"`
		Data User   `json:"data"`
	}{
		Type: "user_updated",
		Data: user,
	})

	if err != nil {
		http.Error(w, "Could not create realtime event", http.StatusInternalServerError)
		return
	}

	serverIDs, err := h.repo.GetServerIDsForUser(userID)
	if err != nil {
		http.Error(w, "Could not get user servers", http.StatusInternalServerError)
		return
	}

	for _, serverID := range serverIDs {
		h.hub.BroadcastServer(serverID, eventData)
	}

	response := UserResponse{
		Message: "Avatar updated",
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
