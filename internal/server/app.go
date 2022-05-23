package server

import (
	"context"
	"database/sql"
	"github.com/putalexey/goph-keeper/internal/server/config"
	"github.com/putalexey/goph-keeper/internal/server/interfaces"
	"github.com/putalexey/goph-keeper/internal/server/storage"
	"log"
	"sync"
)

func Run(ctx context.Context, cfg *config.ServerConfig) error {
	db, err := storage.NewDBConnection(cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	server := &Server{
		db: db,
	}
	initStorages(server)

	grpcServer := interfaces.NewGopherGRPCServer(ctx, server.StoragesContainer, cfg.Address)

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
