package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"

	"agora/internal/users"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) GetNotifications(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	notifications, err := h.repo.GetNotifications(userID)
	if err != nil {
		http.Error(
			w,
			"Could not get notifications",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(
		notifications,
	); err != nil {
		return
	}
}

func (h *Handler) MarkRead(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	notificationID, err := strconv.Atoi(
		r.PathValue("id"),
	)
	if err != nil {
		http.Error(
			w,
			"Invalid notification ID",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.repo.MarkRead(
		userID,
		notificationID,
	); err != nil {
		http.Error(
			w,
			"Could not mark notification as read",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkChannelRead(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := users.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channelID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.MarkChannelRead(userID, channelID); err != nil {
		http.Error(
			w,
			"Could not mark channel notifications as read",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
