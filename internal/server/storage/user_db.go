package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/putalexey/goph-keeper/internal/common/models"
)

var _ UserStorager = &UserDBStorage{}

var usersTableName = "users"

type UserDBStorage struct {
	db *sql.DB
}

func NewUserDBStorage(db *sql.DB) *UserDBStorage {
	return &UserDBStorage{db: db}
}

func (s *UserDBStorage) Create(ctx context.Context, user *models.User) error {
	user.UUID = uuid.NewString()
	insertSQL := fmt.Sprintf(`INSERT INTO "%s" ("uuid", "login", "password") VALUES ($1, $2, $3)`, usersTableName)
	_, err := s.db.ExecContext(ctx, insertSQL, user.UUID, user.Login, user.Password)
	return err
}

func (s *UserDBStorage) FindByUUID(ctx context.Context, uuid string) (*models.User, error) {
	user := &models.User{}

	selectSQL := fmt.Sprintf(`SELECT "uuid", "login", "password" FROM "%s" WHERE "uuid" = $1 LIMIT 1`, usersTableName)
	row := s.db.QueryRowContext(ctx, selectSQL, uuid)
	err := row.Scan(&user.UUID, &user.Login, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserDBStorage) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}

	selectSQL := fmt.Sprintf(`SELECT "uuid", "login", "password" FROM "%s" WHERE "login" = $1 LIMIT 1`, usersTableName)
	row := s.db.QueryRowContext(ctx, selectSQL, login)
	err := row.Scan(&user.UUID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserDBStorage) Update(ctx context.Context, user *models.User) error {
	updateSQL := fmt.Sprintf(`UPDATE "%s" SET "login" = $1, "password" = $2 WHERE "uuid" = $3`, usersTableName)
	res, err := s.db.ExecContext(ctx, updateSQL, user.Login, user.Password, user.UUID)

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf(`users with UUID "%s" not found`, user.UUID)
	}

	return err
}

func (s *UserDBStorage) Delete(ctx context.Context, user *models.User) error {
	deleteSQL := fmt.Sprintf(`DELETE FROM "%s" WHERE "uuid" = $1`, usersTableName)
	_, err := s.db.ExecContext(ctx, deleteSQL, user.UUID)
	return err
}
