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
) (int, error) {
	if userID == otherUserID {
		return 0, errors.New("cannot create DM with yourself")
	}

	var conversationID int

	err := r.db.QueryRow(`
		SELECT c.id
		FROM direct_conversations c
		JOIN direct_conversation_members m1
			ON m1.conversation_id = c.id
		JOIN direct_conversation_members m2
			ON m2.conversation_id = c.id
		WHERE m1.user_id = ?
		  AND m2.user_id = ?
		  AND (
			SELECT COUNT(*)
			FROM direct_conversation_members
			WHERE conversation_id = c.id
		  ) = 2
		LIMIT 1
	`, userID, otherUserID).Scan(&conversationID)

	if err == nil {
		return conversationID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO direct_conversations DEFAULT VALUES
	`)
	if err != nil {
		return 0, err
	}

	conversationID64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	conversationID = int(conversationID64)

	_, err = tx.Exec(`
		INSERT INTO direct_conversation_members (
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
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return conversationID, nil
}

func (r *Repository) GetConversations(userID int) ([]Conversation, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id,
			u.id,
			u.username,
			u.avatar_url,
			c.created_at,

			COUNT(dm.id) AS unread_count,

			(
				SELECT COUNT(*)
				FROM notifications n
				WHERE n.user_id = ?
				  AND n.conversation_id = c.id
				  AND n.type = 'mention'
				  AND n.read = 0
			) AS mention_count

		FROM direct_conversations c

		JOIN direct_conversation_members m
			ON m.conversation_id = c.id

		JOIN direct_conversation_members other
			ON other.conversation_id = c.id
			AND other.user_id != ?

		JOIN users u
			ON u.id = other.user_id

		LEFT JOIN direct_messages dm
			ON dm.conversation_id = c.id
			AND dm.user_id != ?
			AND dm.id > COALESCE(
				(
					SELECT last_read_message_id
					FROM direct_conversation_read_state
					WHERE user_id = ?
					  AND conversation_id = c.id
				),
				0
			)

		WHERE m.user_id = ?

		GROUP BY
			c.id,
			u.id,
			u.username,
			u.avatar_url,
			c.created_at

		ORDER BY c.id DESC
	`,
		userID,
		userID,
		userID,
		userID,
		userID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]Conversation, 0)

	for rows.Next() {
		var conversation Conversation

		if err := rows.Scan(
			&conversation.ID,
			&conversation.UserID,
			&conversation.Username,
			&conversation.AvatarURL,
			&conversation.CreatedAt,
			&conversation.UnreadCount,
			&conversation.MentionCount,
		); err != nil {
			return nil, err
		}

		conversations = append(conversations, conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}
func (r *Repository) GetOtherUser(
	conversationID int,
	userID int,
) (int, error) {
	var otherUserID int

	err := r.db.QueryRow(`
		SELECT user_id
		FROM direct_conversation_members
		WHERE conversation_id = ?
		  AND user_id != ?
		LIMIT 1
	`, conversationID, userID).Scan(&otherUserID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrConversationNotFound
	}

	if err != nil {
		return 0, err
	}

	return otherUserID, nil
}

func (r *Repository) IsMember(
	conversationID int,
	userID int,
) (bool, error) {
	var exists int

	err := r.db.QueryRow(`
		SELECT 1
		FROM direct_conversation_members
		WHERE conversation_id = ?
		  AND user_id = ?
		LIMIT 1
	`, conversationID, userID).Scan(&exists)

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
	userID int,
) ([]DirectMessage, error) {
	isMember, err := r.IsMember(
		conversationID,
		userID,
	)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, ErrConversationNotFound
	}

	rows, err := r.db.Query(`
		SELECT
			m.id,
			m.conversation_id,
			m.user_id,
			u.username,
			u.avatar_url,
			m.content,
			m.created_at
		FROM direct_messages m
		JOIN users u
			ON u.id = m.user_id
		WHERE m.conversation_id = ?
		ORDER BY m.id ASC
	`, conversationID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	messages := make([]DirectMessage, 0)

	for rows.Next() {
		var message DirectMessage

		if err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.UserID,
			&message.Username,
			&message.AvatarURL,
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

func (r *Repository) CreateMessage(
	conversationID int,
	userID int,
	content string,
) (DirectMessage, error) {
	isMember, err := r.IsMember(
		conversationID,
		userID,
	)
	if err != nil {
		return DirectMessage{}, err
	}

	if !isMember {
		return DirectMessage{}, ErrConversationNotFound
	}

	result, err := r.db.Exec(`
		INSERT INTO direct_messages (
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
		return DirectMessage{}, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return DirectMessage{}, err
	}

	var message DirectMessage

	err = r.db.QueryRow(`
		SELECT
			m.id,
			m.conversation_id,
			m.user_id,
			u.username,
			u.avatar_url,
			m.content,
			m.created_at
		FROM direct_messages m
		JOIN users u
			ON u.id = m.user_id
		WHERE m.id = ?
	`, messageID).Scan(
		&message.ID,
		&message.ConversationID,
		&message.UserID,
		&message.Username,
		&message.AvatarURL,
		&message.Content,
		&message.CreatedAt,
	)
	if err != nil {
		return DirectMessage{}, err
	}

	return message, nil
}

func (r *Repository) MarkConversationRead(
	conversationID int,
	userID int,
	messageID int,
) error {
	isMember, err := r.IsMember(
		conversationID,
		userID,
	)
	if err != nil {
		return err
	}

	if !isMember {
		return ErrConversationNotFound
	}

	_, err = r.db.Exec(`
		INSERT INTO direct_conversation_read_state (
			user_id,
			conversation_id,
			last_read_message_id
		)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, conversation_id)
		DO UPDATE SET
			last_read_message_id = excluded.last_read_message_id,
			updated_at = CURRENT_TIMESTAMP
	`,
		userID,
		conversationID,
		messageID,
	)

	return err
}

func (r *Repository) GetUser(
	userID int,
) (string, error) {
	var username string

	err := r.db.QueryRow(`
		SELECT username
		FROM users
		WHERE id = ?
	`, userID).Scan(&username)

	return username, err
}
