package channels

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	serverID int,
	name string,
	channelType string,
) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO channels (server_id, name, type)
		 VALUES (?, ?, ?)`,
		serverID,
		name,
		channelType,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) GetByID(id int) (Channel, error) {
	var channel Channel

	err := r.db.QueryRow(
		`SELECT
			id,
			server_id,
			name,
			type
		FROM channels
		WHERE id = ?`,
		id,
	).Scan(
		&channel.ID,
		&channel.ServerID,
		&channel.Name,
		&channel.Type,
	)

	return channel, err
}

func (r *Repository) GetByServerID(serverID int) ([]Channel, error) {
	rows, err := r.db.Query(
		`SELECT
			id,
			server_id,
			name,
			type
		FROM channels
		WHERE server_id = ?
		ORDER BY id`,
		serverID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	channels := make([]Channel, 0)

	for rows.Next() {
		var channel Channel

		if err := rows.Scan(
			&channel.ID,
			&channel.ServerID,
			&channel.Name,
			&channel.Type,
		); err != nil {
			return nil, err
		}

		channels = append(channels, channel)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return channels, nil
}
