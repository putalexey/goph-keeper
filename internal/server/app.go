package server

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/putalexey/goph-keeper/internal/server/config"
	"github.com/putalexey/goph-keeper/internal/server/interfaces"
	"github.com/putalexey/goph-keeper/internal/server/storage"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"sync"
	"time"
)

func Run(ctx context.Context, logger *zap.SugaredLogger, cfg *config.ServerConfig) error {
	rand.Seed(time.Now().UnixMicro())
	db, err := storage.NewDBConnection(cfg.DatabaseDSN)
	if err != nil {
		return errors.Wrap(err, "cannot connect to DB")
	}
	server := &Server{
		db: db,
	}
	initStorages(server)

	grpcServer := interfaces.NewGopherGRPCServer(ctx, logger, server.StoragesContainer, cfg.Address)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Serve()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("GRPC server stopped.")
	}()

	wg.Wait()
	log.Println("Shutting down server...")

	return nil
}

type Server struct {
	db                *sql.DB
	StoragesContainer *storage.StoragesContainer
}

// initStorage initializes storagers container
func initStorages(server *Server) {
	server.StoragesContainer = &storage.StoragesContainer{
		UserStorage:   storage.NewUserDBStorage(server.db),
		AuthStorage:   storage.NewAuthDBStorage(server.db),
		RecordStorage: storage.NewRecordDBStorage(server.db),
		EventStorage:  storage.NewEventStorager(server.db),
	}
}
