package servers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"agora/internal/users"
)

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
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

	data.Name = strings.TrimSpace(data.Name)

	if data.Name == "" {
		http.Error(w, "Server name is required", http.StatusBadRequest)
		return
	}

	if len(data.Name) > 50 {
		http.Error(w, "Server name too long", http.StatusBadRequest)
		return
	}

	ServerID, err := h.repo.Create(data.Name, userID)
	if err != nil {
		http.Error(w, "Could not create server", http.StatusInternalServerError)
		return
	}

	server := Server{
		ID:      int(ServerID),
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

	servers, err := h.repo.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := ServersResponse{
		Servers: servers,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
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

	member, err := h.repo.IsMember(userID, serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	server, err := h.repo.GetByID(serverID)
	if err == sql.ErrNoRows {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(server); err != nil {
		return
	}
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

	member, err := h.repo.IsMember(userID, serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	members, err := h.repo.GetMembers(serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := MembersResponse{Members: members}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

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

	server, err := h.repo.GetByID(serverID)
	if err == sql.ErrNoRows {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if server.OwnerID == userID {
		http.Error(
			w,
			"Server owner cannot leave the server",
			http.StatusBadRequest,
		)
		return
	}

	removed, err := h.repo.RemoveMember(userID, serverID)
	if err != nil {
		http.Error(w, "Could not leave server", http.StatusInternalServerError)
		return
	}
	if !removed {
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

	member, err := h.repo.IsMember(userID, serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !member {
		http.Error(w, "You are not a member of this server", http.StatusForbidden)
		return
	}

	code, err := generateInviteCode()
	if err != nil {
		http.Error(w, "Could not generate invite", http.StatusInternalServerError)
		return
	}

	if err := h.repo.CreateInvite(code, serverID, userID); err != nil {
		http.Error(w, "Could not create invite", http.StatusInternalServerError)
		return
	}

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

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
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

	serverID, err := h.repo.GetServerIDByInviteCode(code)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid invite", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.repo.AddMember(userID, serverID); err != nil {
		http.Error(w, "Could not join server", http.StatusInternalServerError)
		return
	}

	server, err := h.repo.GetByID(serverID)
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

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("userID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data UpdateMemberRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if data.Role != RoleAdmin && data.Role != RoleMember {
		http.Error(
			w,
			"Role must be admin or member",
			http.StatusBadRequest,
		)
		return
	}

	// The person making the request must be an owner.
	requesterRole, err := h.repo.GetMemberRole(userID, serverID)

	if err == sql.ErrNoRows {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if requesterRole != RoleOwner {
		http.Error(
			w,
			"Only the server owner can change member roles",
			http.StatusForbidden,
		)
		return
	}

	// Make sure the target user is actually a member.
	targetRole, err := h.repo.GetMemberRole(targetUserID, serverID)

	if err == sql.ErrNoRows {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// The owner cannot be changed through this endpoint.
	if targetRole == RoleOwner {
		http.Error(
			w,
			"The server owner cannot be changed",
			http.StatusForbidden,
		)
		return
	}

	if err := h.repo.UpdateMemberRole(
		serverID,
		targetUserID,
		data.Role,
	); err != nil {
		http.Error(
			w,
			"Could not update member role",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("userID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the role of the person making the request.
	requesterRole, err := h.repo.GetMemberRole(userID, serverID)

	if err == sql.ErrNoRows {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get the role of the person being removed.
	targetRole, err := h.repo.GetMemberRole(targetUserID, serverID)

	if err == sql.ErrNoRows {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// The owner cannot be removed.
	if targetRole == RoleOwner {
		http.Error(
			w,
			"The server owner cannot be removed",
			http.StatusForbidden,
		)
		return
	}

	// Owner can remove anyone except the owner.
	if requesterRole == RoleOwner {
		// Allowed.
	} else if requesterRole == RoleAdmin {
		// Admins can only remove regular members.
		if targetRole != RoleMember {
			http.Error(
				w,
				"Admins can only remove members",
				http.StatusForbidden,
			)
			return
		}
	} else {
		http.Error(
			w,
			"You do not have permission to remove members",
			http.StatusForbidden,
		)
		return
	}

	removed, err := h.repo.RemoveMember(serverID, targetUserID)

	if err != sql.ErrNoRows || !removed {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	if err != nil {

		http.Error(
			w,
			"Could not remove member",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
