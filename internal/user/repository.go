package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByEmail(email string) (*User, error)
	GetByUID(uid string) (*User, error)
	CreateUser(user *User) error
	GetAllUsers() ([]*User, error)
	UpdateRole(uid string, role string) error
	UpdateProfile(uid string, email string) (*User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) GetByEmail(email string) (*User, error) {
	ctx := context.Background()
	row := r.pool.QueryRow(ctx,
		`SELECT firebase_uid, email, role, created_at, updated_at
		   FROM users
		  WHERE email = $1`, email)

	var user User
	err := row.Scan(&user.FirebaseUID, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("ไม่พบผู้ใช้")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUID(uid string) (*User, error) {
	ctx := context.Background()
	row := r.pool.QueryRow(ctx,
		`SELECT firebase_uid, email, role, created_at, updated_at
		   FROM users
		  WHERE firebase_uid = $1`, uid)

	var user User
	err := row.Scan(&user.FirebaseUID, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("ไม่พบผู้ใช้")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(user *User) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (firebase_uid, email, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.FirebaseUID, user.Email, user.Role, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *userRepository) GetAllUsers() ([]*User, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT firebase_uid, email, role, created_at, updated_at
		   FROM users
		  ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.FirebaseUID, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}

func (r *userRepository) UpdateRole(uid string, role string) error {
	ctx := context.Background()
	cmd, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $1, updated_at = $2 WHERE firebase_uid = $3`,
		role, time.Now(), uid)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("ไม่พบผู้ใช้")
	}
	return nil
}

func (r *userRepository) UpdateProfile(uid string, email string) (*User, error) {
	ctx := context.Background()
	cmd, err := r.pool.Exec(ctx,
		`UPDATE users SET email = $1, updated_at = $2 WHERE firebase_uid = $3`,
		email, time.Now(), uid)
	if err != nil {
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, errors.New("ไม่พบผู้ใช้")
	}
	return r.GetByUID(uid)
}
