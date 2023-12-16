package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
)

type UsersRepo struct {
	s *Storage
}

func NewUsersRepo(s *Storage) *UsersRepo {
	return &UsersRepo{
		s: s,
	}
}

const queryGetUser = `SELECT id, login, password FROM users WHERE id=$1;`

// Get finds user by id. When requested user doesn't exist
// storage.ErrNotFound error is returned.
func (r *UsersRepo) Get(id int64) (user model.User, err error) {
	stmt, err := r.s.db.Prepare(queryGetUser)
	if err != nil {
		return user, err
	}
	defer stmt.Close()

	if err = stmt.QueryRow(id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNotFound
		}
		return user, err
	}

	return user, nil
}

const queryFindUserByLogin = `SELECT id, login, password FROM users WHERE login=$1;`

// FindByLogin finds user by login. When requested user doesn't exist
// storage.ErrNotFound error is returned.
func (r *UsersRepo) FindByLogin(login string) (user model.User, err error) {
	login = strings.ToLower(login)

	stmt, err := r.s.db.Prepare(queryFindUserByLogin)
	if err != nil {
		return user, err
	}
	defer stmt.Close()

	if err = stmt.QueryRow(login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNotFound
		}
		return user, err
	}

	return user, nil
}

const queryCreateUser = `
	INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id;
`

func (r *UsersRepo) Create(user model.User) (id int64, err error) {
	user.Login = strings.ToLower(user.Login)

	stmt, err := r.s.db.Prepare(queryCreateUser)
	if err != nil {
		return id, err
	}
	defer stmt.Close()

	// TODO: handle login uniqueness error
	err = stmt.QueryRow(user.Login, user.PasswordHash).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, err
}

const queryDeleteUser = `DELETE FROM users WHERE id=$1;`

func (r *UsersRepo) Delete(id int64) error {
	stmt, err := r.s.db.Prepare(queryDeleteUser)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)

	return err
}
