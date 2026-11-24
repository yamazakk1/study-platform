package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"study-platform/internal/handler"
	"study-platform/internal/repository"
	"study-platform/internal/service"

	"github.com/lib/pq"
)

func main() {
	self := flag.Bool("self", false, "use in-memory storage")
	flag.Parse()

	if err := run(*self); err != nil {
		log.Fatal(err)
	}
}

func run(self bool) error {
	var store service.Store
	switch dsn := os.Getenv("DATABASE_URL"); {
	case self:
		log.Print("self mode: enabled by --self")
	case dsn == "":
		log.Print("self mode: DATABASE_URL is not set")
	default:
		db, err := openDatabase(dsn)
		if err != nil {
			log.Printf("self mode: database unavailable: %v", err)
		} else {
			defer db.Close()
			store = repository.New(db)
		}
	}
	if store == nil {
		store = repository.NewMemory()
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	h, err := handler.New(service.New(store))
	if err != nil {
		return err
	}
	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, h.Routes())
}

func openDatabase(dsn string) (*sql.DB, error) {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		if err != nil {
			return nil, errors.New("invalid DATABASE_URL")
		}
	}
	connector, err := pq.NewConnector(dsn + " connect_timeout=2")
	if err != nil {
		return nil, errors.New("invalid DATABASE_URL")
	}

	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
