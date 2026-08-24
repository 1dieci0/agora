package messages

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetByID(id int) (Message, error) {
	var message Message

	err := r.db.QueryRow(
		`SELECT
			m.id,
			m.channel_id,
			m.user_id,
			u.username,
			m.content,
			m.created_at
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = ?`,
		id,
	).Scan(
		&message.ID,
		&message.ChannelID,
		&message.UserID,
		&message.Username,
		&message.Content,
		&message.CreatedAt,
	)

	return message, err
}

func (r *Repository) Create(
	channelID int,
	userID int,
	content string,
) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO messages (
			channel_id,
			user_id,
			content
		)
		VALUES (?, ?, ?)`,
		channelID,
		userID,
		content,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(id int, content string) error {
	_, err := r.db.Exec(
		`UPDATE messages
		 SET content = ?
		 WHERE id = ?`,
		content,
		id,
	)

	return err
}

func (r *Repository) Delete(id int) error {
	_, err := r.db.Exec(
		`DELETE FROM messages
		 WHERE id = ?`,
		id,
	)

	return err
}

func (r *Repository) IsMember(userID, channelID int) (bool, error) {
	var member bool

	err := r.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM server_members sm
			JOIN channels c ON c.server_id = sm.server_id
			WHERE sm.user_id = ?
			AND c.id = ?
		)`,
		userID,
		channelID,
	).Scan(&member)

	return member, err
}

func (r *Repository) GetByChannelID(channelID int) ([]Message, error) {
	rows, err := r.db.Query(
		`SELECT
			m.id,
			m.channel_id,
			m.user_id,
			u.username,
			m.content,
			m.created_at
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.channel_id = ?
		ORDER BY m.id DESC
		LIMIT 100`,
		channelID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0)

	for rows.Next() {
		var message Message

		if err := rows.Scan(
			&message.ID,
			&message.ChannelID,
			&message.UserID,
			&message.Username,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
