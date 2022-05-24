// Package storage defines storagers interfaces and contains sql implementation of it
package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/pkg/errors"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"time"
)

func NewDBConnection(databaseDSN string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxIdleTime(30 * time.Second)
	db.SetConnMaxLifetime(2 * time.Minute)

	return db, db.Ping()
}

//ErrNotFound error returned on "not found" error from underlying storage driver
var ErrNotFound = errors.New("not found")

type StoragesContainer struct {
	UserStorage   UserStorager
	AuthStorage   AuthStorager
	RecordStorage RecordStorager
	EventStorage  EventStorager
}

type UserStorager interface {
	Create(ctx context.Context, user *models.User) error
	FindByUUID(ctx context.Context, uuid string) (*models.User, error)
	FindByLogin(ctx context.Context, login string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, user *models.User) error
}

type AuthStorager interface {
	GenerateForUser(ctx context.Context, user *models.User) (*models.Auth, error)
	FindByToken(ctx context.Context, token string) (*models.Auth, error)
	FindByUserUUID(ctx context.Context, userUuid string) ([]models.Auth, error)
}

type RecordStorager interface {
	Create(ctx context.Context, record *models.Record) error
	FindByUUID(ctx context.Context, uuid string) (*models.Record, error)
	FindByUserUUID(ctx context.Context, userUuid string) ([]models.Record, error)
	Update(ctx context.Context, record *models.Record) error
	Delete(ctx context.Context, record *models.Record) error
}

type EventStorager interface {
	Create(ctx context.Context, event *models.Event) error
	FindByUserUUID(ctx context.Context, userUuid string, fromTime *time.Time) ([]models.Event, error)
	FindByRecordUUID(ctx context.Context, recordUuid string, fromTime *time.Time) ([]models.Event, error)
}
