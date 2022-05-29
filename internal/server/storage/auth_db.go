package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"math/rand"
	"time"
)

var _ AuthStorager = &AuthDBStorage{}

const dict = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"

var authsTableName = "auths"

type AuthDBStorage struct {
	db *sql.DB
}

func NewAuthDBStorage(db *sql.DB) *AuthDBStorage {
	return &AuthDBStorage{db: db}
}

func (s *AuthDBStorage) GenerateForUser(ctx context.Context, user *models.User) (*models.Auth, error) {
	now := time.Now()
	auth := &models.Auth{}
	auth.UUID = uuid.NewString()
	auth.UserUUID = user.UUID
	auth.Token = generateToken()
	auth.CreatedAt = &now

	insertSQL := fmt.Sprintf(`INSERT INTO "%s" ("uuid", "user_uuid", "token", "created_at") VALUES ($1, $2, $3, $4)`, authsTableName)
	_, err := s.db.ExecContext(ctx, insertSQL, auth.UUID, auth.UserUUID, auth.Token, auth.CreatedAt)
	return auth, err
}

func generateToken() string {
	token := make([]byte, 64)
	for i := 0; i < len(token); i++ {
		n := rand.Intn(len(dict))
		token[i] = dict[n]
	}
	return string(token)
}

func (s *AuthDBStorage) FindByToken(ctx context.Context, token string) (*models.Auth, error) {
	auth := &models.Auth{}

	selectSQL := fmt.Sprintf(`SELECT "uuid", "user_uuid", "token", "created_at" FROM "%s" u WHERE "token" = $1 LIMIT 1`, authsTableName)
	row := s.db.QueryRowContext(ctx, selectSQL, token)
	err := row.Scan(&auth.UUID, &auth.UserUUID, &auth.Token, &auth.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return auth, nil
}

func (s *AuthDBStorage) FindByUserUUID(ctx context.Context, userUuid string) ([]models.Auth, error) {
	auths := make([]models.Auth, 0)
	selectSQL := fmt.Sprintf(`SELECT "uuid", "user_uuid", "token", "created_at" FROM "%s" WHERE "user_uuid" = $1 LIMIT 1`, authsTableName)
	rows, err := s.db.QueryContext(ctx, selectSQL, userUuid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		auth := models.Auth{}
		err = rows.Scan(&auth.UUID, &auth.UserUUID, &auth.Token, &auth.CreatedAt)
		if err != nil {
			return nil, err
		}
		auths = append(auths, auth)
	}
	return auths, nil
}
