package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/common/models"
)

var _ RecordStorager = &RecordDBStorage{}

var recordsTableName = "records"

type RecordDBStorage struct {
	db *sql.DB
}

var recordAllFieldsSQL = `"uuid", "user_uuid", "name", "type", "data", "comment", "created_at", "updated_at", "deleted_at"`

func (s *RecordDBStorage) Create(ctx context.Context, record *models.Record) error {
	insertSQL := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, recordsTableName, recordAllFieldsSQL)
	_, err := s.db.ExecContext(ctx, insertSQL,
		record.UUID,
		record.UserUUID,
		record.Name,
		record.Type,
		record.Data,
		record.Comment,
		record.CreatedAt,
		record.UpdatedAt,
		record.DeletedAt,
	)
	return err
}

func (s *RecordDBStorage) FindByUUID(ctx context.Context, uuid string) (*models.Record, error) {
	record := &models.Record{}

	selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "uuid" = $1 LIMIT 1`, recordAllFieldsSQL, recordsTableName)
	row := s.db.QueryRowContext(ctx, selectSQL, uuid)
	err := row.Scan(
		&record.UUID,
		&record.UserUUID,
		&record.Name,
		&record.Type,
		&record.Data,
		&record.Comment,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *RecordDBStorage) FindByUserUUID(ctx context.Context, userUuid string) ([]models.Record, error) {
	records := make([]models.Record, 0)
	selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "user_uuid" = $1`, recordAllFieldsSQL, recordsTableName)
	rows, err := s.db.QueryContext(ctx, selectSQL, userUuid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		record := models.Record{}
		err = rows.Scan(
			&record.UUID,
			&record.UserUUID,
			&record.Name,
			&record.Type,
			&record.Data,
			&record.Comment,
			&record.CreatedAt,
			&record.UpdatedAt,
			&record.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *RecordDBStorage) Update(ctx context.Context, record *models.Record) error {
	updateSQL := fmt.Sprintf(`UPDATE "%s" SET
"name" = $1,
"type" = $2,
"data" = $3,
"comment" = $4,
"created_at" = $5,
"updated_at" = $6,
"deleted_at" = $7
WHERE "uuid" = $8`, recordsTableName)
	res, err := s.db.ExecContext(
		ctx,
		updateSQL,
		record.Name,
		record.Type,
		record.Data,
		record.Comment,
		record.CreatedAt,
		record.UpdatedAt,
		record.DeletedAt,
		record.UUID,
	)

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf(`record with UUID "%s" not found`, record.UUID)
	}

	return err
}

func (s *RecordDBStorage) Delete(ctx context.Context, record *models.Record) error {
	deleteSQL := fmt.Sprintf(`DELETE FROM "%s" WHERE "uuid" = $1`, recordsTableName)
	_, err := s.db.ExecContext(ctx, deleteSQL, record.UUID)
	return err
}
