package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	logger *log.Logger
	config config
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8008, "http server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "dsn", "postgres://scoretable:Noah2002ndw@localhost/scoretable",
		"DB connection string")

	app := &application{
		logger: log.New(os.Stdout, "", 0),
		config: cfg,
	}

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(app.config.port),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	app.logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
