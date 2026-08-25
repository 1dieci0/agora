package channels

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"agora/internal/servers"
	"agora/internal/users"
)

type Handler struct {
	repo       *Repository
	serverRepo *servers.Repository
}

func NewHandler(
	repo *Repository,
	serverRepo *servers.Repository,
) *Handler {
	return &Handler{
		repo:       repo,
		serverRepo: serverRepo,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("serverID"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data CreateChannelRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data.Name = strings.TrimSpace(data.Name)

	if data.Name == "" {
		http.Error(w, "Channel name is required", http.StatusBadRequest)
		return
	}

	if len(data.Name) > 50 {
		http.Error(w, "Channel name too long", http.StatusBadRequest)
		return
	}

	if data.Type != "text" && data.Type != "voice" {
		http.Error(w, "Invalid channel type", http.StatusBadRequest)
		return
	}

	allowed, err := h.serverRepo.HasAdminPermission(userID, serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !allowed {
		http.Error(w, "You do not have permission to create channels", http.StatusForbidden)
		return
	}

	channelID, err := h.repo.Create(
		serverID,
		data.Name,
		data.Type,
	)

	if err != nil {
		http.Error(w, "Could not create channel", http.StatusInternalServerError)
		return
	}

	channel := Channel{
		ID:       int(channelID),
		ServerID: serverID,
		Name:     data.Name,
		Type:     data.Type,
	}

	response := ChannelResponse{
		Message: "Channel created",
		Channel: channel,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) GetChannels(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.Atoi(r.PathValue("serverID"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	member, err := h.serverRepo.IsMember(userID, serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	channels, err := h.repo.GetByServerID(serverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := ChannelsResponse{
		Channels: channels,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	channelID, err := strconv.Atoi(r.PathValue("channelID"))
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel, err := h.repo.GetByID(channelID)

	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	member, err := h.serverRepo.IsMember(userID, channel.ServerID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := ChannelResponse{
		Channel: channel,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	channelID, err := strconv.Atoi(r.PathValue("channelID"))
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel, err := h.repo.GetByID(channelID)

	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	allowed, err := h.serverRepo.HasAdminPermission(
		userID,
		channel.ServerID,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !allowed {
		http.Error(
			w,
			"You do not have permission to edit this channel",
			http.StatusForbidden,
		)
		return
	}

	var data UpdateChannelRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data.Name = strings.TrimSpace(data.Name)

	if data.Name == "" {
		http.Error(w, "Channel name is required", http.StatusBadRequest)
		return
	}

	if len(data.Name) > 50 {
		http.Error(w, "Channel name too long", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateName(channelID, data.Name); err != nil {
		http.Error(
			w,
			"Could not update channel",
			http.StatusInternalServerError,
		)
		return
	}

	channel.Name = data.Name

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(ChannelResponse{
		Message: "Channel updated",
		Channel: channel,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	channelID, err := strconv.Atoi(r.PathValue("channelID"))
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel, err := h.repo.GetByID(channelID)

	if err == sql.ErrNoRows {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	allowed, err := h.serverRepo.HasAdminPermission(
		userID,
		channel.ServerID,
	)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !allowed {
		http.Error(
			w,
			"You do not have permission to delete this channel",
			http.StatusForbidden,
		)
		return
	}

	if err := h.repo.Delete(channelID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Channel not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"Could not delete channel",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
