package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func insertUser(ctx context.Context, db *sql.DB, email, passwordHash, fullName string) (string, error) {
	id := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, full_name, status) VALUES ($1, $2, $3, $4, 'active')`,
		id, email, passwordHash, fullName,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureAdminRoleAndAssign(ctx context.Context, db *sql.DB, userID string) (roles []string, permissions []string, err error) {
	var roleID string
	err = db.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'ADMIN' LIMIT 1`).Scan(&roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			roleID = uuid.New().String()
			_, err = db.ExecContext(ctx,
				`INSERT INTO roles (id, name, description, is_system) VALUES ($1, 'ADMIN', 'Administrator', true)`,
				roleID,
			)
			if err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT (user_id, role_id) DO NOTHING`,
		userID, roleID,
	)
	if err != nil {
		return nil, nil, err
	}
	return []string{"ADMIN"}, nil, nil
}

func findUserAndVerify(ctx context.Context, db *sql.DB, email, password string) (userID string, roles []string, permissions []string, err error) {
	var hash string
	err = db.QueryRowContext(ctx,
		`SELECT id, password_hash FROM users WHERE email = $1 AND status = 'active'`,
		email,
	).Scan(&userID, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, nil, errors.New("invalid email or password")
		}
		return "", nil, nil, err
	}
	if err := comparePassword(hash, password); err != nil {
		return "", nil, nil, errors.New("invalid email or password")
	}
	rows, err := db.QueryContext(ctx,
		`SELECT r.name FROM roles r JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id = $1`,
		userID,
	)
	if err != nil {
		return userID, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return "", nil, nil, err
		}
		roles = append(roles, name)
	}
	return userID, roles, nil, nil
}
