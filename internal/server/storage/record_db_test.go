package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRecordDBStorage_Create(t *testing.T) {
	type fields struct {
		db         *sql.DB
		encryptKey string
	}
	type args struct {
		ctx    context.Context
		record *models.Record
	}
	userStore := NewUserDBStorage(sharedDB)
	user := &models.User{
		Login:    "test",
		Password: "test",
	}
	err := userStore.Create(context.Background(), user)
	require.NoError(t, err)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{name: "saves record unencrypted in db when key is not provided", fields: fields{
			db:         sharedDB,
			encryptKey: "",
		}, args: args{
			ctx: context.Background(),
			record: &models.Record{
				UserUUID: user.UUID,
				Name:     "test",
				Type:     "text",
				Data:     []byte("some text data"),
				Comment:  "comment",
			},
		}, wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NoError(t, err)
			_args := i[1].(args)
			record := _args.record
			row := sharedDB.QueryRow(`select "data" from records where "uuid"=$1`, record.UUID)
			var data []byte
			err = row.Scan(&data)
			assert.NoError(t, err)
			if err != nil {
				return false
			}
			return assert.Equal(t, record.Data, data)
		}},
		{name: "saves record encrypted in db when key is provided", fields: fields{
			db:         sharedDB,
			encryptKey: "some-key",
		}, args: args{
			ctx: context.Background(),
			record: &models.Record{
				UserUUID: user.UUID,
				Name:     "test",
				Type:     "text",
				Data:     []byte("some text data"),
				Comment:  "comment",
			},
		}, wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NoError(t, err)
			_args := i[1].(args)
			record := _args.record
			row := sharedDB.QueryRow(`select "data" from records where "uuid"=$1`, record.UUID)
			var data []byte
			err = row.Scan(&data)
			assert.NoError(t, err)
			if err != nil {
				return false
			}
			return assert.NotEqual(t, record.Data, data)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RecordDBStorage{
				db:         tt.fields.db,
				encryptKey: tt.fields.encryptKey,
			}
			tt.wantErr(t, s.Create(tt.args.ctx, tt.args.record), fmt.Sprintf("Create(%v, %v)", tt.args.ctx, tt.args.record), tt.args)
		})
	}
}
