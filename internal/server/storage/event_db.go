package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"time"
)

var _ EventStorager = &EventDBStorage{}

var eventsTableName = "events"

type EventDBStorage struct {
	db *sql.DB
}

func NewEventStorager(db *sql.DB) *EventDBStorage {
	return &EventDBStorage{db: db}
}

var eventAllFieldsSQL = `"uuid", "user_uuid", "record_uuid", "date", "action", "data"`

func (s *EventDBStorage) Create(ctx context.Context, event *models.Event) error {
	insertSQL := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES ($1, $2, $3, $4, $5, $6)`, eventsTableName, eventAllFieldsSQL)
	_, err := s.db.ExecContext(ctx, insertSQL,
		event.UUID,
		event.UserUUID,
		event.RecordUUID,
		event.Date,
		event.Action,
		event.Data,
	)
	return err
}

func (s *EventDBStorage) FindByUserUUID(ctx context.Context, userUuid string, fromTime *time.Time) ([]models.Event, error) {
	var err error
	var rows *sql.Rows
	if fromTime != nil {
		selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "user_uuid" = $1 AND "date" >= $2 ORDER BY "date" ASC`, eventAllFieldsSQL, eventsTableName)
		rows, err = s.db.QueryContext(ctx, selectSQL, userUuid, fromTime)
	} else {
		selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "user_uuid" = $1 ORDER BY "date" ASC`, eventAllFieldsSQL, eventsTableName)
		rows, err = s.db.QueryContext(ctx, selectSQL, userUuid)
	}
	if err != nil {
		return nil, err
	}
	return s.scanAll(rows)
}

func (s *EventDBStorage) FindByRecordUUID(ctx context.Context, recordUuid string, fromTime *time.Time) ([]models.Event, error) {
	var err error
	var rows *sql.Rows
	if fromTime != nil {
		selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "record_uuid" = $1 AND "date" >= $2 ORDER BY "date" ASC`, eventAllFieldsSQL, eventsTableName)
		rows, err = s.db.QueryContext(ctx, selectSQL, recordUuid, fromTime)
	} else {
		selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "record_uuid" = $1 ORDER BY "date" ASC`, eventAllFieldsSQL, eventsTableName)
		rows, err = s.db.QueryContext(ctx, selectSQL, recordUuid)
	}
	if err != nil {
		return nil, err
	}
	return s.scanAll(rows)
}

func (s *EventDBStorage) scanAll(rows *sql.Rows) ([]models.Event, error) {
	events := make([]models.Event, 0)
	for rows.Next() {
		event := models.Event{}
		err := rows.Scan(
			&event.UUID,
			&event.UserUUID,
			&event.RecordUUID,
			&event.Date,
			&event.Action,
			&event.Data,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
