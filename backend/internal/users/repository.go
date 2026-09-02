package users

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(username, passwordHash string) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (username, password_hash)
		 VALUES (?, ?)`,
		username,
		passwordHash,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) GetByUsername(username string) (
	int,
	string,
	string,
	error,
) {
	var (
		userID       int
		dbUsername   string
		passwordHash string
	)

	err := r.db.QueryRow(
		`SELECT
			id,
			username,
			password_hash
		FROM users
		WHERE username = ?`,
		username,
	).Scan(
		&userID,
		&dbUsername,
		&passwordHash,
	)

	return userID, dbUsername, passwordHash, err
}

func (r *Repository) GetByID(id int) (User, error) {
	var (
		user      User
		avatarURL sql.NullString
	)

	err := r.db.QueryRow(
		`SELECT
			id,
			username,
			avatar_url
		FROM users
		WHERE id = ?`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&avatarURL,
	)

	if err != nil {
		return user, err
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return user, nil
}

func (r *Repository) CreateSession(userID int) (string, error) {
	return createSession(r.db, userID)
}

func (r *Repository) DeleteSession(sessionID string) error {
	return deleteSession(r.db, sessionID)
}

func (r *Repository) GetUserIDFromSession(sessionID string) (int, error) {
	var userID int

	err := r.db.QueryRow(
		`SELECT user_id
		 FROM sessions
		 WHERE id = ?
		 AND expires_at > datetime('now')`,
		sessionID,
	).Scan(&userID)

	return userID, err
}

func (r *Repository) UpdateAvatar(userID int, avatarURL string) error {
	_, err := r.db.Exec(
		`UPDATE users
		 SET avatar_url = ?
		 WHERE id = ?`,
		avatarURL,
		userID,
	)

	return err
}

func (r *Repository) GetAvatarURL(userID int) (string, error) {
	var avatarURL sql.NullString

	err := r.db.QueryRow(
		`SELECT avatar_url
		 FROM users
		 WHERE id = ?`,
		userID,
	).Scan(&avatarURL)

	if err != nil {
		return "", err
	}

	if !avatarURL.Valid {
		return "", nil
	}

	return avatarURL.String, nil
}

func (r *Repository) GetServerIDsForUser(userID int) ([]int, error) {
	rows, err := r.db.Query(`
		SELECT server_id
		FROM server_members
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var serverIDs []int

	for rows.Next() {
		var serverID int

		if err := rows.Scan(&serverID); err != nil {
			return nil, err
		}

		serverIDs = append(serverIDs, serverID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return serverIDs, nil
}
