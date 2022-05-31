package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"time"
)

var _ RecordStorager = &RecordDBStorage{}

var recordsTableName = "records"

type RecordDBStorage struct {
	db         *sql.DB
	encryptKey string
}

func NewRecordDBStorage(db *sql.DB, encryptKey string) *RecordDBStorage {
	return &RecordDBStorage{db: db, encryptKey: encryptKey}
}

var recordAllFieldsSQL = `"uuid", "user_uuid", "name", "type", "data", "comment", "created_at", "updated_at", "deleted_at", "data_encrypted"`

func (s *RecordDBStorage) Create(ctx context.Context, record *models.Record) error {
	if record.UUID == "" {
		record.UUID = uuid.NewString()
	}
	if record.CreatedAt == nil {
		createdAt := time.Now()
		record.CreatedAt = &createdAt
	}
	if record.UpdatedAt == nil {
		updatedAt := time.Now()
		record.UpdatedAt = &updatedAt
	}

	data, isEncrypted, err := s.encryptData(record.Data)
	if err != nil {
		return fmt.Errorf("update record: encrypt data error: %w", err)
	}
	insertSQL := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, recordsTableName, recordAllFieldsSQL)
	_, err = s.db.ExecContext(ctx, insertSQL,
		record.UUID,
		record.UserUUID,
		record.Name,
		record.Type,
		data,
		record.Comment,
		record.CreatedAt,
		record.UpdatedAt,
		record.DeletedAt,
		isEncrypted,
	)
	return err
}

func (s *RecordDBStorage) GetByUUID(ctx context.Context, uuid string) (*models.Record, error) {
	var dataWasEncrypted bool
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
		&dataWasEncrypted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	decrypted, err := s.decryptData(record.Data, dataWasEncrypted)
	if err != nil {
		return nil, fmt.Errorf("get record by UUID: decrypt data error: %w", err)
	}
	record.Data = decrypted
	return record, nil
}

func (s *RecordDBStorage) FindByUserUUID(ctx context.Context, userUuid string) ([]models.Record, error) {
	var dataWasEncrypted bool
	records := make([]models.Record, 0)
	selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "user_uuid" = $1 and deleted_at is null`, recordAllFieldsSQL, recordsTableName)
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
			&dataWasEncrypted,
		)
		if err != nil {
			return nil, err
		}
		decrypted, err := s.decryptData(record.Data, dataWasEncrypted)
		if err != nil {
			return nil, fmt.Errorf("get record by UUID: decrypt data error: %w", err)
		}
		record.Data = decrypted
		records = append(records, record)
	}
	return records, nil
}

func (s *RecordDBStorage) GetByUserUUIDAndName(ctx context.Context, userUuid string, name string) (*models.Record, error) {
	var dataWasEncrypted bool
	record := &models.Record{}

	selectSQL := fmt.Sprintf(`SELECT %s FROM "%s" WHERE "user_uuid" = $1 and "name" = $2 LIMIT 1`, recordAllFieldsSQL, recordsTableName)
	row := s.db.QueryRowContext(ctx, selectSQL, userUuid, name)
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
		&dataWasEncrypted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	decrypted, err := s.decryptData(record.Data, dataWasEncrypted)
	if err != nil {
		return nil, fmt.Errorf("get record by UUID: decrypt data error: %w", err)
	}
	record.Data = decrypted
	return record, nil
}

func (s *RecordDBStorage) Update(ctx context.Context, record *models.Record) error {
	data, isEncrypted, err := s.encryptData(record.Data)
	if err != nil {
		return fmt.Errorf("update record: encrypt data error: %w", err)
	}
	updatedAt := time.Now()
	record.UpdatedAt = &updatedAt
	updateSQL := fmt.Sprintf(`UPDATE "%s" SET
"name" = $1,
"type" = $2,
"data" = $3,
"comment" = $4,
"created_at" = $5,
"updated_at" = $6,
"deleted_at" = $7,
"data_encrypted" = $8
WHERE "uuid" = $9`, recordsTableName)
	res, err := s.db.ExecContext(
		ctx,
		updateSQL,
		record.Name,
		record.Type,
		data,
		record.Comment,
		record.CreatedAt,
		record.UpdatedAt,
		record.DeletedAt,
		isEncrypted,
		record.UUID,
	)

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return err
}

func (s *RecordDBStorage) encryptData(srcData []byte) ([]byte, bool, error) {
	if s.encryptKey == "" {
		return srcData, false, nil
	}
	key := sha256.Sum256([]byte(s.encryptKey))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, false, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, false, err
	}

	// создаём вектор инициализации
	nonce := key[len(key)-aesgcm.NonceSize():]

	// расшифровываем
	encrypted := aesgcm.Seal(nil, nonce, srcData, nil)
	return encrypted, true, nil
}
func (s *RecordDBStorage) decryptData(srcData []byte, dataEncrypted bool) ([]byte, error) {
	if !dataEncrypted {
		return srcData, nil
	}
	if s.encryptKey == "" {
		return nil, errors.New("encryption key is empty")
	}

	key := sha256.Sum256([]byte(s.encryptKey))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	// создаём вектор инициализации
	nonce := key[len(key)-aesgcm.NonceSize():]

	// расшифровываем
	decrypted, err := aesgcm.Open(nil, nonce, srcData, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func (s *RecordDBStorage) Delete(ctx context.Context, record *models.Record) error {
	//deleteSQL := fmt.Sprintf(`DELETE FROM "%s" WHERE "uuid" = $1`, recordsTableName)
	//_, err := s.db.ExecContext(ctx, deleteSQL, record.UUID)
	//return err

	deletedAt := time.Now()
	record.DeletedAt = &deletedAt

	updateSQL := fmt.Sprintf(`UPDATE "%s" SET "deleted_at" = $1 WHERE "uuid" = $2`, recordsTableName)
	res, err := s.db.ExecContext(
		ctx,
		updateSQL,
		record.DeletedAt,
		record.UUID,
	)

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return err
}
