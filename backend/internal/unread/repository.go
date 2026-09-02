package unread

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

type ChannelUnread struct {
	ServerID          int `json:"server_id"`
	ChannelID         int `json:"channel_id"`
	UnreadCount       int `json:"unread_count"`
	LastMessageID     int `json:"last_message_id"`
	LastReadMessageID int `json:"last_read_message_id"`
}

func (r *Repository) GetUnreadChannels(userID int) ([]ChannelUnread, error) {
	rows, err := r.db.Query(`
		SELECT
			c.server_id,
			c.id,
			COUNT(m.id) AS unread_count,
			COALESCE(MAX(m.id), 0) AS last_message_id,
			COALESCE(rs.last_read_message_id, 0) AS last_read_message_id
		FROM channels c
		LEFT JOIN channel_read_state rs
			ON rs.channel_id = c.id
			AND rs.user_id = ?
		LEFT JOIN messages m
			ON m.channel_id = c.id
			AND m.id > COALESCE(rs.last_read_message_id, 0)
			AND m.user_id != ?
		JOIN server_members sm
			ON sm.server_id = c.server_id
			AND sm.user_id = ?
		GROUP BY
			c.server_id,
			c.id,
			rs.last_read_message_id
		HAVING COUNT(m.id) > 0
		ORDER BY MAX(m.id) DESC
	`, userID, userID, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]ChannelUnread, 0)

	for rows.Next() {
		var unread ChannelUnread

		if err := rows.Scan(
			&unread.ServerID,
			&unread.ChannelID,
			&unread.UnreadCount,
			&unread.LastMessageID,
			&unread.LastReadMessageID,
		); err != nil {
			return nil, err
		}

		result = append(result, unread)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) MarkChannelRead(
	userID int,
	channelID int,
	messageID int,
) error {
	_, err := r.db.Exec(`
		INSERT INTO channel_read_state (
			user_id,
			channel_id,
			last_read_message_id
		)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, channel_id)
		DO UPDATE SET
			last_read_message_id = MAX(
				last_read_message_id,
				excluded.last_read_message_id
			),
			updated_at = CURRENT_TIMESTAMP
	`, userID, channelID, messageID)

	return err
}

func (r *Repository) MessageExists(
	messageID int,
	channelID int,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM messages
			WHERE id = ?
			  AND channel_id = ?
		)
	`, messageID, channelID).Scan(&exists)

	return exists, err
}
