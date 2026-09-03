package notifications

import (
	"database/sql"
)

type Notification struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	Type       string `json:"type"`
	ServerID   *int   `json:"server_id"`
	ChannelID  *int   `json:"channel_id"`
	MessageID  *int   `json:"message_id"`
	FromUserID *int   `json:"from_user_id"`
	Read       bool   `json:"read"`
	CreatedAt  string `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateMention(
	userID int,
	serverID int,
	channelID int,
	messageID int,
	fromUserID int,
) (int, error) {
	result, err := r.db.Exec(`
		INSERT INTO notifications (
			user_id,
			type,
			server_id,
			channel_id,
			message_id,
			from_user_id
		)
		VALUES (?, 'mention', ?, ?, ?, ?)
	`,
		userID,
		serverID,
		channelID,
		messageID,
		fromUserID,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (r *Repository) GetNotifications(
	userID int,
) ([]Notification, error) {
	rows, err := r.db.Query(`
		SELECT
			id,
			user_id,
			type,
			server_id,
			channel_id,
			message_id,
			from_user_id,
			read,
			created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
	`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]Notification, 0)

	for rows.Next() {
		var notification Notification

		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.ServerID,
			&notification.ChannelID,
			&notification.MessageID,
			&notification.FromUserID,
			&notification.Read,
			&notification.CreatedAt,
		); err != nil {
			return nil, err
		}

		notifications = append(
			notifications,
			notification,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}
