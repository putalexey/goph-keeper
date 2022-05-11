// Package storage defines storagers interfaces and contains sql implementation of it
package storage

import (
	"context"
	"database/sql"
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

type UserStorager interface {
	Create(ctx context.Context, user *models.User) error
	FindByUUID(ctx context.Context, uuid string) (*models.User, error)
	FindByLogin(ctx context.Context, login string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, user *models.User) error
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
