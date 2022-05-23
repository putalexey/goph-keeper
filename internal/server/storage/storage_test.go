package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/pressly/goose/v3"
	"github.com/putalexey/goph-keeper/internal/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
)

var sharedDB *sql.DB
var sharedDSN string

func TestMain(m *testing.M) {
	var code int
	// call flag.Parse() here if TestMain uses flags
	err := withDockerDB(func(databaseDSN string, _db *sql.DB) {
		sharedDB = _db
		sharedDSN = databaseDSN
		code = m.Run()
		sharedDB = nil
		sharedDSN = ""
	})
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

func withDockerDB(f func(databaseDSN string, db *sql.DB)) error {
	var db *sql.DB
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("could not connect to docker: %w", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=secret"})
	if err != nil {
		return fmt.Errorf("could not start resource: %w", err)
	}
	defer func() {
		// You can't defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}()

	log.Printf("Started resource: %s\n", resource.Container.Name)

	databaseDSN := fmt.Sprintf("postgres://postgres:secret@localhost:%s/postgres", resource.GetPort("5432/tcp"))
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		db, err = sql.Open("pgx", databaseDSN)
		if err != nil {
			return err
		}
		return db.Ping()
	})
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	dir, _ := os.Getwd()
	log.Println(dir)

	log.Println("Applying migrations")
	err = goose.Up(db, "../../../migrations")
	if err != nil {
		return err
	}
	log.Println("Migrations finished")

	f(databaseDSN, db)

	return nil
}

func TestNewConnection(t *testing.T) {
	t.Run("created new connection", func(t *testing.T) {
		db, err := NewDBConnection(sharedDSN)
		require.NoError(t, err)

		err = db.Close()
		assert.NoError(t, err)
	})
	t.Run("returns error on wrong dsn", func(t *testing.T) {
		_, err := NewDBConnection("wrong formatted dsn")
		assert.Error(t, err)
	})
}
func TestDBStorage(t *testing.T) {
	sharedUUID := uuid.New().String()
	users := []*models.User{
		{
			UUID:     sharedUUID,
			Login:    "test user",
			Password: "123456",
		},
		{
			UUID:     sharedUUID,
			Login:    "test user 2",
			Password: "123456",
		},
		{
			UUID:     uuid.New().String(),
			Login:    "test user",
			Password: "123456",
		},
	}
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx  context.Context
		user *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "creates user in db", fields: fields{db: sharedDB}, args: args{
			ctx:  context.Background(),
			user: users[0],
		}, wantErr: false},
		{name: "fails to add with same uuid", fields: fields{db: sharedDB}, args: args{
			ctx:  context.Background(),
			user: users[1],
		}, wantErr: true},
		{name: "fails to add with same login", fields: fields{db: sharedDB}, args: args{
			ctx:  context.Background(),
			user: users[2],
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserDBStorage{
				db: tt.fields.db,
			}
			if err := s.Create(tt.args.ctx, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// FindByUUID tests
	t.Run("finds user by uuid", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}
		user, err := s.FindByUUID(context.Background(), users[0].UUID)
		assert.NoError(t, err, "FindByUUID() error = %v, wantErr %v", err, false)
		assert.Equal(t, users[0], user, "must be equal")
	})

	t.Run("returns error, if user not found by uuid", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}
		_, err := s.FindByUUID(context.Background(), uuid.New().String())
		assert.Error(t, err, "FindByUUID() error = %v, wantErr %v", err, true)
	})

	// FindByLogin tests
	t.Run("finds user by login", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}
		user, err := s.FindByLogin(context.Background(), users[0].Login)
		assert.NoError(t, err, "FindByLogin() error = %v, wantErr %v", err, false)
		assert.Equal(t, users[0], user, "must be equal")
	})

	t.Run("returns error, if user not found by login", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}
		_, err := s.FindByLogin(context.Background(), "non existed login")
		assert.Error(t, err, "FindByLogin() error = %v, wantErr %v", err, true)
	})

	// Update tests
	t.Run("updates user", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}
		var newUser models.User
		newUser = *users[0]
		newUser.Login = "Changed user Login"
		require.NotEqual(t, newUser.Login, users[0].Login, "bug in test, need copy of struct")

		err := s.Update(context.Background(), &newUser)
		assert.NoError(t, err, "Update() error = %v, wantErr %v", err, false)

		updatedUser, err := s.FindByUUID(context.Background(), users[0].UUID)
		assert.NoError(t, err, "FindByUUID() error = %v, wantErr %v", err, false)

		assert.Equal(t, newUser.Login, updatedUser.Login, "received user must have same new Login value")
		assert.NotEqual(t, users[0].Login, updatedUser.Login, "received user must have new Login value")
	})
	t.Run("returns error, if updated user not found ", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}
		err := s.Update(context.Background(), &models.User{UUID: uuid.New().String()})
		assert.Error(t, err, "Update() error = %v, wantErr %v", err, true)
	})

	// Delete tests
	t.Run("not returns error, when trying to delete non existent user", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}
		err := s.Delete(context.Background(), &models.User{UUID: uuid.New().String()})
		assert.NoError(t, err, "Delete() error = %v, wantErr %v", err, false)
	})

	t.Run("deletes user", func(t *testing.T) {
		s := &UserDBStorage{
			db: sharedDB,
		}

		user, err := s.FindByUUID(context.Background(), users[0].UUID)
		require.NoError(t, err, "record must exists")

		err = s.Delete(context.Background(), user)
		require.NoError(t, err, "Delete() error = %v, wantErr %v", err, false)

		_, err = s.FindByUUID(context.Background(), users[0].UUID)
		require.Error(t, err, "user must be deleted by previous request")
	})
}

func Test_generateToken(t *testing.T) {
	t1 := generateToken()
	t2 := generateToken()
	assert.NotEqual(t, t1, t2)
	fmt.Println(t1, t2)
}
