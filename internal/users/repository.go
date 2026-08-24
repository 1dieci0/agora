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
	var user User

	err := r.db.QueryRow(
		`SELECT
			id,
			username
		FROM users
		WHERE id = ?`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
	)

	return user, err
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
