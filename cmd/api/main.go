package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"seasons/backend/gen/dbstore"
	appdb "seasons/backend/internal/db"
	"seasons/backend/internal/cleaner"
	"seasons/backend/internal/server"
	"seasons/backend/internal/server/handlers"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading environment directly")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := appdb.Connect(ctx, appdb.DefaultConfig(dbURL))
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()

	queries := dbstore.New(db)

	cleaner.StartSessionCleaner(ctx, queries, 1*time.Hour)

	srv := server.New(server.Dependencies{
		Queries: queries,
		Health:  handlers.NewHealthHandler(db),
		Auth:    handlers.NewAuthHandler(queries),
	})

	if err := srv.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
