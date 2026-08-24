package servers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"agora/internal/users"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data CreateServerRequest

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	nameLen := len(data.Name)

	if nameLen < 1 {
		http.Error(w, "Server name is required", http.StatusBadRequest)
		return
	}

	if nameLen > 50 {
		http.Error(w, "Server name too long", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(
		"INSERT INTO servers (name, owner_id) VALUES (?, ?)",
		data.Name,
		userID,
	)

	if err != nil {
		http.Error(w, "Could not create server", http.StatusInternalServerError)
		return
	}

	serverID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Could not get server ID", http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec(
		"INSERT INTO server_members (server_id, user_id) VALUES (?, ?)",
		serverID,
		userID,
	)

	if err != nil {
		http.Error(w, "Could not add owner to server", http.StatusInternalServerError)
		return
	}

	server := Server{
		ID:      int(serverID),
		Name:    data.Name,
		OwnerID: userID,
	}

	response := ServerResponse{
		Message: "Server created",
		Server:  server,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) GetServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(
		`SELECT s.id, s.name, s.owner_id
		 FROM servers s
		 JOIN server_members sm ON sm.server_id = s.id
		 WHERE sm.user_id = ?
		 ORDER BY s.id`,
		userID,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	servers := make([]Server, 0)

	for rows.Next() {
		var server Server

		err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.OwnerID,
		)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := ServersResponse{
		Servers: servers,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var server Server

	err = h.db.QueryRow(
		`SELECT s.id, s.name, s.owner_id
		 FROM servers s
		 JOIN server_members sm
		   ON sm.server_id = s.id
		 WHERE s.id = ?
		   AND sm.user_id = ?`,
		serverID,
		userID,
	).Scan(
		&server.ID,
		&server.Name,
		&server.OwnerID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Server not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(server)
}

func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var exists bool

	err = h.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM server_members
			WHERE server_id = ?
			AND user_id = ?
		)`,
		serverID,
		userID,
	).Scan(&exists)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	rows, err := h.db.Query(
		`SELECT u.id, u.username
		 FROM users u
		 JOIN server_members sm
		   ON sm.user_id = u.id
		 WHERE sm.server_id = ?
		 ORDER BY u.username`,
		serverID,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	members := make([]Member, 0)

	for rows.Next() {
		var member Member

		err := rows.Scan(
			&member.ID,
			&member.Username,
		)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := MembersResponse{
		Members: members,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
// 	serverID, err := strconv.Atoi(r.PathValue("id"))
// 	if err != nil {
// 		http.Error(w, "Invalid server ID", http.StatusBadRequest)
// 		return
// 	}

// 	userID, ok := users.UserIDFromContext(r.Context())
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	var exists bool

// 	err = h.db.QueryRow(
// 		"SELECT EXISTS(SELECT 1 FROM servers WHERE id = ?)",
// 		serverID,
// 	).Scan(&exists)

// 	if err != nil {
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	if !exists {
// 		http.Error(w, "Server not found", http.StatusNotFound)
// 		return
// 	}

// 	_, err = h.db.Exec(
// 		`INSERT OR IGNORE INTO server_members (server_id, user_id)
// 		VALUES (?, ?)`,
// 		serverID,
// 		userID,
// 	)

// 	if err != nil {
// 		http.Error(w, "Could not join server", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }

func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var ownerID int

	err = h.db.QueryRow(
		"SELECT owner_id FROM servers WHERE id = ?",
		serverID,
	).Scan(&ownerID)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Server not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if ownerID == userID {
		http.Error(
			w,
			"Server owner cannot leave the server",
			http.StatusBadRequest,
		)
		return
	}

	result, err := h.db.Exec(
		`DELETE FROM server_members
		 WHERE server_id = ?
		 AND user_id = ?`,
		serverID,
		userID,
	)

	if err != nil {
		http.Error(w, "Could not leave server", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		http.Error(w, "You are not a member of this server", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var memberExists bool

	err = h.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM server_members
			WHERE server_id = ?
			AND user_id = ?
		)`,
		serverID,
		userID,
	).Scan(&memberExists)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !memberExists {
		http.Error(w, "You are not a member of this server", http.StatusForbidden)
		return
	}

	code, err := generateInviteCode()
	if err != nil {
		http.Error(w, "Could not generate invite", http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec(
		`INSERT INTO invites (code, server_id, created_by)
		 VALUES (?, ?, ?)`,
		code,
		serverID,
		userID,
	)

	if err != nil {
		http.Error(w, "Could not create invite", http.StatusInternalServerError)
		return
	}

	response := CreateInviteResponse{
		Message: "Invite created",
		Code:    code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}

func generateInviteCode() (string, error) {
	bytes := make([]byte, 8)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (h *Handler) JoinInvite(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	if code == "" {
		http.Error(w, "Invalid invite code", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var serverID int

	err := h.db.QueryRow(
		"SELECT server_id FROM invites WHERE code = ?",
		code,
	).Scan(&serverID)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid invite", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec(
		`INSERT OR IGNORE INTO server_members (server_id, user_id)
		 VALUES (?, ?)`,
		serverID,
		userID,
	)

	if err != nil {
		http.Error(w, "Could not join server", http.StatusInternalServerError)
		return
	}

	var server Server

	err = h.db.QueryRow(
		"SELECT id, name, owner_id FROM servers WHERE id = ?",
		serverID,
	).Scan(
		&server.ID,
		&server.Name,
		&server.OwnerID,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := ServerResponse{
		Message: "Joined server",
		Server:  server,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}
