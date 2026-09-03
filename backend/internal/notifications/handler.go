package notifications

import (
	"encoding/json"
	"net/http"

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
