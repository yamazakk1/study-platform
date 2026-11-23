package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"study-platform/internal/handler"
	"study-platform/internal/repository"
	"study-platform/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	for i := 0; i < 20; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatal("database unavailable: ", err)
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	h, err := handler.New(service.New(repository.New(db)))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, h.Routes()))
}
