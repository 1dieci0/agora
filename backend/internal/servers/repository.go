package servers

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(name string, ownerID int) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(`INSERT INTO servers (name, owner_id) VALUES (?, ?)`, name, ownerID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	serverID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec(`INSERT INTO server_members (server_id, user_id, role) VALUES (?, ?, 'owner')`, serverID, ownerID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return serverID, nil
}

func (r *Repository) GetByID(id int) (Server, error) {
	var server Server

	err := r.db.QueryRow(
		`SELECT
			id,
			name,
			owner_id
		FROM servers
		WHERE id = ?`,
		id,
	).Scan(
		&server.ID,
		&server.Name,
		&server.OwnerID,
	)

	return server, err
}

func (r *Repository) IsMember(userID, serverID int) (bool, error) {
	var member bool

	err := r.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM server_members
			WHERE server_id = ?
			AND user_id = ?
		)`,
		serverID,
		userID,
	).Scan(&member)

	return member, err
}

func (r *Repository) GetByUserID(userID int) ([]Server, error) {
	rows, err := r.db.Query(`SELECT s.id, s.name, s.owner_id FROM servers s JOIN server_members sm ON sm.server_id = s.id WHERE sm.user_id = ? ORDER BY s.id`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	servers := make([]Server, 0)

	for rows.Next() {

		var server Server
		if err := rows.Scan(&server.ID, &server.Name, &server.OwnerID); err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

func (r *Repository) AddMember(userID, serverID int) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO server_members (server_id, user_id, role)
		 VALUES (?, ?, 'member')`,
		serverID,
		userID,
	)

	return err
}

func (r *Repository) RemoveMember(userID, serverID int) (bool, error) {
	result, err := r.db.Exec(
		`DELETE FROM server_members
		 WHERE server_id = ?
		 AND user_id = ?`,
		serverID,
		userID,
	)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, sql.ErrNoRows
	}

	return rows > 0, nil
}

func (r *Repository) GetMembers(serverID int) ([]Member, error) {
	rows, err := r.db.Query(
		`SELECT 
			u.id,
			u.username,
			sm.role 
		FROM users u 
		JOIN server_members sm 
			ON sm.user_id = u.id 
		WHERE sm.server_id = ? ORDER BY u.username`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]Member, 0)

	for rows.Next() {

		var member Member
		if err := rows.Scan(
			&member.ID,
			&member.Username,
			&member.Role,
		); err != nil {
			return nil, err
		}

		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *Repository) CreateInvite(code string, serverID int, createdBy int) error {
	_, err := r.db.Exec(`INSERT INTO invites ( code, server_id, created_by ) VALUES (?, ?, ?)`, code, serverID, createdBy)

	return err
}

func (r *Repository) GetServerIDByInviteCode(code string) (int, error) {
	var serverID int
	err := r.db.QueryRow(`SELECT server_id FROM invites WHERE code = ?`, code).Scan(&serverID)

	return serverID, err
}

func (r *Repository) GetMemberRole(userID, serverID int) (Role, error) {
	var role Role

	err := r.db.QueryRow(
		`SELECT role
		 FROM server_members
		 WHERE server_id = ?
		 AND user_id = ?`,
		serverID,
		userID,
	).Scan(&role)

	return role, err
}

func (r *Repository) IsOwner(userID, serverID int) (bool, error) {
	role, err := r.GetMemberRole(userID, serverID)

	if err != nil {
		return false, err
	}

	return role == RoleOwner, err
}

func (r *Repository) HasAdminPermission(userID, serverID int) (bool, error) {
	role, err := r.GetMemberRole(userID, serverID)

	if err != nil {
		return false, err
	}

	return role == RoleOwner || role == RoleAdmin, err
}

func (r *Repository) UpdateMemberRole(
	serverID int,
	userID int,
	role Role,
) error {
	_, err := r.db.Exec(
		`UPDATE server_members
		 SET role = ?
		 WHERE server_id = ?
		 AND user_id = ?`,
		role,
		serverID,
		userID,
	)

	return err
}

func (r *Repository) Delete(serverID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete messages belonging to channels in this server.
	_, err = tx.Exec(
		`DELETE FROM messages
		 WHERE channel_id IN (
			SELECT id
			FROM channels
			WHERE server_id = ?
		 )`,
		serverID,
	)
	if err != nil {
		return err
	}

	// Delete channels.
	_, err = tx.Exec(
		`DELETE FROM channels
		 WHERE server_id = ?`,
		serverID,
	)
	if err != nil {
		return err
	}

	// Delete invites.
	_, err = tx.Exec(
		`DELETE FROM invites
		 WHERE server_id = ?`,
		serverID,
	)
	if err != nil {
		return err
	}

	// Delete server members.
	_, err = tx.Exec(
		`DELETE FROM server_members
		 WHERE server_id = ?`,
		serverID,
	)
	if err != nil {
		return err
	}

	// Finally delete the server.
	result, err := tx.Exec(
		`DELETE FROM servers
		 WHERE id = ?`,
		serverID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
