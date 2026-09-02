package dms

import (
	"database/sql"
	"errors"
)

var ErrConversationNotFound = errors.New("conversation not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetOrCreateConversation(
	userID int,
	otherUserID int,
) (*Conversation, error) {
	if userID == otherUserID {
		return nil, errors.New("cannot create a conversation with yourself")
	}

	/*
	 * Look for an existing conversation containing
	 * exactly these two users.
	 */
	var conversationID int
	var createdAt string

	err := r.db.QueryRow(`
		SELECT
			c.id,
			c.created_at
		FROM dm_conversations c
		JOIN dm_participants p1
			ON p1.conversation_id = c.id
		JOIN dm_participants p2
			ON p2.conversation_id = c.id
		WHERE p1.user_id = ?
		  AND p2.user_id = ?
		LIMIT 1
	`,
		userID,
		otherUserID,
	).Scan(
		&conversationID,
		&createdAt,
	)

	if err == nil {
		return r.getConversation(conversationID)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	/*
	 * No conversation exists, so create one.
	 */
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO dm_conversations DEFAULT VALUES
	`)

	if err != nil {
		return nil, err
	}

	conversationID64, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	conversationID = int(conversationID64)

	_, err = tx.Exec(`
		INSERT INTO dm_participants (
			conversation_id,
			user_id
		)
		VALUES (?, ?), (?, ?)
	`,
		conversationID,
		userID,
		conversationID,
		otherUserID,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.getConversation(conversationID)
}

func (r *Repository) getConversation(
	conversationID int,
) (*Conversation, error) {
	var conversation Conversation

	err := r.db.QueryRow(`
		SELECT id, created_at
		FROM dm_conversations
		WHERE id = ?
	`,
		conversationID,
	).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT
			u.id,
			u.username,
			u.avatar_url
		FROM dm_participants p
		JOIN users u
			ON u.id = p.user_id
		WHERE p.conversation_id = ?
		ORDER BY u.id ASC
	`,
		conversationID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var participant Participant

		err := rows.Scan(
			&participant.ID,
			&participant.Username,
			&participant.AvatarURL,
		)

		if err != nil {
			return nil, err
		}

		conversation.Participants = append(
			conversation.Participants,
			participant,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *Repository) GetConversationsForUser(
	userID int,
) ([]Conversation, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id,
			c.created_at
		FROM dm_conversations c
		JOIN dm_participants p
			ON p.conversation_id = c.id
		WHERE p.user_id = ?
		ORDER BY (
			SELECT MAX(m.created_at)
			FROM dm_messages m
			WHERE m.conversation_id = c.id
		) DESC,
		c.created_at DESC
	`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var conversations []Conversation

	for rows.Next() {
		var conversation Conversation

		err := rows.Scan(
			&conversation.ID,
			&conversation.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		fullConversation, err := r.getConversation(
			conversation.ID,
		)

		if err != nil {
			return nil, err
		}

		conversations = append(
			conversations,
			*fullConversation,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *Repository) IsParticipant(
	conversationID int,
	userID int,
) (bool, error) {
	var exists int

	err := r.db.QueryRow(`
		SELECT 1
		FROM dm_participants
		WHERE conversation_id = ?
		  AND user_id = ?
		LIMIT 1
	`,
		conversationID,
		userID,
	).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) GetMessages(
	conversationID int,
) ([]Message, error) {
	rows, err := r.db.Query(`
		SELECT
			m.id,
			m.conversation_id,
			m.user_id,
			u.username,
			m.content,
			m.created_at
		FROM dm_messages m
		JOIN users u
			ON u.id = m.user_id
		WHERE m.conversation_id = ?
		ORDER BY m.created_at ASC, m.id ASC
	`,
		conversationID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var message Message

		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.UserID,
			&message.Username,
			&message.Content,
			&message.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *Repository) CreateMessage(
	conversationID int,
	userID int,
	content string,
) (*Message, error) {
	result, err := r.db.Exec(`
		INSERT INTO dm_messages (
			conversation_id,
			user_id,
			content
		)
		VALUES (?, ?, ?)
	`,
		conversationID,
		userID,
		content,
	)

	if err != nil {
		return nil, err
	}

	messageID64, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	var message Message

	err = r.db.QueryRow(`
		SELECT
			m.id,
			m.conversation_id,
			m.user_id,
			u.username,
			m.content,
			m.created_at
		FROM dm_messages m
		JOIN users u
			ON u.id = m.user_id
		WHERE m.id = ?
	`,
		messageID64,
	).Scan(
		&message.ID,
		&message.ConversationID,
		&message.UserID,
		&message.Username,
		&message.Content,
		&message.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &message, nil
}
